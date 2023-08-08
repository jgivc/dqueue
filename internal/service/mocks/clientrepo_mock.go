package mocks

import (
	"github.com/jgivc/dqueue/internal/entity"
	"github.com/stretchr/testify/mock"
)

type ClientRepoMock struct {
	mock.Mock
}

func (m *ClientRepoMock) New(number string, data interface{}) (*entity.Client, error) {
	args := m.Called(number, data)

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

func (m *ClientRepoMock) Remove(number string, data interface{}) error {
	args := m.Called(number, data)

	return args.Error(0)
}
