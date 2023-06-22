package service

import "github.com/jgivc/vapp/internal/entity"

type QueueService struct {
}

func (s *QueueService) IsFull() bool {
	return false
}

func (s *QueueService) Push(client *entity.Client) error {
	return nil
}

func (s *QueueService) Pop() (*entity.Client, error) {
	return nil, nil
}

func (s *QueueService) Close() error {
	return nil
}
