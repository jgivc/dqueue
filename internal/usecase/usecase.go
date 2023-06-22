package usecase

import (
	"context"

	"github.com/jgivc/vapp/internal/entity"
)

/*
	Usecase are the features that the app provides.
*/

type (
	clientRepo interface {
		New(host, uniqueID, channel, number string) (*entity.Client, error)
		Remove(host, uniqueID, channel string) error
	}

	queueService interface {
		IsFull() bool
		Push(client *entity.Client) error
		Pop() (*entity.Client, error)
	}

	dialerService interface {
		Notify()
	}

	voipService interface {
		Answer(ctx context.Context, client *entity.Client) error
		Playback(ctx context.Context, client *entity.Client) error
		StartMOH(ctx context.Context, client *entity.Client) error
		StopMOH(ctx context.Context, client *entity.Client) error
	}
)
