package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/pkg/ami"
	"github.com/jgivc/vapp/pkg/logger"
	"github.com/stretchr/testify/suite"
)

type AmiIntegrationTestSuite struct {
	suite.Suite
	// srv amiServer
	ami ami.Ami
}

func (s *AmiIntegrationTestSuite) SetupTest() {
	// logger := new(mocks.LoggerMock)
	logger := logger.New()

	cfg := &config.AmiConfig{
		ActionTimeout: 2 * time.Second,
		PSConfig: config.PubSubConfig{
			PublishQueueSize:    100,
			SubscriberQueueSize: 1000,
		},
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
	// s.ps = newPubSub(&config.PubSubConfig{PublishQueueSize: 100, SubscriberQueueSize: 1000}, logger)
	// s.ami = newAmi(cfg, newConnectionFactory(), s.ps, logger)
	s.ami = ami.New(cfg, logger)
}

func (s *AmiIntegrationTestSuite) TestOne() {
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

			subs := s.ami.Subscribe(func(e *ami.Event) bool {
				return e.Name == "Event" && e.Get("Event") == "AsyncAGIStart"
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

func TestAmiIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AmiIntegrationTestSuite))
}
