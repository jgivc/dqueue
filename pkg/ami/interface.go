package ami

import (
	"context"
	"net"
)

type (
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
