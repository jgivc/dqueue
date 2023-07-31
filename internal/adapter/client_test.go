package adapter_test

import (
	"testing"

	"github.com/jgivc/vapp/internal/adapter"
	"github.com/jgivc/vapp/internal/adapter/mocks"
	"github.com/stretchr/testify/suite"
)

type ClientRepoTestSuite struct {
	suite.Suite
	repo *adapter.ClientRepo
}

func (s *ClientRepoTestSuite) SetupTest() {
	s.repo = adapter.NewClientRepo(&mocks.LoggerMock{})
}

func (s *ClientRepoTestSuite) AfterTest(_, _ string) {
	s.repo.Close()
}

func (s *ClientRepoTestSuite) TestNewFailNoData() {
	_, err := s.repo.New("1234", nil)

	s.Require().Error(err)

	_, err = s.repo.New("1234", struct{}{})
	s.Require().Error(err)
}

func (s *ClientRepoTestSuite) TestNewFailNoID() {
	data := new(mocks.GetUniqueIDMock)
	data.On("GetUniqueID").Return("")

	_, err := s.repo.New("1234", data)

	s.Require().Error(err)
	data.AssertExpectations(s.T())
}

func (s *ClientRepoTestSuite) TestNewFailNoNumber() {
	data := new(mocks.GetUniqueIDMock)
	data.On("GetUniqueID").Return("12345")

	_, err := s.repo.New("", data)

	s.Require().Error(err)
}

func (s *ClientRepoTestSuite) TestNew() {
	number := "1234"
	id := "1234567890"
	data := new(mocks.GetUniqueIDMock)
	data.On("GetUniqueID").Return(id)

	client, err := s.repo.New(number, data)
	s.Require().NoError(err)

	s.Assert().Equal(number, client.Number, "Number mismatch")
	s.Assert().Equal(data, client.Data, "Data mismatch")
	s.Assert().True(client.IsAlive())

	data.AssertExpectations(s.T())
}

func (s *ClientRepoTestSuite) TestRemoveFailNoData() {
	err := s.repo.Remove("", nil)

	s.Require().Error(err)

	err = s.repo.Remove("", struct{}{})
	s.Require().Error(err)
}

func (s *ClientRepoTestSuite) TestRemoveFailNoID() {
	data := new(mocks.GetUniqueIDMock)
	data.On("GetUniqueID").Return("")

	err := s.repo.Remove("", data)

	s.Require().Error(err)
	data.AssertExpectations(s.T())
}

func (s *ClientRepoTestSuite) TestRemove() {
	number := "1234"
	id := "1234567890"
	data := new(mocks.GetUniqueIDMock)
	data.On("GetUniqueID").Return(id)

	client, err := s.repo.New(number, data)
	s.Require().NoError(err)

	s.Assert().Equal(number, client.Number, "Number mismatch")
	s.Assert().Equal(data, client.Data, "Data mismatch")
	s.Assert().True(client.IsAlive())

	id2 := "0987654321"
	data2 := new(mocks.GetUniqueIDMock)
	data2.On("GetUniqueID").Return(id2)

	err = s.repo.Remove(number, data2)
	s.Require().Error(err)
	s.Assert().True(client.IsAlive())

	err = s.repo.Remove(number, data)
	s.Require().NoError(err)
	s.Assert().False(client.IsAlive())

	data.AssertExpectations(s.T())
	data2.AssertExpectations(s.T())
}

func TestClientRepoTestSuite(t *testing.T) {
	suite.Run(t, new(ClientRepoTestSuite))
}
