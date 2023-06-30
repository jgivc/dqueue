package service

import (
	"context"

	"github.com/jgivc/vapp/internal/entity"
)

type OperatorService struct {
	repo OperatorRepo
}

func (s *OperatorService) GetOperators(ctx context.Context) ([]*entity.Operator, error) {
	return s.repo.GetOperators(ctx)
}

func (s *OperatorService) SetBusy(number string, busy bool) error {
	return s.repo.SetBusy(number, busy)
}

func NewOperatorService(repo OperatorRepo) *OperatorService {
	return &OperatorService{
		repo: repo,
	}
}
