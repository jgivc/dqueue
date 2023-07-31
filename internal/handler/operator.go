package handler

import (
	"github.com/jgivc/vapp/pkg/ami"
	"github.com/jgivc/vapp/pkg/ami/events"
	"github.com/jgivc/vapp/pkg/ami/fields"
	"github.com/jgivc/vapp/pkg/ami/keys"
	"github.com/jgivc/vapp/pkg/ami/types"
	"github.com/jgivc/vapp/pkg/logger"
)

type operatorService interface {
	SetBusy(number string, busy bool) error
}

type OperatorHandler struct {
	srv    operatorService
	logger logger.Logger
}

func (h *OperatorHandler) Register(ps pubSub) {
	go func() {
		h.logger.Info("msg", "OperatorHandler start")

		subs := ps.Subscribe(func(e *ami.Event) bool {
			if e.Name == types.Event {
				eventName := e.Get(keys.Event)
				return eventName == events.Newchannel || eventName == events.Hangup || eventName == events.BridgeEnter
			}

			return false
		})
		defer func() {
			subs.Close()
			h.logger.Info("msg", "OperatorHandler done")
		}()

		for e := range subs.Events() {
			switch e.Get(keys.Event) {
			case events.Hangup:
				if err := h.srv.SetBusy(e.CallerIDNum, false); err != nil {
					h.logger.Error("msg", "Cannot set operator free", "number",
						e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err)
				}
				// case events.Newchannel, events.BridgeEnter:
			case events.BridgeEnter:
				num := e.Get(fields.ConnectedLineNum)
				if err := h.srv.SetBusy(num, true); err != nil {
					h.logger.Error("msg", "Cannot set operator busy", "number",
						num, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err)
				}
				fallthrough
			case events.Newchannel:
				if err := h.srv.SetBusy(e.CallerIDNum, true); err != nil {
					h.logger.Error("msg", "Cannot set operator busy", "number",
						e.CallerIDNum, "host", e.Host, "unique_id", e.Get(fields.Uniqueid), "error", err)
				}
			}
		}
	}()
}

func NewOperatorHandler(srv operatorService, logger logger.Logger) *OperatorHandler {
	return &OperatorHandler{
		srv:    srv,
		logger: logger,
	}
}
