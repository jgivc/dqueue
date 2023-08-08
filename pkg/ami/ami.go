package ami

import (
	"context"
	"errors"
	"fmt"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/pkg/logger"
)

var (
	errAmiError = errors.New("ami error")
)

type ami struct {
	cfg     *config.AmiConfig
	servers map[string]amiServer
	ps      pubSubIf
	cf      connectionFactory
	logger  logger.Logger
}

func (a *ami) getServer(host string) (amiServer, error) {
	if _, exists := a.servers[host]; exists {
		return a.servers[host], nil
	}

	return nil, fmt.Errorf("%w: cannot find server %s", errAmiError, host)
}

func (a *ami) Subscribe(filter Filter) Subscriber {
	return a.ps.Subscribe(filter)
}

func (a *ami) Close() error {
	a.ps.Close()

	hosts := make([]string, len(a.servers))
	i := 0
	for key := range a.servers {
		hosts[i] = key
		i++
	}

	for _, host := range hosts {
		if err := a.servers[host].Close(); err != nil {
			a.logger.Error("msg", "Cannot close server", "host", host, "error", err)
		}
		delete(a.servers, host)
	}

	return nil
}

func (a *ami) Start(ctx context.Context) error {
	for i := range a.cfg.Servers {
		cfg := a.cfg.Servers[i]
		srv := newAmiServer(&cfg, a.cf, a.ps, a.logger)
		if err := srv.Start(ctx); err != nil {
			return fmt.Errorf("cannot connect to server %s:%d, %w", cfg.Host, cfg.Port, err)
		}

		a.servers[cfg.Host] = srv
	}

	return nil
}

func newAmi(cfg *config.AmiConfig, cf connectionFactory, ps pubSubIf, logger logger.Logger) Ami {
	return &ami{
		cfg:     cfg,
		servers: make(map[string]amiServer),
		ps:      ps,
		cf:      cf,
		logger:  logger,
	}
}

func New(cfg *config.AmiConfig, logger logger.Logger) Ami {
	ps := newPubSub(&cfg.PSConfig, logger)
	return newAmi(cfg, newConnectionFactory(), ps, logger)
}
