package ami

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/service/mocks"
	"github.com/jgivc/vapp/pkg/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	errServerTest   = errors.New("server test error")
	endOfLineRegexp = regexp.MustCompile(`:?\s+`)
)

type AmiServerTestSuite struct {
	suite.Suite
	srv        amiServer
	cfg        *config.AmiServer
	connMock   *connectionMock
	readerMock *amiReaderMock
	subsMock   *subscriberMock
	cf         *connectionFactoryMock
	rf         *readerFactoryMock
	ps         *pubSubMock
	// logger     *mocks.LoggerMock
	logger logger.Logger
}

func (s *AmiServerTestSuite) SetupTest() {
	s.connMock = new(connectionMock)
	s.readerMock = new(amiReaderMock)
	s.subsMock = new(subscriberMock)
	s.cf = new(connectionFactoryMock)
	s.rf = new(readerFactoryMock)
	s.cfg = &config.AmiServer{}
	s.ps = new(pubSubMock)

	s.logger = new(mocks.LoggerMock)
	// s.logger = logger.New()
}

// func (s *AmiServerTestSuite) TestOne() {
// 	s.cf.On("Connect", mock.Anything, mock.Anything).Return(nil, errServerTest)
// 	s.srv = newAmiServer(s.cfg, s.cf, s.rf, s.ps, s.logger)
// 	err := s.srv.Start(context.Background())
// 	s.Require().NoError(err)
// 	_, err = s.srv.Write([]byte("123"))
// 	s.Assert().Error(err)
// 	// s.srv.Close()

// 	s.cf.AssertExpectations(s.T())
// }

// func (s *AmiServerTestSuite) TestTwo() {
// 	s.cf.On("Connect", mock.Anything, mock.Anything).Return(nil, errServerTest)
// 	s.srv = newAmiServer(s.cfg, s.cf, s.rf, s.ps, s.logger)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel()
// 	err := s.srv.Start(ctx)
// 	s.Require().NoError(err)
// 	_, err = s.srv.Write([]byte("123"))
// 	s.Assert().Error(err)

// 	s.cf.AssertExpectations(s.T())
// }

func (s *AmiServerTestSuite) TestThree() {
	s.cfg.ActionTimeout = 3 * time.Second

	wait := make(chan struct{})
	var once sync.Once
	events := make(chan *Event, 1)

	defer close(events)
	start := make(chan struct{})

	const (
		stageLogin = iota
		stagePublish
		stageLogoff
	)

	stage := stageLogin

	ids := make(chan string, 2)

	s.connMock.On("Write", mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		b, _ := args.Get(0).([]byte)

		data := endOfLineRegexp.Split(string(b), -1)
		for i, s := range data {
			if s == keyActionID {
				ids <- data[i+1]
				break
			}
		}

		if strings.Contains(string(b), "Logoff") {
			stage = stageLogoff

			go func() {
				close(start)
				id := <-ids
				events <- &Event{Name: keyResponse, Data: map[string]string{
					keyResponse: goodbye, keyActionID: id,
				}}
			}()

			<-start
		}
	})

	s.connMock.On("Close").Return(nil).Once()

	s.cf.On("Connect", mock.Anything, mock.Anything).Return(s.connMock, nil)

	s.rf.On("GetAmiReader", mock.Anything).Return(s.readerMock).Run(func(args mock.Arguments) {
		once.Do(func() {
			close(wait)
		})
	})

	rm := s.readerMock.On("Read").Return(nil)
	rm.Run(func(args mock.Arguments) {
		switch stage {
		case stageLogin:
			id := <-ids
			rm.ReturnArguments = mock.Arguments{Event{Name: keyResponse, Data: map[string]string{
				keyResponse: success, keyActionID: id,
			}}, nil}
			stage = stagePublish
		default:
			rm.ReturnArguments = mock.Arguments{Event{Name: "TestEvent", Data: map[string]string{
				"TestEvent": "blahblah",
			}}, nil}
		}
	})

	s.readerMock.On("Close").Return(nil)

	s.ps.On("Publish", mock.Anything)

	s.subsMock.On("Events").Return(events).Once()
	s.subsMock.On("Close").Once()

	s.ps.On("Subscribe", mock.Anything).Return(s.subsMock)

	s.srv = newAmiServer(s.cfg, s.cf, s.rf, s.ps, s.logger)
	err := s.srv.Start(context.Background())
	s.Require().NoError(err)

	<-wait
	time.Sleep(time.Second)
	s.srv.Close()
	<-start
	time.Sleep(time.Second)

	s.cf.AssertExpectations(s.T())
	s.connMock.AssertExpectations(s.T())
	s.readerMock.AssertExpectations(s.T())
	s.ps.AssertExpectations(s.T())
	s.subsMock.AssertExpectations(s.T())
}

func (s *AmiServerTestSuite) TearDownTest() {
	s.connMock = nil
	s.readerMock = nil
	s.cf = nil
	s.rf = nil
	s.cfg = nil
	s.ps = nil
	s.logger = nil
	// s.srv.Close()
}

func TestAmiServerTestSuite(t *testing.T) {
	suite.Run(t, new(AmiServerTestSuite))
}
