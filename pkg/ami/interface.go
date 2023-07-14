package ami

import (
	"context"
	"net"
)

type (
	Ami interface {
		Start(ctx context.Context) error
		Subscribe(filter Filter) Subscriber
		Close() error

		Answer(ctx context.Context, host string, channel string) error
		Playback(ctx context.Context, host string, channel string, fileName string) error
		StartMOH(ctx context.Context, host string, channel string) error
		StopMOH(ctx context.Context, host string, channel string) error
		Hangup(ctx context.Context, host string, channel string, cause int) error
	}

	Filter func(e *Event) bool

	Subscriber interface {
		Events() <-chan *Event
		Close()
	}

	pubSubIf interface {
		Subscribe(f Filter) Subscriber
		Publish(e *Event)
		Close()
	}

	connectionFactory interface {
		Connect(ctx context.Context, addr string) (net.Conn, error)
	}
)
