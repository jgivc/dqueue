package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/entity"
)

const (
	notifyChannelBufferLength = 2
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

func hasFreeOperators(ops []*entity.Operator) bool {
	for i := range ops {
		if !ops[i].IsBusy() {
			return true
		}
	}

	return false
}

func (s *DialerService) handle(ctx context.Context) {
	i := 1
	for range s.notify {
		s.logger.Info("msg", "Recv notify")
		for {
			if !s.queue.HasClients() {
				break
			}

			ops, err := s.repo.GetOperators(ctx)
			if err != nil || len(ops) < 1 {
				s.logger.Error("msg", "Cannot get operators", "error", err)
				break
			}

			if !hasFreeOperators(ops) {
				s.logger.Info("msg", "No free operators")
				break
			}

			client, err := s.queue.Pop()
			if err != nil {
				s.logger.Error("msg", "Cannot get client from queue", "error", err)
				break
			}
			s.logger.Info("msg", "Get client from queue", "id", client.ID)

			fmt.Println("####################### Dial to", client.ID, "--------------------------->>>>>>>", i)
			if err2 := s.strategy.Dial(ctx, client, ops); err2 != nil {
				s.logger.Error("msg", "Cannot handle dial to operators", "error", err2)
			}

			fmt.Println("####################### Dial done", client.ID, "---------------------------<<<<<<<", i)
			// s.logger.Error("msg", "Dial done")
			i++
		}
	}
}

func (s *DialerService) Notify() {
	select {
	case s.notify <- struct{}{}:
	default:
		s.logger.Info("msg", "Cannot send notify to channel")
	}
}

func NewDialerService(cfg *config.DialerConfig, queue Queue, repo OperatorRepo,
	strategy Strategy, logger Logger) *DialerService {
	return &DialerService{
		queue:    queue,
		repo:     repo,
		logger:   logger,
		notify:   make(chan struct{}, notifyChannelBufferLength),
		cfg:      cfg,
		strategy: strategy,
	}
}
