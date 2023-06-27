package adapter_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type OperatorRepoTestCase struct {
	suite.Suite
}

func TestOperatorRepoTestCase(t *testing.T) {
	suite.Run(t, new(OperatorRepoTestCase))
}
