package organizer

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Organizer struct {
	cfg      Config
	logger   *log.Logger
	mu       sync.Mutex
	inflight map[string]struct{}

	handlerMu   sync.RWMutex
	moveHandler func(MoveEvent)
}

type MoveEvent struct {
	Source      string
	Destination string
	Category    string
	MovedAt     time.Time
}

func New(cfg Config, logger *log.Logger) *Organizer {
	if logger == nil {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}

	return &Organizer{
		cfg:      cfg,
		logger:   logger,
		inflight: make(map[string]struct{}),
	}
}

func (o *Organizer) ScanAndOrganize() error {
	entries, err := os.ReadDir(o.cfg.DownloadsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(o.cfg.DownloadsDir, entry.Name())
		o.TryOrganize(filePath)
	}

	return nil
}

func (o *Organizer) TryOrganize(filePath string) {
	cleanPath, err := filepath.Abs(filePath)
	if err != nil {
		o.logger.Printf("skip file %q: cannot resolve path: %v", filePath, err)
		return
	}

	if !o.markInFlight(cleanPath) {
		return
	}
	defer o.unmarkInFlight(cleanPath)

	if err := o.organizeOne(cleanPath); err != nil {
		o.logger.Printf("skip file %q: %v", cleanPath, err)
	}
}

func (o *Organizer) SetMoveHandler(handler func(MoveEvent)) {
	o.handlerMu.Lock()
	defer o.handlerMu.Unlock()

	o.moveHandler = handler
}

func (o *Organizer) organizeOne(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if info.IsDir() {
		return nil
	}

	if filepath.Dir(filePath) != o.cfg.DownloadsDir {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(info.Name()))
	if _, ignored := o.cfg.IgnoredExtensions[ext]; ignored {
		return nil
	}

	category, ok := o.cfg.CategoryByExtension[ext]
	if !ok {
		return nil
	}

	stable, err := o.waitForStableFile(filePath)
	if err != nil {
		return err
	}
	if !stable {
		return fmt.Errorf("file is still changing")
	}

	targetDir := filepath.Join(o.cfg.DownloadsDir, category)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	targetPath, err := nextAvailablePath(targetDir, info.Name())
	if err != nil {
		return err
	}

	if err := os.Rename(filePath, targetPath); err != nil {
		return err
	}

	o.logger.Printf("moved %q -> %q", filePath, targetPath)
	o.notifyMove(MoveEvent{
		Source:      filePath,
		Destination: targetPath,
		Category:    category,
		MovedAt:     time.Now(),
	})
	return nil
}

func (o *Organizer) notifyMove(event MoveEvent) {
	o.handlerMu.RLock()
	handler := o.moveHandler
	o.handlerMu.RUnlock()

	if handler != nil {
		handler(event)
	}
}

func (o *Organizer) waitForStableFile(filePath string) (bool, error) {
	if o.cfg.StabilityChecks <= 0 {
		return true, nil
	}

	lastSize := int64(-1)
	for i := 0; i < o.cfg.StabilityChecks; i++ {
		info, err := os.Stat(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return false, nil
			}
			return false, err
		}

		size := info.Size()
		if size == lastSize && i > 0 {
			return true, nil
		}

		lastSize = size
		time.Sleep(o.cfg.StabilityDelay)
	}

	return false, nil
}

func nextAvailablePath(dir string, fileName string) (string, error) {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	ext := filepath.Ext(fileName)

	candidate := filepath.Join(dir, fileName)
	_, err := os.Stat(candidate)
	if err == nil {
		for i := 1; i < 10000; i++ {
			candidate = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
			if _, statErr := os.Stat(candidate); errors.Is(statErr, os.ErrNotExist) {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("could not find free name for %q", fileName)
	}

	if errors.Is(err, os.ErrNotExist) {
		return candidate, nil
	}

	return "", err
}

func (o *Organizer) markInFlight(path string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.inflight[path]; exists {
		return false
	}

	o.inflight[path] = struct{}{}
	return true
}

func (o *Organizer) unmarkInFlight(path string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.inflight, path)
}
