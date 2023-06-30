package ami_test

import (
	"os"
	"sync"
	"testing"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/service/mocks"
	"github.com/jgivc/vapp/pkg/ami"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

const (
	pubsubEventsFile  = "testdata/pubsub_events.yml"
	filterCallerIDNum = "3333"
)

type PubSubTestSuite struct {
	suite.Suite
	ps     ami.PubSub
	events []*ami.Event
}

func (s *PubSubTestSuite) SetupSuite() {
	ofile, err := os.Open(pubsubEventsFile)
	if err != nil {
		s.Require().NoError(err)
	}
	defer ofile.Close()

	dec := yaml.NewDecoder(ofile)
	if err2 := dec.Decode(&s.events); err2 != nil {
		s.Require().NoError(err)
	}

	s.ps = ami.NewPubSub(&config.PubSubConfig{
		PublishQueueSize:    10,
		SubscriberQueueSize: 10,
	}, &mocks.LoggerMock{})
}

func (s *PubSubTestSuite) TestSubscribe() {
	out := make([]*ami.Event, 0)
	outFiltered := make([]*ami.Event, 0)
	var wg sync.WaitGroup

	s.T().Run("group", func(t *testing.T) {
		subs := s.ps.Subscribe(func(e *ami.Event) bool {
			return true
		})

		subsFiltered := s.ps.Subscribe(func(e *ami.Event) bool {
			return e.CallerIDNum == filterCallerIDNum
		})

		wg.Add(1)
		t.Run("subscribe", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()
			defer subsFiltered.Close()
			defer subs.Close()

			for e := range subs.Events() {
				out = append(out, e)
				if len(out) == len(s.events) {
					return
				}
			}
		})

		wg.Add(1)
		t.Run("filtered", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			for e := range subsFiltered.Events() {
				outFiltered = append(outFiltered, e)
			}
		})

		wg.Add(1)
		t.Run("publish", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			for i := range s.events {
				s.ps.Publish(s.events[i])
			}
		})
	})

	wg.Wait()
	s.Assert().ElementsMatch(s.events, out)

	expectedOutFiltered := make([]*ami.Event, 0)
	for i, e := range s.events {
		if e.CallerIDNum == filterCallerIDNum {
			expectedOutFiltered = append(expectedOutFiltered, s.events[i])
		}
	}
	s.Assert().ElementsMatch(expectedOutFiltered, outFiltered)
}

func (s *PubSubTestSuite) TearDownSuite() {
	s.ps.Close()
}

func TestPubSubTestSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}
