package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/entity"
	"github.com/jgivc/dqueue/internal/service"
	"github.com/jgivc/dqueue/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	errTestError = errors.New("test error")
)

type ClientTestSuite struct {
	suite.Suite
	voip          *mocks.VoipMock
	queue         *mocks.QueueMock
	repo          *mocks.ClientRepoMock
	dialer        *mocks.DialerMock
	logger        *mocks.LoggerMock
	clientService *service.ClientService
	cfg           *config.ClientService
}

func (s *ClientTestSuite) SetupTest() {
	s.voip = new(mocks.VoipMock)
	s.queue = new(mocks.QueueMock)
	s.repo = new(mocks.ClientRepoMock)
	s.dialer = new(mocks.DialerMock)
	s.logger = new(mocks.LoggerMock)
	s.cfg = &config.ClientService{}
}

func (s *ClientTestSuite) BeforeTest(_, testName string) {
	switch testName {
	case "TestZeroBuffer":
	default:
		s.cfg.ChannelBufferSize = 10
		s.cfg.HandleClientTimeout = 2 * time.Second
	}

	s.clientService = service.NewClientService(
		s.cfg, s.voip, s.queue, s.repo, s.dialer, s.logger,
	)
}

func (s *ClientTestSuite) TestZeroBuffer() {
	err := s.clientService.NewClient("", nil)
	s.Require().Error(err)
}

func (s *ClientTestSuite) TestOne() {
	err := s.clientService.NewClient("", nil)
	s.Require().NoError(err)
}

func (s *ClientTestSuite) TestFailVoipOne() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := entity.NewClient("1234", nil)

	s.repo.On("New", mock.Anything, mock.Anything).Return(&client, nil)
	s.voip.On("Answer", mock.Anything, mock.Anything).Return(errTestError)
	s.clientService.Start(ctx)

	err := s.clientService.NewClient("1234", nil)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)
	s.voip.AssertExpectations(s.T())
}

func (s *ClientTestSuite) TestFailVoipTwo() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := entity.NewClient("1234", nil)

	s.repo.On("New", mock.Anything, mock.Anything).Return(&client, nil)
	s.voip.On("Answer", mock.Anything, mock.Anything).Return(nil)
	s.voip.On("StartMOH", mock.Anything, mock.Anything).Return(errTestError)
	s.queue.On("Push", mock.Anything).Return(nil)
	s.clientService.Start(ctx)

	err := s.clientService.NewClient("1234", nil)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)
	s.voip.AssertExpectations(s.T())
	s.queue.AssertNotCalled(s.T(), "Push", mock.Anything)
}

func (s *ClientTestSuite) TestFailQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := entity.NewClient("1234", nil)

	s.repo.On("New", mock.Anything, mock.Anything).Return(&client, nil)
	s.voip.On("Answer", mock.Anything, mock.Anything).Return(nil)
	s.voip.On("StartMOH", mock.Anything, mock.Anything).Return(nil)
	s.queue.On("Push", mock.Anything).Return(errTestError)
	s.dialer.On("Notify")

	s.clientService.Start(ctx)

	err := s.clientService.NewClient("1234", nil)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)
	s.voip.AssertExpectations(s.T())
	s.queue.AssertExpectations(s.T())
	s.dialer.AssertNotCalled(s.T(), "Notify")
}

func (s *ClientTestSuite) TestOk() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := entity.NewClient("1234", nil)

	s.repo.On("New", mock.Anything, mock.Anything).Return(&client, nil)
	s.voip.On("Answer", mock.Anything, mock.Anything).Return(nil)
	s.voip.On("StartMOH", mock.Anything, mock.Anything).Return(nil)
	s.queue.On("Push", mock.Anything).Return(nil)
	s.dialer.On("Notify")

	s.clientService.Start(ctx)

	err := s.clientService.NewClient("1234", nil)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)
	s.voip.AssertExpectations(s.T())
	s.queue.AssertExpectations(s.T())
	s.dialer.AssertExpectations(s.T())
}

func (s *ClientTestSuite) TestFailClient() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := entity.NewClient("1234", nil)
	client.Close()

	s.repo.On("New", mock.Anything, mock.Anything).Return(&client, nil)
	s.voip.On("Answer", mock.Anything, mock.Anything).Return(nil)
	s.clientService.Start(ctx)

	err := s.clientService.NewClient("1234", nil)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)
	s.voip.AssertNotCalled(s.T(), "Answer", mock.Anything, mock.Anything)
}

func (s *ClientTestSuite) AfterTest(_, _ string) {
	s.voip = nil
	s.queue = nil
	s.repo = nil
	s.dialer = nil
	s.logger = nil
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
