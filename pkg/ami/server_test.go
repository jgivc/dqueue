package ami

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/service/mocks"
	"github.com/jgivc/dqueue/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	errServerTest = errors.New("server test error")
)

type AmiServerTestSuite struct {
	suite.Suite
	srv      amiServer
	cfg      *config.AmiServerConfig
	connMock *connectionMock
	cf       *connectionFactoryMock
	logger   logger.Logger
}

func (s *AmiServerTestSuite) SetupTest() {
	s.connMock = new(connectionMock)
	s.cf = new(connectionFactoryMock)
	s.cfg = &config.AmiServerConfig{}

	s.logger = new(mocks.LoggerMock)
	// s.logger = logger.New()
}

func (s *AmiServerTestSuite) TestFife() { //nolint: gocognit
	srv, cln := net.Pipe()
	chSrvReader := make(chan *Event)
	_, closedCln := net.Pipe()
	closedCln.Close()

	s.cf.On("Connect",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("string")).Return(nil, errServerTest).Once()
	s.cf.On("Connect",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("string")).Return(closedCln, nil).Twice()
	s.cf.On("Connect", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string")).Return(cln, nil).Once()

	ps := newPubSub(&config.PubSubConfig{PublishQueueSize: 100, SubscriberQueueSize: 1000}, s.logger)

	cfg := &config.AmiServerConfig{
		Username:          "admin123",
		Secret:            "p@ssw0rd!23",
		ActionTimeout:     time.Second,
		ReconnectInterval: 100 * time.Millisecond,
		ReaderBuffer:      100,
	}

	var wg sync.WaitGroup
	eventsSend := make([]*Event, 0)
	eventsRecv := make([]*Event, 0)
	var running atomic.Bool

	readyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "server_ready",
		Help: "Server ready state",
	}, []string{"addr"})

	eventCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "server_receive_events",
		Help: "Receive events from server counter",
	}, []string{"addr"})

	s.srv = newAmiServer(cfg, s.cf, ps, readyGauge, eventCounter, s.logger)

	s.T().Run("group", func(t *testing.T) {
		wg.Add(1)
		t.Run("subscriber", func(t *testing.T) {
			t.Parallel()

			defer wg.Done()
			subs := ps.Subscribe(func(e *Event) bool {
				return true
			})
			defer subs.Close()

			for e := range subs.Events() {
				eventsRecv = append(eventsRecv, e)
			}
		})

		wg.Add(1)
		t.Run("server_reader", func(t *testing.T) {
			t.Parallel()

			ar := newAmiReader(srv)
			defer func() {
				close(chSrvReader)
				ar.Close()
				ps.Close()
				wg.Done()
			}()

			for {
				e, err := ar.Read()
				if err != nil {
					return
				}

				chSrvReader <- &e
			}
		})

		wg.Add(1)
		t.Run("server_writer", func(t *testing.T) {
			t.Parallel()

			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case e, ok := <-chSrvReader:
					if !ok {
						return
					}

					if e.Get("Action") == "Login" {
						s.Assert().Equal(e.Get(keyUsername), cfg.Username, "Username mismatch")
						s.Assert().Equal(e.Get(keySecret), cfg.Secret, "Secret mismatch")

						_, err := srv.Write([]byte(fmt.Sprintf("Response: Success\r\nActionID: %s\r\n\r\n", e.Get(keyActionID))))
						s.Assert().NoError(err)
						running.Store(true)
					} else if e.Get("Action") == "Logoff" {
						_, err := srv.Write([]byte(fmt.Sprintf("Response: Goodbye\r\nActionID: %s\r\n\r\n", e.Get(keyActionID))))
						s.Assert().NoError(err)
					}
				default:
					if running.Load() {
						channel := fmt.Sprintf("testchan%d", i)
						clid := strconv.Itoa(i)
						ev := &Event{
							Name:        keyEvent,
							Channel:     channel,
							CallerIDNum: clid,
							Data: map[string]string{
								keyEvent:       "TestEvent",
								keyChannel:     channel,
								keyCallerIDNum: clid,
							},
						}

						err := ev.write(srv)
						s.Assert().NoError(err)
						eventsSend = append(eventsSend, ev)
					}
				}
				// time.Sleep(2 * time.Microsecond)
			}
		})

		wg.Add(1)
		t.Run("client", func(t *testing.T) {
			t.Parallel()

			defer wg.Done()

			err := s.srv.Start(context.Background())
			s.Require().NoError(err)

			time.Sleep(3 * time.Second)
			running.Store(false)
			time.Sleep(100 * time.Millisecond)
			err = s.srv.Close()
			s.Require().NoError(err)
			time.Sleep(100 * time.Millisecond)
		})
	})

	wg.Wait()

	_, err := cln.Write([]byte("123"))
	s.Assert().ErrorContains(err, "closed")
	s.cf.AssertExpectations(s.T())
	s.Require().Equal(len(eventsSend), len(eventsRecv), "not all events receieved")
	s.Assert().ElementsMatch(eventsSend, eventsRecv)
}

func (s *AmiServerTestSuite) TearDownTest() {
	s.connMock = nil
	s.cf = nil
	s.cfg = nil
	s.logger = nil
}

func TestAmiServerTestSuite(t *testing.T) {
	suite.Run(t, new(AmiServerTestSuite))
}
