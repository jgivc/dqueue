package ami

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/stretchr/testify/mock"
)

type amiReaderMock struct {
	mock.Mock
}

func (m *amiReaderMock) Read() (Event, error) {
	args := m.Called()

	var e Event
	if a := args.Get(0); a != nil {
		if e2, ok := a.(Event); ok {
			return e2, nil
		}
	}

	return e, args.Error(1)
}

func (m *amiReaderMock) Close() error {
	args := m.Called()

	return args.Error(0)
}

type connectionMock struct {
	mock.Mock
}

func (m *connectionMock) Read(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *connectionMock) Write(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *connectionMock) Close() error {
	args := m.Called()

	return args.Error(0)
}

func (m *connectionMock) LocalAddr() net.Addr {
	args := m.Called()

	var addr net.Addr

	if a := args.Get(0); a != nil {
		if addr2, ok := a.(net.Addr); ok {
			return addr2
		}
	}

	return addr
}

func (m *connectionMock) RemoteAddr() net.Addr {
	args := m.Called()

	var addr net.Addr

	if a := args.Get(0); a != nil {
		if addr2, ok := a.(net.Addr); ok {
			return addr2
		}
	}

	return addr
}

func (m *connectionMock) SetDeadline(t time.Time) error {
	args := m.Called(t)

	return args.Error(0)
}

func (m *connectionMock) SetReadDeadline(t time.Time) error {
	args := m.Called(t)

	return args.Error(0)
}

func (m *connectionMock) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)

	return args.Error(0)
}

type connectionFactoryMock struct {
	mock.Mock
}

func (m *connectionFactoryMock) Connect(ctx context.Context, addr string) (net.Conn, error) {
	args := m.Called(ctx, addr)

	var conn net.Conn

	if a := args.Get(0); a != nil {
		if c, ok := a.(net.Conn); ok {
			return c, args.Error(1)
		}
	}

	return conn, args.Error(1)
}

type readerFactoryMock struct {
	mock.Mock
}

func (m *readerFactoryMock) GetAmiReader(r io.Reader) amiReaderIf {
	args := m.Called(r)

	var reader amiReaderIf

	if a := args.Get(0); a != nil {
		if r2, ok := a.(amiReaderIf); ok {
			return r2
		}
	}

	return reader
}

type pubSubMock struct {
	mock.Mock
}

func (m *pubSubMock) Subscribe(f Filter) Subscriber {
	args := m.Called(f)

	var s Subscriber
	if a := args.Get(0); a != nil {
		if s2, ok := a.(Subscriber); ok {
			return s2
		}
	}

	return s
}

func (m *pubSubMock) Publish(e *Event) {
	m.Called(e)
}

func (m *pubSubMock) Close() {
	m.Called()
}

type subscriberMock struct {
	mock.Mock
}

func (m *subscriberMock) Events() <-chan *Event {
	args := m.Called()

	var ch <-chan *Event
	if a := args.Get(0); a != nil {
		if ch2, ok := a.(chan *Event); ok {
			return ch2
		}
	}

	return ch
}

func (m *subscriberMock) Close() {
	m.Called()
}
