package app

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"

	"downloads-organizer/internal/organizer"
	"downloads-organizer/internal/watcher"
)

type Service struct {
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	organizer *organizer.Organizer
	watcher   *watcher.Watcher
	logger    *log.Logger
}

func NewService(cfg organizer.Config, logger *log.Logger) *Service {
	if logger == nil {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}

	org := organizer.New(cfg, logger)

	return &Service{
		organizer: org,
		watcher:   watcher.New(cfg.DownloadsDir, org, logger),
		logger:    logger,
	}
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.watcher.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Printf("watcher stopped with error: %v", err)
		}
	}()

	return nil
}

func (s *Service) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	cancel := s.cancel
	s.running = false
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	s.wg.Wait()
}

func (s *Service) ScanNow() error {
	return s.organizer.ScanAndOrganize()
}

func (s *Service) SetMoveHandler(handler func(organizer.MoveEvent)) {
	s.organizer.SetMoveHandler(handler)
}

func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.running
}
