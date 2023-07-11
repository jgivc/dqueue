package ami

import (
	"errors"
	"fmt"
	"time"
)

var (
	errAmiError = errors.New("ami error")
)

type Ami interface {
	Subscribe(filter Filter) Subscriber

	Hangup(host string, channel string, cause int) error
}

type ami struct {
	actionTimeout time.Duration
	servers       map[string]amiServer
	ps            pubSubIf
}

func (a *ami) getServer(host string) (amiServer, error) {
	if _, exists := a.servers[host]; exists {
		return a.servers[host], nil
	}

	return nil, fmt.Errorf("%w: cannot find server %s", errAmiError, host)
}
