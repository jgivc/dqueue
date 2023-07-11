package ami

import (
	"fmt"
	"sync"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/pkg/logger"
)

// type PubSub interface {
// 	Subscribe(f Filter) Subscriber
// 	Publish(e *Event)
// 	Close()
// }

type subscriber struct {
	ch          chan *Event
	stop        chan struct{}
	filter      Filter
	unsubscribe chan *subscriber
}

func (s *subscriber) Events() <-chan *Event {
	return s.ch
}

func (s *subscriber) Close() {
	select {
	case _, ok := <-s.stop:
		if !ok {
			return
		}
	default:
	}

	defer close(s.ch)
	s.unsubscribe <- s
}

type pubSub struct {
	ch            chan *Event
	wg            sync.WaitGroup
	stop          chan struct{}
	subscribe     chan *subscriber
	unsubscribe   chan *subscriber
	subscribers   map[*subscriber]struct{}
	subsQueueSize uint
}

func (ps *pubSub) Subscribe(f Filter) Subscriber {
	subs := &subscriber{
		ch:          make(chan *Event, ps.subsQueueSize),
		stop:        ps.stop,
		filter:      f,
		unsubscribe: ps.unsubscribe,
	}

	ps.subscribe <- subs

	return subs
}

func (ps *pubSub) Publish(e *Event) {
	ps.ch <- e
}

func (ps *pubSub) Close() {
	close(ps.stop)
	close(ps.subscribe)
	close(ps.unsubscribe)
	ps.wg.Wait()

	for s := range ps.subscribers {
		select {
		case _, ok := <-s.ch:
			if !ok {
				continue
			}
		default:
			close(s.ch)
		}
	}

	ps.subscribers = nil

	close(ps.ch)
}

func newPubSub(cfg *config.PubSubConfig, logger logger.Logger) *pubSub {
	ps := &pubSub{
		ch:            make(chan *Event, cfg.PublishQueueSize),
		stop:          make(chan struct{}),
		subscribe:     make(chan *subscriber),
		unsubscribe:   make(chan *subscriber),
		subscribers:   make(map[*subscriber]struct{}),
		subsQueueSize: cfg.SubscriberQueueSize,
	}

	start := make(chan struct{})
	defer close(start)

	ps.wg.Add(1)
	go func() {
		defer func() {
			ps.wg.Done()
			logger.Info("msg", "PubSub done")
		}()

		logger.Info("msg", "PubSub start")
		start <- struct{}{}

		for {
			select {
			case <-ps.stop:
				return
			case s, ok := <-ps.subscribe:
				if ok {
					logger.Info("msg", "Subscribe", "p", fmt.Sprintf("%p", s))
					ps.subscribers[s] = struct{}{}
				}
			case s, ok := <-ps.unsubscribe:
				if ok {
					logger.Info("msg", "Unsubscribe", "p", fmt.Sprintf("%p", s))
					delete(ps.subscribers, s)
				}
			case e := <-ps.ch:
				for sub := range ps.subscribers {
					if sub.filter(e) {
						select {
						case sub.ch <- e.Copy():
							// logger.Info("msg", "Send", "p", fmt.Sprintf("%p", sub))
						default:
							logger.Warn("msg", "Cannot send event to subscriber")
						}
					}
				}
			}
		}
	}()

	<-start

	return ps
}
