package entity

import (
	"github.com/google/uuid"
)

type Client struct {
	ID     string
	Number string
	Data   interface{}
	lost   chan struct{}
}

func (c *Client) String() string {
	return c.ID
}

func (c *Client) Close() {
	//TODO: synchronize
	select {
	case _, ok := <-c.lost:
		if !ok {
			// Already closed
			return
		}
	default:
	}

	close(c.lost)
}

func (c *Client) Lost() <-chan struct{} {
	return c.lost
}

func (c *Client) IsAlive() bool {
	select {
	case _, ok := <-c.lost:
		if !ok {
			return false
		}
	default:
	}

	return true
}

func NewClient(number string, data interface{}) Client {
	return Client{
		ID:     uuid.New().String(),
		Number: number,
		Data:   data,
		lost:   make(chan struct{}),
	}
}
