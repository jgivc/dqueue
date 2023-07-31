package adapter

import "fmt"

type ClientDto struct {
	Host    string
	Channel string
}

func (d *ClientDto) GetUniqueID() string {
	return fmt.Sprintf("%s@%s", d.Channel, d.Host)
}
