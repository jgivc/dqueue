package config

import "time"

type (
	Config struct {
		ClientService `yaml:"client_service"`
		OperatorRepo  `yaml:"operator_repo"`
	}

	ClientService struct {
		ChannelBufferSize   int
		HandleClientTimeout time.Duration
	}

	OperatorRepo struct {
		APIURL     string        `yaml:"api_url"`
		APITimeout time.Duration `yaml:"api_timeout"`
		NoVerify   bool          `yaml:"no_verify"`
	}
)

func New(fileName string) (*Config, error) {
	panic("not implemented")
}
