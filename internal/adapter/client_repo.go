package adapter

import (
	"fmt"
	"sync"

	"github.com/jgivc/vapp/internal/entity"
)

type ClientRepo struct {
	mux     sync.Mutex
	clients map[string]*entity.Client
}

func (r *ClientRepo) getID(host, uniqueID, channel string) string {
	return ""
}

func (r *ClientRepo) exists(id string) bool {
	_, exists := r.clients[id]

	return exists
}

func (r *ClientRepo) New(host, uniqueID, channel, number string) (*entity.Client, error) {
	id := r.getID(host, uniqueID, channel)

	r.mux.Lock()
	defer r.mux.Unlock()

	if r.exists(id) {
		return nil, fmt.Errorf("client with id: %s is already exists", id)
	}

	client := entity.NewClient(host, uniqueID, channel, number)
	r.clients[id] = &client

	return &client, nil
}

func (r *ClientRepo) Remove(host, uniqueID, channel string) error {
	id := r.getID(host, uniqueID, channel)

	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.exists(id) {
		return fmt.Errorf("client with id: %s does not exists", id)
	}

	r.clients[id].Close()
	delete(r.clients, id)

	return nil
}

func (r *ClientRepo) Close() error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for id := range r.clients {
		r.clients[id].Close()
	}

	r.clients = nil

	return nil
}
