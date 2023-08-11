package adapter

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jgivc/dqueue/internal/entity"
	"github.com/jgivc/dqueue/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	errRepoError = errors.New("repo error")
)

type GetUniqueID interface {
	GetUniqueID() string
}

type ClientRepo struct {
	mux              sync.Mutex
	clients          map[string]*entity.Client
	logger           logger.Logger
	promClientsGauge prometheus.Gauge
}

func (r *ClientRepo) New(number string, data interface{}) (*entity.Client, error) {
	if number == "" {
		return nil, fmt.Errorf("number cannot be empty: %w", errRepoError)
	}

	obj, ok := data.(GetUniqueID)
	if !ok {
		return nil, fmt.Errorf("cannot convert data to GetUniqueID interface: %w", errRepoError)
	}

	id := obj.GetUniqueID()
	if id == "" {
		return nil, fmt.Errorf("GetUniqueID return value cannot be empty string: %w", errRepoError)
	}

	client := entity.NewClient(number, data)

	r.mux.Lock()
	defer r.mux.Unlock()

	if _, exists := r.clients[id]; exists {
		return nil, fmt.Errorf("client %s is already exists: %w", id, errRepoError)
	}

	r.logger.Info("msg", "New client", "number", client.Number, "id", client.ID)

	r.clients[id] = &client
	r.promClientsGauge.Inc()

	return &client, nil
}

func (r *ClientRepo) Remove(_ string, data interface{}) error {
	obj, ok := data.(GetUniqueID)
	if !ok {
		return fmt.Errorf("cannot convert data to GetUniqueID interface: %w", errRepoError)
	}

	id := obj.GetUniqueID()
	if id == "" {
		return fmt.Errorf("GetUniqueID return value cannot be empty string: %w", errRepoError)
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	if _, exists := r.clients[id]; !exists {
		// return fmt.Errorf("cannot find client: %s, %w", id, errRepoError)
		return nil
	}

	client := r.clients[id]
	r.logger.Info("msg", "Remove client", "number", client.Number, "id", client.ID)

	r.clients[id].Close()
	delete(r.clients, id)
	r.promClientsGauge.Dec()

	return nil
}

func (r *ClientRepo) Close() {
	r.mux.Lock()
	defer r.mux.Unlock()

	for id := range r.clients {
		r.clients[id].Close()
	}

	r.clients = nil
}

func NewClientRepo(promClientsGauge prometheus.Gauge, logger logger.Logger) *ClientRepo {
	return &ClientRepo{
		clients:          make(map[string]*entity.Client),
		logger:           logger,
		promClientsGauge: promClientsGauge,
	}
}
