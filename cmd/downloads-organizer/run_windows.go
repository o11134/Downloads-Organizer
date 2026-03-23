//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"

	"downloads-organizer/internal/app"
	"downloads-organizer/internal/notify"
	"downloads-organizer/internal/organizer"
	"downloads-organizer/internal/settings"
	"downloads-organizer/internal/startup"
)

func run(service *app.Service, cfg settings.RuntimeConfig, logger *log.Logger, logPath string, store *settings.Store) error {
	ctx := newWindowsRunContext(service, cfg, logger, store)

	onReady := func() {
		systray.SetTitle("Downloads Organizer")
		systray.SetTooltip("Organize downloads automatically")

		if err := ctx.applyStartupSetting(); err != nil {
			logger.Printf("failed to apply startup setting: %v", err)
		}

		if err := ctx.service.Start(); err != nil {
			logger.Printf("failed to start organizer service: %v", err)
		}

		mToggle := systray.AddMenuItem("Pause organizing", "Pause auto organizing")
		mScan := systray.AddMenuItem("Organize now", "Run a manual scan now")
		mOpenDownloads := systray.AddMenuItem("Open Downloads", "Open the Downloads folder")
		mOpenLog := systray.AddMenuItem("Open Log", "Open organizer log file")
		mOpenConfig := systray.AddMenuItem("Open Config", "Open config file")
		mNotifications := systray.AddMenuItemCheckbox("Disable notifications", "Toggle Windows notifications", true)
		mStartup := systray.AddMenuItem("Enable startup", "Run on Windows startup")
		systray.AddSeparator()
		mExit := systray.AddMenuItem("Exit", "Exit Downloads Organizer")

		ctx.updateNotificationMenu(mNotifications)
		updateToggleMenu(mToggle, ctx.service.IsRunning())
		if err := updateStartupMenu(mStartup); err != nil {
			logger.Printf("failed to check startup state: %v", err)
		}

		go func() {
			for {
				select {
				case <-mToggle.ClickedCh:
					if ctx.service.IsRunning() {
						ctx.service.Stop()
						logger.Printf("organizer paused")
					} else {
						if err := ctx.service.Start(); err != nil {
							logger.Printf("failed to resume organizer: %v", err)
						} else {
							logger.Printf("organizer resumed")
						}
					}
					updateToggleMenu(mToggle, ctx.service.IsRunning())
				case <-mScan.ClickedCh:
					go func() {
						if err := ctx.service.ScanNow(); err != nil {
							logger.Printf("manual scan failed: %v", err)
							return
						}
						logger.Printf("manual scan complete")
					}()
				case <-mOpenDownloads.ClickedCh:
					if err := openInExplorer(ctx.cfg.Organizer.DownloadsDir); err != nil {
						logger.Printf("failed to open Downloads: %v", err)
					}
				case <-mOpenLog.ClickedCh:
					if err := openInExplorer(logPath); err != nil {
						logger.Printf("failed to open log: %v", err)
					}
				case <-mOpenConfig.ClickedCh:
					if err := openInExplorer(ctx.store.Path()); err != nil {
						logger.Printf("failed to open config: %v", err)
					}
				case <-mNotifications.ClickedCh:
					next := !ctx.notificationsEnabled()
					if _, err := ctx.store.SetNotificationsEnabled(next); err != nil {
						logger.Printf("failed to persist notifications setting: %v", err)
						continue
					}
					ctx.setNotificationsEnabled(next)
					ctx.updateNotificationMenu(mNotifications)
					logger.Printf("notifications enabled=%t", next)
				case <-mStartup.ClickedCh:
					enabled, err := startup.IsEnabled()
					if err != nil {
						logger.Printf("failed to check startup state: %v", err)
						continue
					}

					next := !enabled
					if next {
						exePath, err := os.Executable()
						if err != nil {
							logger.Printf("failed to get executable path: %v", err)
							continue
						}
						if err := startup.Enable(exePath); err != nil {
							logger.Printf("failed to enable startup: %v", err)
							continue
						}
					} else {
						if err := startup.Disable(); err != nil {
							logger.Printf("failed to disable startup: %v", err)
							continue
						}
					}

					if _, err := ctx.store.SetStartupEnabled(next); err != nil {
						logger.Printf("failed to persist startup setting: %v", err)
					}

					if err := updateStartupMenu(mStartup); err != nil {
						logger.Printf("failed to update startup menu: %v", err)
					}
					logger.Printf("startup enabled=%t", next)
				case <-mExit.ClickedCh:
					ctx.stopNotificationWorker()
					systray.Quit()
					return
				}
			}
		}()
	}

	onExit := func() {
		ctx.service.SetMoveHandler(nil)
		ctx.stopNotificationWorker()
		ctx.service.Stop()
	}

	systray.Run(onReady, onExit)
	return nil
}

type windowsRunContext struct {
	service *app.Service
	cfg     settings.RuntimeConfig
	logger  *log.Logger
	store   *settings.Store

	notifier      *notify.Notifier
	moveEvents    chan organizer.MoveEvent
	stopCh        chan struct{}
	stopOnce      sync.Once
	notifications struct {
		mu      sync.RWMutex
		enabled bool
	}
	notificationSettings settings.NotificationSettings
}

func newWindowsRunContext(service *app.Service, cfg settings.RuntimeConfig, logger *log.Logger, store *settings.Store) *windowsRunContext {
	ctx := &windowsRunContext{
		service:              service,
		cfg:                  cfg,
		logger:               logger,
		store:                store,
		notifier:             notify.New("DownloadsOrganizer"),
		moveEvents:           make(chan organizer.MoveEvent, 256),
		stopCh:               make(chan struct{}),
		notificationSettings: cfg.Notifications,
	}

	ctx.notifications.enabled = cfg.Notifications.Enabled
	ctx.applyMoveHandler()
	go runNotificationWorker(ctx.moveEvents, ctx.stopCh, ctx.notifier, logger, cfg.Notifications)

	return ctx
}

func (c *windowsRunContext) notificationsEnabled() bool {
	c.notifications.mu.RLock()
	defer c.notifications.mu.RUnlock()
	return c.notifications.enabled
}

func (c *windowsRunContext) setNotificationsEnabled(enabled bool) {
	c.notifications.mu.Lock()
	c.notifications.enabled = enabled
	c.notifications.mu.Unlock()
	c.applyMoveHandler()
}

func (c *windowsRunContext) applyMoveHandler() {
	if !c.notificationsEnabled() {
		c.service.SetMoveHandler(nil)
		return
	}

	c.service.SetMoveHandler(func(event organizer.MoveEvent) {
		select {
		case c.moveEvents <- event:
		default:
			c.logger.Printf("dropping move event because queue is full")
		}
	})
}

func (c *windowsRunContext) updateNotificationMenu(item *systray.MenuItem) {
	if c.notificationsEnabled() {
		item.SetTitle("Disable notifications")
		item.Check()
		return
	}

	item.SetTitle("Enable notifications")
	item.Uncheck()
}

func (c *windowsRunContext) stopNotificationWorker() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

func (c *windowsRunContext) applyStartupSetting() error {
	want := c.cfg.StartupEnabled
	has, err := startup.IsEnabled()
	if err != nil {
		return err
	}
	if want == has {
		return nil
	}

	if want {
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		return startup.Enable(exePath)
	}

	return startup.Disable()
}

func updateToggleMenu(item *systray.MenuItem, running bool) {
	if running {
		item.SetTitle("Pause organizing")
		return
	}

	item.SetTitle("Resume organizing")
}

func updateStartupMenu(item *systray.MenuItem) error {
	enabled, err := startup.IsEnabled()
	if err != nil {
		item.SetTitle("Enable startup")
		return err
	}

	if enabled {
		item.SetTitle("Disable startup")
		return nil
	}

	item.SetTitle("Enable startup")
	return nil
}

func runNotificationWorker(
	events <-chan organizer.MoveEvent,
	stop <-chan struct{},
	notifier *notify.Notifier,
	logger *log.Logger,
	settings settings.NotificationSettings,
) {
	ticker := time.NewTicker(settings.BatchInterval)
	defer ticker.Stop()

	counts := map[string]int{}
	total := 0

	flush := func() {
		if total == 0 {
			return
		}

		message := buildNotificationMessage(total, counts)
		if err := notifier.Notify("Downloads Organizer", message); err != nil {
			logger.Printf("failed to send notification: %v", err)
		}

		counts = map[string]int{}
		total = 0
	}

	for {
		select {
		case <-stop:
			flush()
			return
		case event := <-events:
			counts[event.Category]++
			total++
			if total >= settings.BatchMaxFiles {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func buildNotificationMessage(total int, counts map[string]int) string {
	parts := make([]string, 0, 4)
	ordered := []string{"Images", "Documents", "Programs", "Archives"}

	for _, category := range ordered {
		count := counts[category]
		if count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %d", category, count))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Moved %d files.", total)
	}

	return fmt.Sprintf("Moved %d files (%s)", total, strings.Join(parts, ", "))
}

func openInExplorer(path string) error {
	cmd := exec.Command("explorer", path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open explorer: %w", err)
	}
	return nil
}
