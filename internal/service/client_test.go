package service_test

import (
	"testing"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/service"
	"github.com/jgivc/vapp/internal/service/mock"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	voip          service.VoipAdapter
	queue         service.Queue
	repo          service.ClientRepo
	dialer        service.Dialer
	logger        service.Logger
	clientService service.ClientService
	cfg           *config.ClientService
}

func (s *ClientTestSuite) SetupTest() {
	s.voip = new(mock.VoipMock)
	s.queue = new(mock.QueueMock)
	s.repo = new(mock.ClientRepoMock)
	s.dialer = new(mock.DialerMock)
	s.logger = new(mock.LoggerMock)
	s.cfg = &config.ClientService{}
}

func (s *ClientTestSuite) BeforeTest(_, testName string) {
	switch testName {
	case "TestZeroBuffer":
	default:
		s.cfg.ChannelBufferSize = 10
		s.cfg.HandleClientTimeout = 2 * time.Second
	}

	s.clientService = *service.NesClientService(
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
