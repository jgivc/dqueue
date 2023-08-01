package adapter

import (
	"container/list"
	"errors"
	"fmt"
	"sync"

	"github.com/jgivc/vapp/internal/entity"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	errQueue = errors.New("queue error")
)

// TODO: Remove closed clients
// TODO: If working time is over
type Queue struct {
	maxClients int
	mux        sync.Mutex
	clients    *list.List
	promLength prometheus.Gauge
}

func (q *Queue) isFull() bool {
	return !(q.clients.Len() < q.maxClients)
}

func (q *Queue) IsFull() bool {
	q.mux.Lock()
	defer q.mux.Unlock()

	return q.isFull()
}

func (q *Queue) hasClients() bool {
	return q.clients.Len() > 0
}

func (q *Queue) HasClients() bool {
	q.mux.Lock()
	defer q.mux.Unlock()

	return q.hasClients()
}

func (q *Queue) Push(client *entity.Client) error {
	q.mux.Lock()
	defer q.mux.Unlock()

	if client == nil {
		return fmt.Errorf("client cannot be nil: %w", errQueue)
	}

	if q.isFull() {
		return fmt.Errorf("cannot insert client, queue is full: %w", errQueue)
	}

	q.clients.PushBack(client)
	q.promLength.Inc()

	return nil
}

func (q *Queue) Pop() (*entity.Client, error) {
	q.mux.Lock()
	defer q.mux.Unlock()

	if !q.hasClients() {
		return nil, fmt.Errorf("cannot pop client, queue is empty: %w", errQueue)
	}

	el := q.clients.Remove(q.clients.Front())
	if el == nil {
		return nil, fmt.Errorf("cannot get client from list: %w", errQueue)
	}

	q.promLength.Dec()

	return el.(*entity.Client), nil
}

func (q *Queue) Close() {
	q.mux.Lock()
	defer q.mux.Unlock()

	for e := q.clients.Front(); e != nil; q.clients.Remove(e) {
		e = q.clients.Front()

		if e != nil {
			e.Value = nil
		} else {
			break
		}
	}
}

func NewQueue(maxClients int) *Queue {
	return &Queue{
		maxClients: maxClients,
		clients:    list.New(),
		promLength: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_queue_len",
			Help: "Queue length gauge",
		}),
	}
}
