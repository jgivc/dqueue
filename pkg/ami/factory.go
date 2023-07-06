package ami

import (
	"context"
	"io"
	"net"
)

const (
	network = "tcp"
)

type cf struct {
}

func (f *cf) Connect(ctx context.Context, addr string) (net.Conn, error) {
	var dialer net.Dialer

	return dialer.DialContext(ctx, network, addr)
}

func newConnectionFactory() *cf {
	return &cf{}
}

type rf struct {
}

func (f *rf) GetAmiReader(r io.Reader) amiReaderIf {
	return newAmiReader(r)
}

func newReaderFactory() *rf {
	return &rf{}
}
