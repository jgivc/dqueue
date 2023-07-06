package ami

import (
	"context"
	"io"
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

	amiReaderIf interface {
		Read() (Event, error)
		Close() error
	}

	connectionFactory interface {
		Connect(ctx context.Context, addr string) (net.Conn, error)
	}

	readerFactory interface {
		GetAmiReader(r io.Reader) amiReaderIf
	}
)
