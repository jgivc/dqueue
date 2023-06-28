package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type APIClientMock struct {
	mock.Mock
}

func (m *APIClientMock) Get(ctx context.Context) (io.ReadCloser, error) {
	args := m.Called(ctx)

	var (
		rc io.ReadCloser
		ok bool
	)

	if args.Get(0) != nil {
		if rc, ok = args.Get(0).(io.ReadCloser); !ok {
			panic("cannot convert omterface to io.ReadCloser")
		}
	}

	return rc, args.Error(1)
}
