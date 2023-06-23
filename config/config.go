package config

import "time"

type (
	Config struct {
		ClientService `yaml:"client_service"`
	}

	ClientService struct {
		ChannelBufferSize   int
		HandleClientTimeout time.Duration
	}
)

func New(fileName string) (*Config, error) {
	panic("not implemented")
}
