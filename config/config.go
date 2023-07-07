package config

import "time"

type (
	Config struct {
		ClientService `yaml:"client_service"`
		OperatorRepo  `yaml:"operator_repo"`
		PubSubConfig  `yaml:"pubsub"`
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

	PubSubConfig struct {
		PublishQueueSize    uint `yaml:"publish_queue_size"`
		SubscriberQueueSize uint `yaml:"subscriber_queue_size"`
	}

	AmiServer struct {
		Host              string        `yaml:"host"`
		Port              int           `yaml:"port"`
		Username          string        `yaml:"username"`
		Password          string        `yaml:"password"`
		DialTimeout       time.Duration `yaml:"dial_timeout"`
		ActionTimeout     time.Duration `yaml:"action_timeout"`
		ReconnectInterval time.Duration `yaml:"reconnect_interval"`
		ReaderBuffer      uint
	}
)

func New(fileName string) (*Config, error) {
	panic("not implemented")
}
