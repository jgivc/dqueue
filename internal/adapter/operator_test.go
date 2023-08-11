package adapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/jgivc/dqueue/internal/adapter"
	"github.com/jgivc/dqueue/internal/adapter/mocks"
	"github.com/jgivc/dqueue/internal/entity"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	errOperatorRepo = errors.New("operatorRepo error")
)

type OperatorRepoTestSuite struct {
	suite.Suite
	repo      *adapter.OperatorRepo
	api       *mocks.APIClientMock
	operators []*entity.Operator
	rcEmpty   io.ReadCloser
}

func (s *OperatorRepoTestSuite) SetupSuite() {
	s.rcEmpty = io.NopCloser(bytes.NewBuffer(nil))
}

func (s *OperatorRepoTestSuite) setBusy(number string, busy bool) {
	for i := range s.operators {
		if s.operators[i].Number == number {
			s.operators[i].SetBusy(busy)

			return
		}
	}

	s.T().Fatal("No operator found")
}

func (s *OperatorRepoTestSuite) getRC() io.ReadCloser {
	b, err := json.Marshal(adapter.RawOperators{Operators: s.operators})
	if err != nil {
		s.T().Fatal(err)
	}

	return io.NopCloser(bytes.NewBuffer(b))
}

func (s *OperatorRepoTestSuite) SetupTest() {
	s.api = new(mocks.APIClientMock)
	s.repo = adapter.NewOperatorRepoWithAPIClient(s.api, &mocks.LoggerMock{})
	s.operators = []*entity.Operator{
		entity.NewOperator("1111", "Ivan", "Ivanov"),
		entity.NewOperator("2222", "Petr", "Petrov"),
		entity.NewOperator("3333", "Semen", "Semenov"),
	}
}

func (s *OperatorRepoTestSuite) TestGetOperatorsFailOne() {
	s.api.On("Get", mock.Anything).Return(nil, errOperatorRepo)

	_, err := s.repo.GetOperators(context.Background())
	s.Require().Error(err)

	s.api.AssertExpectations(s.T())
}

func (s *OperatorRepoTestSuite) TestGetOperatorsFailTwo() {
	s.api.On("Get", mock.Anything).Return(s.rcEmpty, nil)

	_, err := s.repo.GetOperators(context.Background())
	s.Require().Error(err)

	s.api.AssertExpectations(s.T())
}

func (s *OperatorRepoTestSuite) TestGetOperatorsOne() {
	s.api.On("Get", mock.Anything).Return(s.getRC(), nil)

	operators, err := s.repo.GetOperators(context.Background())
	s.Require().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)
	s.api.AssertExpectations(s.T())

	err = s.repo.SetBusy("NotExists", true)
	s.Assert().NoError(err)

	s.setBusy("1111", true)
	err = s.repo.SetBusy("1111", true)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)

	s.setBusy("2222", true)
	err = s.repo.SetBusy("2222", true)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)

	s.setBusy("1111", false)
	err = s.repo.SetBusy("1111", false)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)
}

func (s *OperatorRepoTestSuite) TestGetOperatorsTwo() {
	s.api.On("Get", mock.Anything).Return(s.getRC(), nil)

	operators, err := s.repo.GetOperators(context.Background())
	s.Require().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)
	s.api.AssertExpectations(s.T())

	s.api.On("Get", mock.Anything).Unset()
	s.operators = append(s.operators, entity.NewOperator("4444", "Anna", "Annova"))
	s.api.On("Get", mock.Anything).Return(s.getRC(), nil)

	operators, err = s.repo.GetOperators(context.Background())
	s.Require().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)
	s.api.AssertExpectations(s.T())

	s.api.On("Get", mock.Anything).Unset()
	s.operators = append(s.operators[:2], s.operators[3:]...)
	s.api.On("Get", mock.Anything).Return(s.getRC(), nil)

	operators, err = s.repo.GetOperators(context.Background())
	s.Require().NoError(err)
	s.Assert().ElementsMatch(s.operators, operators)
	s.api.AssertExpectations(s.T())
}

func TestOperatorRepoTestSuite(t *testing.T) {
	suite.Run(t, new(OperatorRepoTestSuite))
}
