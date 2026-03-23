package watcher

import (
	"context"
	"errors"
	"log"

	"github.com/fsnotify/fsnotify"

	"downloads-organizer/internal/organizer"
)

type Watcher struct {
	downloadsDir string
	organizer    *organizer.Organizer
	logger       *log.Logger
}

func New(downloadsDir string, org *organizer.Organizer, logger *log.Logger) *Watcher {
	return &Watcher{
		downloadsDir: downloadsDir,
		organizer:    org,
		logger:       logger,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	if err := w.organizer.ScanAndOrganize(); err != nil {
		w.logger.Printf("initial scan failed: %v", err)
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer fsWatcher.Close()

	if err := fsWatcher.Add(w.downloadsDir); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-fsWatcher.Events:
			if !ok {
				return nil
			}

			if shouldHandleEvent(event) {
				go w.organizer.TryOrganize(event.Name)
			}
		case err, ok := <-fsWatcher.Errors:
			if !ok {
				return nil
			}
			if errors.Is(err, context.Canceled) {
				return nil
			}
			w.logger.Printf("watcher error: %v", err)
		}
	}
}

func shouldHandleEvent(event fsnotify.Event) bool {
	return event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) != 0
}
