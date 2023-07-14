package ami

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
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
	mux    sync.Mutex // Protect writer against simultaneous use
	addr   string
	cfg    *config.AmiServerConfig
	conn   net.Conn
	logger logger.Logger
	state  serverState
	cf     connectionFactory
	ps     pubSubIf
	stop   chan struct{}
}

func (s *amiServerImpl) connect(ctx context.Context) (net.Conn, error) {
	ctx2, cancel := context.WithTimeout(ctx, s.cfg.DialTimeout)
	defer cancel()

	return s.cf.Connect(ctx2, s.addr)
}

func (s *amiServerImpl) getID() string {
	return uuid.New().String()
}

func (s *amiServerImpl) login(ch chan *Event) error {
	ctx2, cancel := context.WithTimeout(context.Background(), s.cfg.ActionTimeout)
	defer cancel()

	id := s.getID()

	var err error
	_, err = s.conn.Write([]byte(fmt.Sprintf(reqLogin, id, s.cfg.Username, s.cfg.Secret)))
	if err != nil {
		return fmt.Errorf("cannot send to server: %w", err)
	}

out:
	for {
		select {
		case <-ctx2.Done():
			err = fmt.Errorf("cannot login, %w", ctx2.Err())
			break out
		case e, ok := <-ch:
			if ok {
				if e.Name == keyResponse {
					if e.Get(keyActionID) == id && e.Get(keyResponse) == success {
						return nil
					}
				}
			}
		}
	}

	return err
}

func (s *amiServerImpl) logoff(ch chan *Event) {
	id := s.getID()

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ActionTimeout)
	defer cancel()

	if _, err := s.conn.Write([]byte(fmt.Sprintf(reqLogoff, id))); err != nil {
		s.logger.Error("msg", "Cannot logoff from server", "addr", s.addr, "error", err)

		return
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Error("msg", "Cannot logoff from server", "addr", s.addr, "error", ctx.Err())
			return
		case e, ok := <-ch:
			if ok {
				if e.Name == keyResponse {
					if e.Get(keyActionID) == id && e.Get(keyResponse) == goodbye {
						s.logger.Info("msg", "Logoff successful", "addr", s.addr)
						return
					}
				}
			}
		}
	}
}

func (s *amiServerImpl) serve(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := newAmiReader(conn)
	defer reader.Close()

	var i int
	ch := make(chan *Event, s.cfg.ReaderBuffer)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			default:
				e, err2 := reader.Read()
				i++
				if err2 != nil {
					s.logger.Error("msg", "Cannot read on server", "addr", s.addr, "error", err2)
					s.setState(stateDisconnect)
					return
				}

				e.Host = s.cfg.Host
				ch <- &e
			}
		}
	}()

	if err := s.login(ch); err != nil {
		s.logger.Error("msg", "Cannot login to server", "addr", s.addr, "error", err)

		return
	}

	s.logger.Info("msg", "Login to server successful", "addr", s.addr)
	s.setState(stateReady)
	s.logger.Info("msg", "Server ready", "addr", s.addr)

	var j int
	for {
		select {
		case <-ctx.Done():
			s.logoff(ch)
			return
		case <-s.stop:
			s.logoff(ch)
			return
		case e := <-ch:
			j++
			s.ps.Publish(e)
		}
	}
}

func (s *amiServerImpl) Start(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.state == stateShutdown {
		return fmt.Errorf("%w shutdown flag", errAmiServer)
	}

	s.state = stateDisconnect
	start := make(chan struct{})

	go func() {
		close(start)
		s.logger.Info("msg", "Server loop started")

		defer s.setState(stateDisconnect)

		for {
			var (
				err  error
				conn net.Conn
			)

		out:
			for {
				select {
				case <-ctx.Done():
					return
				case <-s.stop:
					return
				default:
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
					break out
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			default:
				s.conn = conn
				s.serve(ctx, conn)
			}
		}
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

	select {
	case _, ok := <-s.stop:
		if !ok {
			return fmt.Errorf("%w already shutdown", errAmiServer)
		}
	default:
	}

	s.state = stateShutdown
	close(s.stop)
	// s.conn.Close in serve method

	return nil
}

func newAmiServer(cfg *config.AmiServerConfig, cf connectionFactory,
	ps pubSubIf, logger logger.Logger) amiServer {
	return &amiServerImpl{
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		cfg:    cfg,
		logger: logger,
		state:  stateDisconnect,
		ps:     ps,
		cf:     cf,
		stop:   make(chan struct{}),
	}
}
