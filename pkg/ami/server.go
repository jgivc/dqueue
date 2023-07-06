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

	"github.com/google/uuid"
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
	reqLogin  = "Action: Login\r\nActionID: %s\r\nUsername: %s\r\nSecret: %s\r\n\r\n"
	reqLogoff = "Action: Logoff\r\nActionID: %s\r\n\r\n"

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
	cf       connectionFactory
	rf       readerFactory
	ps       pubSubIf
	shutdown atomic.Bool
}

func (s *amiServerImpl) connect(ctx context.Context) (net.Conn, error) {
	ctx2, cancel := context.WithTimeout(ctx, s.cfg.DialTimeout)
	defer cancel()

	// var dialer net.Dialer
	// return dialer.DialContext(ctx2, network, s.addr)
	return s.cf.Connect(ctx2, s.addr)
}

func (s *amiServerImpl) getID() string {
	return uuid.New().String()
}

func (s *amiServerImpl) login(ctx context.Context, conn net.Conn) (amiReaderIf, error) {
	reader := s.rf.GetAmiReader(conn)
	ctx2, cancel := context.WithTimeout(ctx, s.cfg.ActionTimeout)
	defer cancel()

	ch := make(chan *Event)
	go func() {
		defer close(ch)

		for {
			e, err := reader.Read()
			if err != nil {
				s.logger.Error("msg", "Cannot read event", "addr", s.addr, "error", err)
				return
			}
			select {
			case <-ctx2.Done():
				return
			case ch <- &e:
			}
		}
	}()

	id := s.getID()

	var err error
	if _, err = conn.Write([]byte(fmt.Sprintf(reqLogin, id, s.cfg.Username, s.cfg.Password))); err == nil {
	out:
		for {
			select {
			case <-ctx2.Done():
				err = fmt.Errorf("cannot login, %w", ctx.Err())
				break out
			case e, ok := <-ch:
				if ok {
					if e.Name == keyResponse {
						if e.Get(keyActionID) == id && e.Get(keyResponse) == success {
							return reader, nil
						}
					}
				}
			}
		}
	}

	return nil, err
}

func (s *amiServerImpl) logoff() {
	if s.conn == nil {
		return
	}
	id := s.getID()
	sub := s.ps.Subscribe(func(e *Event) bool {
		if e.Name == keyResponse {
			if e.Get(keyActionID) == id && e.Get(keyResponse) == goodbye {
				return true
			}
		}

		return false
	})
	defer sub.Close()

	if _, err := s.conn.Write([]byte(fmt.Sprintf(reqLogoff, id))); err != nil {
		s.logger.Error("msg", "Cannot logoff from server", "addr", s.addr, "error", err)

		return
	}

	select {
	case <-time.After(s.cfg.ActionTimeout):
		s.logger.Error("msg", "Cannot logoff from server", "addr", s.addr, "error", "timeout")
		return
	case <-sub.Events():
		s.logger.Info("msg", "Logoff successful", "addr", s.addr)
		break
	}
}

func (s *amiServerImpl) serve(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader, err := s.login(ctx, conn)
	if err != nil {
		s.logger.Error("msg", "Cannot login to server", "addr", s.addr, "error", err)
		return
	}
	defer reader.Close()

	s.logger.Info("msg", "Login to server successful", "addr", s.addr)
	s.setState(stateReady)
	s.logger.Info("msg", "Server ready", "addr", s.addr)

	for !s.shutdown.Load() {
		e, err2 := reader.Read()
		if err2 != nil {
			s.logger.Error("msg", "Cannot read on server", "addr", s.addr, "error", err2)
			s.setState(stateDisconnect)
			return
		}

		s.ps.Publish(&e)
	}
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

	start := make(chan struct{})

	go func() {
		close(start)
		s.logger.Info("msg", "Server loop started")

		for !s.shutdown.Load() {
			var (
				err  error
				conn net.Conn
			)

			for !s.shutdown.Load() {
				if conn, err = s.connect(ctx); err != nil {
					s.logger.Error("msg", "Cannot connect to server", "addr", s.addr, "error", err)
					select {
					case <-ctx.Done():
						s.logger.Info("msg", "Interrupt server reader", "addr", s.addr)
						return
					case <-time.After(s.cfg.ReconnectInterval):
					}

					continue
				}

				s.logger.Info("msg", "Connected to server", "addr", s.addr)
				break
			}

			if !s.shutdown.Load() {
				s.conn = conn
				s.serve(ctx, conn)
			}
		}

		s.logoff()
	}()

	<-start
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
	// s.conn.Close in server method

	return nil
}

func newAmiServer(cfg *config.AmiServer, cf connectionFactory, rf readerFactory,
	ps pubSubIf, logger logger.Logger) amiServer {
	return &amiServerImpl{
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		cfg:    cfg,
		logger: logger,
		state:  stateDisconnect,
		ps:     ps,
		cf:     cf,
		rf:     rf,
	}
}
