package service

import (
	"context"

	"github.com/jgivc/vapp/internal/entity"
)

type (
	// Adapters.

	Queue interface {
		IsFull() bool
		HasClients() bool
		Push(client *entity.Client) error
		Pop() (*entity.Client, error)
	}

	VoipAdapter interface {
		//TODO: Must cancel on context
		Answer(ctx context.Context, client *entity.Client) error
		Playback(ctx context.Context, client *entity.Client, fileName string) error
		StartMOH(ctx context.Context, client *entity.Client) error
		StopMOH(ctx context.Context, client *entity.Client) error
	}

	ClientRepo interface {
		New(number string, data interface{}) (*entity.Client, error)
		Remove(number string, data interface{}) error
	}

	OperatorRepo interface{}

	Logger interface {
		Info(args ...interface{})
		Warn(args ...interface{})
		Error(args ...interface{})
	}

	Dialer interface {
		Notify()
	}
)