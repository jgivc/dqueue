package ami

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	agiStageRequest = iota
	agiStageResponse
	agiStageFail
	agiStageOK
)

/*
Action: AGI
ActionID: <value>
Channel: <value>
Command: <value>
CommandID: <value>
*/

type agiActionHandler struct {
	req   *actionRequest
	ch    chan struct{}
	obj   interface{}
	err   error
	stage int
}

func (h *agiActionHandler) filter(e *Event) bool {
	if e.Name == keyResponse {
		return e.Get(keyActionID) == h.req.id
	}

	if e.Name == keyEvent && e.Get(keyEvent) == actionAsyncAGIExec {
		return e.Get(keyCommandID) == h.req.id
	}

	return false
}

func (h *agiActionHandler) channelClosed() bool {
	select {
	case _, ok := <-h.ch:
		if !ok {
			return true
		}
	default:
	}

	return false
}

func (h *agiActionHandler) handle(e *Event) {
	switch h.stage {
	case agiStageRequest:
		if e.Get(keyResponse) == success {
			h.stage = agiStageResponse
		} else {
			h.err = fmt.Errorf("agi error: %w", errActionError)
			close(h.ch)
			h.stage = agiStageFail
		}
	case agiStageResponse:
		if e.Get(e.Name) == actionAsyncAGIExec {
			var val string
			val, err := url.QueryUnescape(e.Get(keyResult))
			if err != nil {
				h.err = fmt.Errorf("cannot decode value: %w", err)
				close(h.ch)
				h.stage = agiStageFail
				return
			}

			if !strings.HasPrefix(val, "200 ") {
				h.err = fmt.Errorf("agi error, agi acton response is not 200: %w", errActionError)
				h.stage = agiStageFail
			} else {
				h.stage = agiStageOK
			}

			close(h.ch)
		}
	}
}

func (h *agiActionHandler) done() <-chan struct{} {
	return h.ch
}

func (h *agiActionHandler) result() (interface{}, error) {
	if !h.channelClosed() {
		return nil, fmt.Errorf("handler is not done")
	}
	return h.obj, h.err
}

func newAgiActionHandler(req *actionRequest) *agiActionHandler {
	return &agiActionHandler{
		req:   req,
		ch:    make(chan struct{}),
		stage: agiStageRequest,
	}
}

func newAgiAction(timeout time.Duration) *defaultAction {
	req := newActionRequest(actionAGI, timeout)

	return &defaultAction{
		req:     req,
		handler: newAgiActionHandler(req),
	}
}

func (a *ami) Answer(ctx context.Context, host string, channel string) error {
	ac := newAgiAction(a.cfg.ActionTimeout)
	ac.addField(keyChannel, channel)
	ac.addField(keyCommand, agiCmdAnswer)
	ac.addField(keyCommandID, ac.req.id)

	_, err := a.runAction(ctx, host, ac)

	return err
}

func (a *ami) Playback(ctx context.Context, host string, channel string, fileName string) error {
	ac := newAgiAction(a.cfg.ActionTimeout)
	ac.addField(keyChannel, channel)
	ac.addField(keyCommand, fmt.Sprintf("%s %s", agiCmdPlayback, fileName))
	ac.addField(keyCommandID, ac.req.id)

	_, err := a.runAction(ctx, host, ac)

	return err
}

func (a *ami) StartMOH(ctx context.Context, host string, channel string) error {
	ac := newAgiAction(a.cfg.ActionTimeout)
	ac.addField(keyChannel, channel)
	ac.addField(keyCommand, fmt.Sprintf("%s %s", agiCmdSetMusic, agiOn))
	ac.addField(keyCommandID, ac.req.id)

	_, err := a.runAction(ctx, host, ac)

	return err
}

func (a *ami) StopMOH(ctx context.Context, host string, channel string) error {
	ac := newAgiAction(a.cfg.ActionTimeout)
	ac.addField(keyChannel, channel)
	ac.addField(keyCommand, fmt.Sprintf("%s %s", agiCmdSetMusic, agiOff))
	ac.addField(keyCommandID, ac.req.id)

	_, err := a.runAction(ctx, host, ac)

	return err
}
