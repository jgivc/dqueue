package mocks

import (
	"context"

	"github.com/jgivc/vapp/internal/entity"
	"github.com/stretchr/testify/mock"
)

type VoipMock struct {
	mock.Mock
}

func (m *VoipMock) Answer(ctx context.Context, client *entity.Client) error {
	args := m.Called(ctx, client)

	return args.Error(0)
}

func (m *VoipMock) Playback(ctx context.Context, client *entity.Client, fileName string) error {
	args := m.Called(ctx, client, fileName)

	return args.Error(0)
}

func (m *VoipMock) StartMOH(ctx context.Context, client *entity.Client) error {
	args := m.Called(ctx, client)

	return args.Error(0)
}

func (m *VoipMock) StopMOH(ctx context.Context, client *entity.Client) error {
	args := m.Called(ctx, client)

	return args.Error(0)
}

func (m *VoipMock) Dial(ctx context.Context, client *entity.Client, operator *entity.Operator) error {
	args := m.Called(ctx, client)

	return args.Error(0)
}

func (m *VoipMock) Hangup(ctx context.Context, client *entity.Client) error {
	args := m.Called(ctx, client)

	return args.Error(0)
}
