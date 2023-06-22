package service

import (
	"container/list"
	"errors"
	"sync"

	"github.com/jgivc/vapp/internal/entity"
)

var (
	errQueueFull  = errors.New("queue is full")
	errQueueEmpty = errors.New("queue is empty")
	errNoElement  = errors.New("cannot get element")
)

type QueueService struct {
	mux       sync.Mutex
	maxLength int
	clients   *list.List
}

func (s *QueueService) IsFull() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return !(s.clients.Len() < s.maxLength)
}

func (s *QueueService) Push(client *entity.Client) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !(s.clients.Len() < s.maxLength) {
		return errQueueFull
	}

	s.clients.PushBack(client)

	return nil
}

func (s *QueueService) Pop() (*entity.Client, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.clients.Len() < 1 {
		return nil, errQueueEmpty
	}

	el := s.clients.Remove(s.clients.Front())
	if el == nil {
		return nil, errNoElement
	}

	return el.(*entity.Client), nil

}

func (s *QueueService) Close() error {
	return nil
}
