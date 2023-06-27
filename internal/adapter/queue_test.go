package adapter_test

import (
	"testing"

	"github.com/jgivc/vapp/internal/adapter"
	"github.com/jgivc/vapp/internal/entity"
	"github.com/stretchr/testify/suite"
)

const (
	defaultQueueMaxClients = 2
)

type QueueTestSuite struct {
	suite.Suite
	queue adapter.Queue
}

func (s *QueueTestSuite) BeforeTest(_, testName string) {
	switch testName {
	case "":
	default:
		s.queue = *adapter.NewQueue(defaultQueueMaxClients)
	}
}

func (s *QueueTestSuite) AfterTest(_, _ string) {
	s.queue.Close()
}

func (s *QueueTestSuite) TestOne() {
	s.Assert().False(s.queue.IsFull())
	s.Assert().False(s.queue.HasClients())

	_, err := s.queue.Pop()
	s.Require().Error(err, "Queue is empty")

	err = s.queue.Push(nil)
	s.Assert().Error(err)

	client := entity.NewClient("1234", 111)

	err = s.queue.Push(&client)
	s.Assert().NoError(err)
	s.Assert().True(s.queue.HasClients())
	s.Assert().False(s.queue.IsFull())

	client2 := entity.NewClient("5678", 222)

	err = s.queue.Push(&client2)
	s.Assert().NoError(err)
	s.Assert().NoError(err)
	s.Assert().True(s.queue.IsFull())

	client3 := entity.NewClient("9012", 333)
	err = s.queue.Push(&client3)
	s.Assert().Error(err, "Queue is full")

	cl, err := s.queue.Pop()
	s.Assert().NoError(err)
	s.Assert().EqualValues(&client, cl)
	s.Assert().False(s.queue.IsFull())
	s.Assert().True(s.queue.HasClients())

	cl, err = s.queue.Pop()
	s.Assert().NoError(err)
	s.Assert().EqualValues(&client2, cl)
	s.Assert().False(s.queue.IsFull())
	s.Assert().False(s.queue.HasClients())

	_, err = s.queue.Pop()
	s.Assert().Error(err)
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}
