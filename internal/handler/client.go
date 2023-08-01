package handler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jgivc/vapp/internal/adapter"
	"github.com/jgivc/vapp/pkg/ami"
	"github.com/jgivc/vapp/pkg/ami/events"
	"github.com/jgivc/vapp/pkg/ami/fields"
	"github.com/jgivc/vapp/pkg/ami/keys"
	"github.com/jgivc/vapp/pkg/ami/types"
	"github.com/jgivc/vapp/pkg/logger"
)

const (
	keyValueCount = 2
)

type pubSub interface {
	Subscribe(filter ami.Filter) ami.Subscriber
}

type clientService interface {
	NewClient(number string, data interface{}) error
	Hangup(number string, data interface{}) error
	Operator(id, channel string) error
}

type ClientHandler struct {
	srv    clientService
	logger logger.Logger
}

func (h *ClientHandler) Register(ps pubSub) {
	go func() {
		h.logger.Info("msg", "ClientHandler start")

		subs := ps.Subscribe(func(e *ami.Event) bool {
			if e.Name == types.Event {
				return e.Get(keys.Event) == events.AsyncAGIStart || e.Get(keys.Event) == events.Hangup
			}

			return false
		})
		defer func() {
			subs.Close()
			h.logger.Info("msg", "ClientHandler done")
		}()

		for e := range subs.Events() {
			dto := &adapter.ClientDto{
				Host:    e.Host,
				Channel: e.Channel,
			}

			switch e.Get(keys.Event) {
			case events.AsyncAGIStart:
				// fmt.Println(e)
				args, err := parseArgs(e.Get(fields.Env))
				if err != nil {
					h.logger.Error("msg", "Cannot parse AsyncAGIStart env args",
						e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err)
					continue
				}
				if len(args) < 1 {
					if err2 := h.srv.NewClient(e.CallerIDNum, dto); err2 != nil {
						h.logger.Error("msg", "Cannot handle client", "number",
							e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err2)
					}
				} else {
					h.logger.Info("msg", "Operator channel", "number", e.CallerIDNum, "host", e.Host, "unique_id",
						e.Get(fields.Uniqueid), "client_id", args[0])
					if err3 := h.srv.Operator(args[0], e.Channel); err3 != nil {
						h.logger.Error("msg", "Cannot set operator channel", "number",
							e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err3)
					}
				}
			case events.Hangup:
				if err := h.srv.Hangup(e.CallerIDNum, dto); err != nil {
					h.logger.Error("msg", "Cannot hangup client", "number",
						e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err)
				}
			}
		}
	}()
}

func parseArgs(env string) ([]string, error) {
	var args []string

	raw, err := url.QueryUnescape(env)
	if err != nil {
		return nil, err
	}

	rawArgs := strings.Split(raw, "\n")
	if len(rawArgs) < 1 {
		return nil, fmt.Errorf("no args")
	}

	for _, s := range rawArgs {
		if strings.HasPrefix(s, "agi_arg") {
			kv := strings.Split(s, ": ")
			if len(kv) != keyValueCount {
				return nil, fmt.Errorf("no arg key, value")
			}

			args = append(args, kv[1])
		}
	}

	return args, nil
}

func NewClientHandler(srv clientService, logger logger.Logger) *ClientHandler {
	return &ClientHandler{
		srv:    srv,
		logger: logger,
	}
}
