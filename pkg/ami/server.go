package ami

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/pkg/logger"
)

/*
Action: Login
ActionID: <value>
Username: <value>
Secret: <value>

Action: Logoff
ActionID: <value>
*/

const (
	network   = "tcp"
	reqLogin  = "Action: Login\r\nActionID: %d\r\nUsername: %s\r\nSecret: %s\r\n\r\n"
	reqLogoff = "Action: Logoff\r\nActionID: %d\r\n\r\n"

	stateReady = iota
	stateDisconnect
	stateShutdown
)

var (
	errAmiServer = errors.New("ami server error")
)

type serverState int

func (s serverState) String() string {
	switch s {
	case stateReady:
		return "Ready"
	case stateDisconnect:
		return "Disconnect"
	case stateShutdown:
		return "Shutdown"
	}

	return "Unknown"
}

func (s serverState) IsReady() bool {
	return s == stateReady
}

type amiServer interface {
	Start(ctx context.Context) error
	io.Writer
	io.Closer
}

type amiServerImpl struct {
	mux      sync.Mutex // Protect writer against simultaneous use
	addr     string
	cfg      *config.AmiServer
	conn     net.Conn
	logger   logger.Logger
	state    serverState
	ps       pubSub
	shutdown atomic.Bool
}

func (s *amiServerImpl) connect(ctx context.Context) (net.Conn, error) {
	ctx2, cancel := context.WithTimeout(ctx, s.cfg.DialTimeout)
	defer cancel()

	var dialer net.Dialer
	return dialer.DialContext(ctx2, network, s.addr)
}

func (s *amiServerImpl) login(conn net.Conn) (*amiReader, error) {
	panic("not implemented")
}

func (s *amiServerImpl) Start(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.state == stateShutdown {
		return fmt.Errorf("%w shutdown flag", errAmiServer)
	}

	s.state = stateDisconnect

	go func() {
		<-ctx.Done()
		s.shutdown.Store(true)
	}()

	go func() {
		for {
			var (
				err    error
				conn   net.Conn
				reader *amiReader
			)

			if conn, err = s.connect(ctx); err == nil {
				reader, err = s.login(conn)
			}
			if err != nil {
				s.logger.Error("msg", "Cannot connect to server", "addr", s.addr, "error", err)

				select {
				case <-ctx.Done():
					s.logger.Info("msg", "Interrupt server reader", "addr", s.addr)
					return
				case <-time.After(s.cfg.ReconnectInterval):
				}

				continue
			}

			for !s.shutdown.Load() {
				e, err2 := reader.Read()
				if err2 != nil {
					s.logger.Error("msg", "Cannot read on server", "addr", s.addr, "error", err2)
					s.setState(stateDisconnect)
					conn.Close()
					break
				}

				s.ps.Publish(&e)
			}

			if s.shutdown.Load() {
				return
			}
		}
	}()

	return nil
}

func (s *amiServerImpl) setState(state serverState) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.state = state
}

func (s *amiServerImpl) Write(p []byte) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.state.IsReady() && s.conn != nil {
		return s.conn.Write(p)
	}

	return 0, fmt.Errorf("%w not connected", errAmiServer)
}

func (s *amiServerImpl) Close() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.state == stateShutdown {
		return fmt.Errorf("%w already shutdown", errAmiServer)
	}

	s.state = stateShutdown
	s.shutdown.Store(true)
	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}

func newAmiServer(cfg *config.AmiServer, ps pubSub, logger logger.Logger) amiServer {
	return &amiServerImpl{
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		cfg:    cfg,
		logger: logger,
		state:  stateDisconnect,
		ps:     ps,
	}
}
