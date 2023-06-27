package mocks

import "github.com/stretchr/testify/mock"

type LoggerMock struct {
	mock.Mock
}

func (m *LoggerMock) Info(_ ...interface{}) {

}

func (m *LoggerMock) Warn(_ ...interface{}) {

}

func (m *LoggerMock) Error(_ ...interface{}) {

}
