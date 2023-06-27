package mocks

import (
	"github.com/jgivc/vapp/internal/entity"
	"github.com/stretchr/testify/mock"
)

type QueueMock struct {
	mock.Mock
}

func (m *QueueMock) IsFull() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *QueueMock) HasClients() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *QueueMock) Push(client *entity.Client) error {
	args := m.Called(client)

	return args.Error(0)
}

func (m *QueueMock) Pop() (*entity.Client, error) {
	args := m.Called()

	var (
		client *entity.Client
		ok     bool
	)

	if args.Get(0) != nil {
		client, ok = args.Get(0).(*entity.Client)
		if !ok {
			panic("cannot convert to *entity.Client")
		}
	}

	return client, args.Error(1)
}
