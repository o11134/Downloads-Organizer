package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"downloads-organizer/internal/app"
	"downloads-organizer/internal/organizer"
	"downloads-organizer/internal/settings"
)

func main() {
	defaultDownloadsDir, err := organizer.DefaultDownloadsDir()
	if err != nil {
		log.Fatalf("cannot resolve Downloads directory: %v", err)
	}

	logger, appDir, logPath, closeLogger, err := newFileLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}
	defer closeLogger()

	store, runtimeCfg, err := settings.Open(appDir, defaultDownloadsDir)
	if err != nil {
		log.Fatalf("cannot open settings: %v", err)
	}
	logger.Printf("using config file: %s", store.Path())

	service := app.NewService(runtimeCfg.Organizer, logger)

	if err := run(service, runtimeCfg, logger, logPath, store); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func newFileLogger() (*log.Logger, string, string, func(), error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, "", "", nil, err
	}

	appDir := filepath.Join(configDir, "DownloadsOrganizer")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return nil, "", "", nil, err
	}

	logPath := filepath.Join(appDir, "organizer.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, "", "", nil, err
	}

	logger := log.New(file, "", log.LstdFlags)
	logger.Printf("starting Downloads Organizer")

	closer := func() {
		if closeErr := file.Close(); closeErr != nil && !errors.Is(closeErr, os.ErrClosed) {
			fmt.Fprintf(os.Stderr, "failed to close log file: %v\n", closeErr)
		}
	}

	return logger, appDir, logPath, closer, nil
}
