//go:build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"downloads-organizer/internal/app"
	"downloads-organizer/internal/settings"
)

func run(service *app.Service, cfg settings.RuntimeConfig, logger *log.Logger, logPath string, store *settings.Store) error {
	if err := service.Start(); err != nil {
		return err
	}

	logger.Printf("running in console mode on non-Windows")
	fmt.Printf("Downloads Organizer is watching %s\n", cfg.Organizer.DownloadsDir)
	fmt.Printf("log file: %s\n", logPath)
	fmt.Printf("config file: %s\n", store.Path())
	fmt.Println("press Ctrl+C to exit")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	service.Stop()
	return nil
}
