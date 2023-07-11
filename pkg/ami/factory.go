package ami

import (
	"context"
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
