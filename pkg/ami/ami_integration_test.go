package ami

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/pkg/logger"
	"github.com/stretchr/testify/suite"
)

type ServerIntegrationTestSuite struct {
	suite.Suite
	// srv amiServer
	ami Ami
	ps  pubSubIf
}

func (s *ServerIntegrationTestSuite) SetupTest() {
	// logger := new(mocks.LoggerMock)
	logger := logger.New()

	cfg := &config.AmiConfig{
		ActionTimeout: 2 * time.Second,
		Servers: []config.AmiServerConfig{
			{
				Host:              "asterisk",
				Port:              5038,
				Username:          "admin",
				Secret:            "P@$$w0rD!",
				DialTimeout:       time.Second,
				ActionTimeout:     5 * time.Second,
				ReconnectInterval: 10 * time.Second,
				ReaderBuffer:      100,
			},
		},
	}
	s.ps = newPubSub(&config.PubSubConfig{PublishQueueSize: 100, SubscriberQueueSize: 1000}, logger)
	s.ami = New(cfg, &cf{}, s.ps, logger)
}

func (s *ServerIntegrationTestSuite) TestOne() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wait := make(chan struct{})

	s.T().Run("group", func(t *testing.T) {
		wg.Add(1)
		t.Run("srv", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			err := s.ami.Start(ctx)
			s.Require().NoError(err)

			defer s.ami.Close()
			<-wait
		})

		wg.Add(1)
		t.Run("subs", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()
			defer close(wait)

			subs := s.ps.Subscribe(func(e *Event) bool {
				return e.Name == keyEvent && e.Get(keyEvent) == "AsyncAGIStart"
			})
			defer subs.Close()

			e := <-subs.Events()

			err := s.ami.Answer(ctx, e.Host, e.Channel)
			s.Require().NoError(err)

			time.Sleep(time.Second)
			err = s.ami.Playback(ctx, e.Host, e.Channel, "beep")
			s.Require().NoError(err)

			time.Sleep(time.Second)
			err = s.ami.StartMOH(ctx, e.Host, e.Channel)
			s.Require().NoError(err)

			time.Sleep(3 * time.Second)
			err = s.ami.StopMOH(ctx, e.Host, e.Channel)
			s.Require().NoError(err)

			time.Sleep(time.Second)
			err = s.ami.Hangup(ctx, e.Host, e.Channel, 10)
			s.Require().NoError(err)
		})
	})

	wg.Wait()
}

func TestServerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ServerIntegrationTestSuite))
}
