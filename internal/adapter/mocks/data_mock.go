package mocks

import (
	"github.com/stretchr/testify/mock"
)

type GetUniqueIDMock struct {
	mock.Mock
}

func (m *GetUniqueIDMock) GetUniqueID() string {
	args := m.Called()

	return args.String(0)
}
