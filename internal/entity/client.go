package entity

import "fmt"

type Client struct {
	Host     string
	UniqueID string
	Channel  string
	Number   string
	ch       chan struct{}
}

func (c *Client) String() string {
	return fmt.Sprintf("%s@%s", c.Channel, c.Host)
}

func (c *Client) Close() {
	//It synchronized by QueueService mutex
	select {
	case _, ok := <-c.ch:
		if !ok {
			// Already closed
			return
		}
	default:
	}

	close(c.ch)
}

func (c *Client) Hangup() <-chan struct{} {
	return c.ch
}

func (c *Client) IsAlive() bool {
	select {
	case _, ok := <-c.ch:
		if !ok {
			return false
		}
	default:
	}

	return true
}

func NewClient(host, uniqueID, channel, number string) Client {
	return Client{
		Host:     host,
		UniqueID: uniqueID,
		Channel:  channel,
		Number:   number,
		ch:       make(chan struct{}),
	}
}
