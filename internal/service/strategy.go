package service

import (
	"container/list"
	"context"
	"fmt"
	"time"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/entity"
)

type RrStrategy struct {
	operators *list.List
	voip      VoipAdapter
	repo      OperatorRepo
	cfg       *config.DialerConfig
	logger    Logger
}

func (s *RrStrategy) updateOperators(operators []*entity.Operator) {
	ops := make(map[string]int)
	orig := make(map[string]int)
	remove := make([]*list.Element, 0)
	for i, op := range operators {
		ops[op.Number] = i
	}

	for e := s.operators.Front(); e != nil; e = e.Next() {
		num := e.Value.(*entity.Operator).Number
		if _, exists := ops[num]; !exists {
			remove = append(remove, e)
			continue
		}
		orig[num] = 1
	}

	for i := range remove {
		s.operators.Remove(remove[i])
	}

	for i := range operators {
		if _, exists := orig[operators[i].Number]; !exists {
			s.operators.PushBack(operators[i])
		}
	}
}

func (s *RrStrategy) Dial(ctx context.Context, client *entity.Client, operators []*entity.Operator) error {
	s.updateOperators(operators)
	s.operators.PushBack(nil)
	listNil := s.operators.Back()
	defer s.operators.Remove(listNil)

	ctx2, cancel := context.WithTimeout(ctx, s.cfg.DialToAllOperatorsTimeout)
	defer cancel()

	chErr := make(chan error)

	go func() {
		defer close(chErr)

		select {
		case <-ctx2.Done():
			chErr <- ctx2.Err()
		case <-client.Lost():
			chErr <- fmt.Errorf("client lost")
		}
	}()

	for {
		e := s.operators.Front()
		if e.Value == nil {
			s.operators.MoveToBack(e)
			select {
			case err := <-chErr:
				return err
			case <-time.After(s.cfg.DialPause):
				continue
			}
		}
		op, ok := e.Value.(*entity.Operator)
		if !ok {
			s.logger.Info("msg", "Cannot convert list.Entry to Operator", "client", client.Number)
			continue
		}
		s.operators.MoveToBack(e)
		if op.IsBusy() {
			s.logger.Info("msg", "Operator busy", "client", client.Number, "operator", op.Number)
			continue
		}

		select {
		case err := <-chErr:
			return err
		default:
		}

		if err := s.voip.Dial(ctx2, client, op); err != nil {
			s.logger.Info("msg", "Dial operator failed", "client", client.Number, "operator", op.Number, "error", err)
			continue
		}

		s.logger.Info("msg", "Dial success", "client", client.Number, "operator", op.Number)

		if err := s.repo.SetBusy(op.Number, true); err != nil {
			s.logger.Info("msg", "Cannot set operator busy", "client", client.Number, "operator", op.Number, "error", err)
		}
		return nil
	}
}

func NewRrStrategy(cfg *config.DialerConfig, voip VoipAdapter, repo OperatorRepo, logger Logger) Strategy {
	return &RrStrategy{
		operators: list.New(),
		voip:      voip,
		repo:      repo,
		cfg:       cfg,
		logger:    logger,
	}
}
