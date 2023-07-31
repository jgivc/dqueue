package service

import (
	"context"
	"time"

	"github.com/jgivc/vapp/config"
)

type DialerService struct {
	queue    Queue
	repo     OperatorRepo
	logger   Logger
	notify   chan struct{}
	cfg      *config.DialerConfig
	strategy Strategy
}

func (s *DialerService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.CheckInterval)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			select {
			case s.notify <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		defer close(s.notify)

		<-ctx.Done()
	}()

	go s.handle(ctx)
}

func (s *DialerService) handle(ctx context.Context) {
	for range s.notify {
		for {
			if !s.queue.HasClients() {
				break
			}

			ops, err := s.repo.GetOperators(ctx)
			if err != nil || len(ops) < 1 {
				s.logger.Error("msg", "Cannot get operators", "error", err)
				break
			}

			client, err := s.queue.Pop()
			if err != nil {
				s.logger.Error("msg", "Cannot get client from queue", "error", err)
				break
			}

			if err2 := s.strategy.Dial(ctx, client, ops); err2 != nil {
				s.logger.Error("msg", "Cannot handle dial to operators", "error", err2)
			}
		}
	}
}

func (s *DialerService) Notify() {
	select {
	case s.notify <- struct{}{}:
	default:
	}
}

func NewDialerService(cfg *config.DialerConfig, queue Queue, repo OperatorRepo,
	strategy Strategy, logger Logger) *DialerService {
	return &DialerService{
		queue:    queue,
		repo:     repo,
		logger:   logger,
		notify:   make(chan struct{}, 1),
		cfg:      cfg,
		strategy: strategy,
	}
}
