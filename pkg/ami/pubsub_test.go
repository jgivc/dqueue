package ami

import (
	"os"
	"sync"
	"testing"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/service/mocks"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

const (
	pubsubEventsFile        = "testdata/pubsub_events.yml"
	filterCallerIDNum       = "3333"
	unsubscribeCaptureCount = 2
)

type PubSubTestSuite struct {
	suite.Suite
	ps     *pubSub
	events []*Event
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

	s.ps = newPubSub(&config.PubSubConfig{
		PublishQueueSize:    10,
		SubscriberQueueSize: 10,
		// }, logger.New())
	}, &mocks.LoggerMock{})
}

func (s *PubSubTestSuite) TestSubscribe() {
	out := make([]*Event, 0)
	outFiltered := make([]*Event, 0)
	var wg sync.WaitGroup

	s.T().Run("group", func(t *testing.T) {
		subs := s.ps.Subscribe(func(e *Event) bool {
			return true
		})

		subsFiltered := s.ps.Subscribe(func(e *Event) bool {
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

	expectedOutFiltered := make([]*Event, 0)
	for i, e := range s.events {
		if e.CallerIDNum == filterCallerIDNum {
			expectedOutFiltered = append(expectedOutFiltered, s.events[i])
		}
	}
	s.Assert().ElementsMatch(expectedOutFiltered, outFiltered)
}

func (s *PubSubTestSuite) TestUnsubscribe() {
	out := make([]*Event, 0)
	out2 := make([]*Event, 0)
	var wg sync.WaitGroup

	s.T().Run("group", func(t *testing.T) {
		subs := s.ps.Subscribe(func(e *Event) bool {
			return true
		})

		subs2 := s.ps.Subscribe(func(e *Event) bool {
			return true
		})

		wait := make(chan struct{})

		wg.Add(1)
		t.Run("subscribe", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()
			defer subs.Close()

			for e := range subs.Events() {
				out = append(out, e)
				if len(out) == len(s.events) {
					return
				}
			}
		})

		wg.Add(1)
		t.Run("unsubscribe", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			for e := range subs2.Events() {
				out2 = append(out2, e)
				if len(out2) == unsubscribeCaptureCount {
					subs2.Close()
					close(wait)
				}
			}
		})

		wg.Add(1)
		t.Run("publish", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			for i := range s.events {
				s.ps.Publish(s.events[i])
				if i == unsubscribeCaptureCount-1 {
					<-wait
				}
			}
		})
	})

	wg.Wait()
	s.Assert().ElementsMatch(s.events, out)
	s.Assert().ElementsMatch(s.events[:unsubscribeCaptureCount], out2)
}

func (s *PubSubTestSuite) TearDownSuite() {
	s.ps.Close()
}

func TestPubSubTestSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}
