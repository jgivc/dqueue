package ami

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ActionTestSuite struct {
	suite.Suite
	events []*Event
	ps     pubSubIf
	host   string
	psCfg  *config.PubSubConfig
	// req        *actionRequest
	timeout     time.Duration
	channelName string
	actionName  string
	action      *defaultAction
	buf         *bytes.Buffer
}

func (s *ActionTestSuite) getInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func (s *ActionTestSuite) SetupSuite() {
	rand.Seed(time.Now().UnixNano())
}

func (s *ActionTestSuite) SetupTest() {
	min := 20
	max := 100

	s.host = "1.2.3.4"
	c := s.getInt(min, max)
	s.events = make([]*Event, c)
	for i := 0; i < c; i++ {
		var e *Event

		if s.getInt(min, max)%3 == 0 {
			e = &Event{
				Name: keyResponse,
				Data: make(map[string]string),
			}
			switch s.getInt(1, 3) {
			case 2:
				e.Data[keyResponse] = success
			case 3:
				e.Data[keyResponse] = goodbye
			default:
				e.Data[keyResponse] = "Error"
			}

			e.Data[keyChannel] = fmt.Sprintf("TestChannel-%d", time.Now().Unix())
		} else {
			e = &Event{
				Name: keyEvent,
				Data: make(map[string]string),
			}
			e.Data[keyEvent] = "TestEvent"
			e.Data[keyChannel] = fmt.Sprintf("TestChannel-%d", time.Now().Unix())
			e.Data[keyCallerIDNum] = strconv.Itoa(s.getInt(min, max))
		}

		e.Host = s.host
		s.events[i] = e
	}

	s.channelName = fmt.Sprintf("TestChannel-%d", time.Now().Unix())
	s.actionName = "TestAction"
	s.timeout = 2 * time.Second

	s.action = newDefaultAction(s.actionName, s.timeout)
	s.action.addField(keyChannel, s.channelName)

	s.psCfg = &config.PubSubConfig{PublishQueueSize: uint(max), SubscriberQueueSize: uint(max)}
	s.ps = newPubSub(s.psCfg, new(mocks.LoggerMock))
	// s.ps = newPubSub(s.psCfg, logger.New())

	s.buf = &bytes.Buffer{}
}

func (s *ActionTestSuite) TearDownTest() {
	s.ps.Close()
}

func (s *ActionTestSuite) TestCancelledContext() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.action.run(ctx, "", s.ps, s.buf)
	s.Assert().Error(err)
	s.Assert().ErrorContains(err, "timeout")

	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%s: %s\r\n", keyAction, s.actionName))
	b.WriteString(fmt.Sprintf("%s: %s\r\n", keyActionID, s.action.req.id))
	b.WriteString(fmt.Sprintf("%s: %s\r\n", keyChannel, s.channelName))
	b.WriteString("\r\n")

	s.Assert().ElementsMatch(b.Bytes(), s.buf.Bytes())
}

func (s *ActionTestSuite) TestBadWriter() {
	w := new(writerMock)
	w.On("Write", mock.Anything).Return(0, io.ErrClosedPipe)
	_, err := s.action.run(context.Background(), "", s.ps, w)
	s.Assert().Error(err)
	s.Assert().ErrorContains(err, "cannot write")
}

func (s *ActionTestSuite) TestNoResponse() {
	var wg sync.WaitGroup
	wait := make(chan struct{})
	waitSubscriber := make(chan struct{})

	s.T().Run("group", func(t *testing.T) {
		wg.Add(1)
		t.Run("publisher", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			<-wait
			<-waitSubscriber

			for i := range s.events {
				s.ps.Publish(s.events[i])
			}
		})

		wg.Add(1)
		t.Run("subscriber", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			subs := s.ps.Subscribe(func(e *Event) bool {
				return true
			})
			defer subs.Close()

			close(waitSubscriber)
			var i int
			for range subs.Events() {
				i++
				if i == len(s.events) {
					return
				}
			}
		})

		wg.Add(1)
		t.Run("action", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			close(wait)
			_, err := s.action.run(context.Background(), s.host, s.ps, s.buf)
			s.Assert().Error(err)
			s.Assert().ErrorContains(err, "timeout")
		})
	})

	wg.Wait()
}

func (s *ActionTestSuite) TestSuccess() {
	var wg sync.WaitGroup
	waitAction := make(chan struct{})
	waitSubscriber := make(chan struct{})

	e := &Event{
		Name: keyResponse,
		Host: s.host,
		Data: make(map[string]string),
	}
	e.Data[keyResponse] = success
	e.Data[keyActionID] = s.action.req.id

	s.events = append(s.events, e)
	rand.Shuffle(len(s.events), func(i, j int) {
		s.events[i], s.events[j] = s.events[j], s.events[i]
	})

	s.T().Run("group", func(t *testing.T) {
		wg.Add(1)
		t.Run("publisher", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			<-waitAction
			<-waitSubscriber

			for i := range s.events {
				s.ps.Publish(s.events[i])
			}
		})

		wg.Add(1)
		t.Run("subscriber", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			subs := s.ps.Subscribe(func(e *Event) bool {
				return true
			})
			defer subs.Close()

			close(waitSubscriber)
			var i int
			for range subs.Events() {
				i++
				if i == len(s.events) {
					break
				}
			}
		})

		wg.Add(1)
		t.Run("action", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			close(waitAction)
			_, err := s.action.run(context.Background(), s.host, s.ps, s.buf)
			//FIXME: Some time this has error
			s.Assert().NoError(err)
		})
	})

	wg.Wait()
}

func (s *ActionTestSuite) TestErrorResponse() {
	var wg sync.WaitGroup
	waitAction := make(chan struct{})
	waitSubscriber := make(chan struct{})

	e := &Event{
		Name: keyResponse,
		Host: s.host,
		Data: make(map[string]string),
	}
	e.Data[keyResponse] = "Error"
	e.Data[keyActionID] = s.action.req.id

	s.events = append(s.events, e)
	rand.Shuffle(len(s.events), func(i, j int) {
		s.events[i], s.events[j] = s.events[j], s.events[i]
	})

	s.T().Run("group", func(t *testing.T) {
		wg.Add(1)
		t.Run("publisher", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			<-waitAction
			<-waitSubscriber

			for i := range s.events {
				s.ps.Publish(s.events[i])
			}
		})

		wg.Add(1)
		t.Run("subscriber", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			subs := s.ps.Subscribe(func(e *Event) bool {
				return true
			})
			defer subs.Close()

			close(waitSubscriber)
			var i int
			for range subs.Events() {
				i++
				if i == len(s.events) {
					break
				}
			}
		})

		wg.Add(1)
		t.Run("action", func(t *testing.T) {
			t.Parallel()
			defer wg.Done()

			close(waitAction)
			_, err := s.action.run(context.Background(), s.host, s.ps, s.buf)
			s.Assert().Error(err)
			s.Assert().ErrorContains(err, "action error")
		})
	})

	wg.Wait()
}

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}
