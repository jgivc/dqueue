package mocks

import "github.com/stretchr/testify/mock"

type DialerMock struct {
	mock.Mock
}

func (m *DialerMock) Notify() {
	m.Called()
}
