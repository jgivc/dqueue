package handler

import (
	"github.com/jgivc/vapp/internal/adapter"
	"github.com/jgivc/vapp/pkg/ami"
	"github.com/jgivc/vapp/pkg/ami/events"
	"github.com/jgivc/vapp/pkg/ami/fields"
	"github.com/jgivc/vapp/pkg/ami/keys"
	"github.com/jgivc/vapp/pkg/ami/types"
	"github.com/jgivc/vapp/pkg/logger"
)

type pubSub interface {
	Subscribe(filter ami.Filter) ami.Subscriber
}

type clientService interface {
	NewClient(number string, data interface{}) error
	Hangup(number string, data interface{}) error
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
			dto := adapter.ClientDto{
				Host:    e.Host,
				Channel: e.Channel,
			}

			switch e.Get(keys.Event) {
			case events.AsyncAGIStart:
				if err := h.srv.NewClient(e.CallerIDNum, dto); err != nil {
					h.logger.Error("msg", "Cannot handle client", "number",
						e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err)
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

func NewClientHandler(srv clientService, logger logger.Logger) *ClientHandler {
	return &ClientHandler{
		srv:    srv,
		logger: logger,
	}
}
