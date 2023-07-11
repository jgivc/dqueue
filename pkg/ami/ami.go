package ami

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	errAmiError = errors.New("ami error")
)

type Ami interface {
	Subscribe(filter Filter) Subscriber

	// Answer(ctx context.Context, client *entity.Client) error
	// Playback(ctx context.Context, client *entity.Client, fileName string) error
	// StartMOH(ctx context.Context, client *entity.Client) error
	// StopMOH(ctx context.Context, client *entity.Client) error
	// Dial(ctx context.Context, client *entity.Client, operators ...entity.Operator) error
	// Hangup(ctx context.Context, client *entity.Client) error

	Answer(ctx context.Context, host string, channel string) error
	Playback(ctx context.Context, host string, channel string, fileName string) error
	StartMOH(ctx context.Context, host string, channel string) error
	StopMOH(ctx context.Context, host string, channel string) error
	Hangup(ctx context.Context, host string, channel string, cause int) error
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
