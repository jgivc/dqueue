package ami

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	errActionError = fmt.Errorf("action error")
)

type action interface {
	addField(name string, value interface{})
	run(ctx context.Context, host string, ps pubSubIf, w io.Writer) (interface{}, error)
}

type actionRequest struct {
	id      string
	name    string
	timeout time.Duration
	buf     bytes.Buffer
}

func (req *actionRequest) addField(name string, value interface{}) {
	req.buf.WriteString(fmt.Sprintf("%s: %v\r\n", name, value))
}

func (req *actionRequest) writeTo(w io.Writer) (int64, error) {
	return req.buf.WriteTo(w)
}

func newActionRequest(name string, timeout time.Duration) *actionRequest {
	req := &actionRequest{
		id:      uuid.NewString(),
		name:    name,
		timeout: timeout,
	}

	req.addField(keyAction, name)
	req.addField(keyActionID, req.id)

	return req
}

type actionHandler interface {
	filter(e *Event) bool
	handle(e *Event)
	done() <-chan struct{}
	result() (interface{}, error)
}

type defaultActionHandler struct {
	req *actionRequest
	ch  chan struct{}
	obj interface{}
	err error
}

func (h *defaultActionHandler) filter(e *Event) bool {
	return e.Name == keyResponse && e.Get(keyActionID) == h.req.id
}

func (h *defaultActionHandler) channelClosed() bool {
	select {
	case _, ok := <-h.ch:
		if !ok {
			return true
		}
	default:
	}

	return false
}

func (h *defaultActionHandler) handle(e *Event) {
	if !(e.Get(keyResponse) == success || e.Get(keyResponse) == goodbye) {
		h.err = errActionError
	}

	if !h.channelClosed() {
		close(h.ch)
	}
}

func (h *defaultActionHandler) done() <-chan struct{} {
	return h.ch
}

func (h *defaultActionHandler) result() (interface{}, error) {
	if !h.channelClosed() {
		return nil, fmt.Errorf("handler is not done")
	}
	return h.obj, h.err
}

func newDefaultHandler(req *actionRequest) *defaultActionHandler {
	return &defaultActionHandler{
		req: req,
		ch:  make(chan struct{}),
	}
}

type defaultAction struct {
	req     *actionRequest
	handler actionHandler
}

func (da *defaultAction) addField(name string, value interface{}) {
	da.req.addField(name, value)
}

func (da *defaultAction) run(ctx context.Context, host string, ps pubSubIf, w io.Writer) (interface{}, error) {
	ctx2, cancel := context.WithTimeout(ctx, da.req.timeout)
	defer cancel()

	sub := ps.Subscribe(func(e *Event) bool {
		if e.Host == host {
			return da.handler.filter(e)
		}

		return false
	})
	defer sub.Close()

	go func() {
		for e := range sub.Events() {
			da.handler.handle(e)
		}
	}()

	if _, err := da.req.writeTo(w); err != nil {
		return nil, fmt.Errorf("cannot write action request: %w", err)
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return nil, fmt.Errorf("cannot write action request: %w", err)
	}

	select {
	case <-ctx2.Done():
		return nil, fmt.Errorf("action timeout: %w", errActionError)
	case <-da.handler.done():
		return da.handler.result()
	}
}

func newDefaultAction(name string, timeout time.Duration) *defaultAction {
	req := newActionRequest(name, timeout)

	return &defaultAction{
		req:     req,
		handler: newDefaultHandler(req),
	}
}

func (a *ami) runAction(ctx context.Context, host string, ac action) (interface{}, error) {
	srv, err := a.getServer(host)
	if err != nil {
		return nil, fmt.Errorf("cannot get host: %w", err)
	}

	return ac.run(ctx, host, a.ps, srv)
}

func (a *ami) Hangup(ctx context.Context, host string, channel string, cause int) error {
	ac := newDefaultAction(actionHangup, a.actionTimeout)
	ac.addField(keyChannel, channel)
	ac.addField(keyCause, cause)

	_, err := a.runAction(ctx, host, ac)

	return err
}
