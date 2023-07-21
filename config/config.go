package config

import "time"

type (
	Config struct {
		ClientService     `yaml:"client_service"`
		OperatorRepo      `yaml:"operator_repo"`
		PubSubConfig      `yaml:"pubsub"`
		VoipAdapterConfig `yaml:"voip_adapter"`
		AmiConfig         `yaml:"ami"`
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

	VoipAdapterConfig struct {
		DialTimeout       time.Duration `yaml:"dial_timeout"`
		DialContext       string        `yaml:"dial_context"`
		DialTemplate      string        `yaml:"dial_template"` // e.g. PJSIP/%s@context, %s - operator number
		DialExten         string        `yaml:"dial_to_exten"` // Originate second leg
		VarClientChannel  string        `yaml:"var_client_channel"`
		VarClientID       string        `yaml:"var_client_id"`
		VarOperatorNumber string        `yaml:"var_operator_number"`
	}

	AmiServerConfig struct {
		Host              string        `yaml:"host"`
		Port              int           `yaml:"port"`
		Username          string        `yaml:"username"`
		Secret            string        `yaml:"secret"`
		DialTimeout       time.Duration `yaml:"-"`
		ActionTimeout     time.Duration `yaml:"-"`
		ReconnectInterval time.Duration `yaml:"-"`
		ReaderBuffer      uint          `yaml:"-"`
	}

	AmiConfig struct {
		DialTimeout       time.Duration     `yaml:"dial_timeout"`
		ActionTimeout     time.Duration     `yaml:"action_timeout"`
		ReconnectInterval time.Duration     `yaml:"reconnect_interval"`
		ReaderBuffer      uint              `yaml:"reader_buffer"`
		PSConfig          PubSubConfig      `yaml:"pubsub"`
		Servers           []AmiServerConfig `yaml:"servers"`
	}
)

// TODO: Fill AmiServerConfig after load
// func (c *AmiServerConfig) FillValues(amiCfg *AmiConfig) {
// 	c.DialTimeout = amiCfg.DialTimeout
// 	c.ActionTimeout = amiCfg.ActionTimeout
// 	c.ReconnectInterval = amiCfg.ReconnectInterval
// 	c.ReaderBuffer = amiCfg.ReaderBuffer
// }

func New(fileName string) (*Config, error) {
	panic("not implemented")
}
