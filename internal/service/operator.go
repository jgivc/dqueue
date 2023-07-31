package service

import (
	"context"
	"fmt"

	"github.com/jgivc/vapp/internal/entity"
)

type OperatorService struct {
	repo   OperatorRepo
	dialer Dialer
	logger Logger
}

func (s *OperatorService) GetOperators(ctx context.Context) ([]*entity.Operator, error) {
	return s.repo.GetOperators(ctx)
}

func (s *OperatorService) SetBusy(number string, busy bool) error {
	err := s.repo.SetBusy(number, busy)
	if err != nil {
		return err
	}

	if !busy && s.repo.Exists(number) {
		s.dialer.Notify()
		s.logger.Info("msg", "Notify dialer, operator free", "number", number)
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ ### ~~~~~~~~~~~~~~~~ ### ~~~~~~~~~~~~~~~~ ### Free:", number)
	}

	return nil
}

func NewOperatorService(repo OperatorRepo, dialer Dialer, logger Logger) *OperatorService {
	return &OperatorService{
		repo:   repo,
		dialer: dialer,
		logger: logger,
	}
}
