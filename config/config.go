package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		ListenAddr        string            `yaml:"listen_addr" env-default:":8080"`
		MetricPath        string            `yaml:"metric_path" env-default:"/metric"`
		QueueConfig       QueueConfig       `yaml:"queue"`
		ClientService     ClientService     `yaml:"client_service"`
		OperatorRepo      OperatorRepo      `yaml:"operator_repo"`
		PubSubConfig      PubSubConfig      `yaml:"pubsub"`
		VoipAdapterConfig VoipAdapterConfig `yaml:"voip_adapter"`
		DialerConfig      DialerConfig      `yaml:"dialer"`
		AmiConfig         AmiConfig         `yaml:"ami"`
	}

	QueueConfig struct {
		MaxClients uint `yaml:"max_clients" env-default:"100"`
	}

	ClientService struct {
		ChannelBufferSize   int           `yaml:"channel_buffer_size" env-default:"100"`
		HandleClientTimeout time.Duration `yaml:"handle_timeout" env-default:"10s"`
	}

	OperatorRepo struct {
		APIURL     string        `yaml:"api_url"`
		APITimeout time.Duration `yaml:"api_timeout" env-default:"10s"`
		NoVerify   bool          `yaml:"no_verify"`
		Operators  []string      `yaml:"operators"` // If defined api will not be used
	}

	PubSubConfig struct {
		PublishQueueSize    uint `yaml:"publish_queue_size" env-default:"100"`
		SubscriberQueueSize uint `yaml:"subscriber_queue_size" env-default:"100"`
	}

	VoipAdapterConfig struct {
		DialTimeout time.Duration `yaml:"dial_timeout" env-default:"10s"`
		// DialContext string        `yaml:"dial_context" env-default:"default"`
		// DialExten   string        `yaml:"dial_exten" env-default:"s"`
		// e.g. PJSIP/%s@context, %s - operator number
		OriginateTechData string `yaml:"tech_data" env-default:"PJSIP/%s@default"`
		Application       string `yaml:"application"`
		Data              string `yaml:"data"`
		Context           string `yaml:"context"`
		Exten             string `yaml:"exten"`
		Priority          uint   `yaml:"priority"`
		// DialExten         string `yaml:"dial_to_exten"` // Originate second leg
		// VarClientChannel  string `yaml:"var_client_channel" env-default:"CLIENT_CHANNEL"`
		VarClientID string `yaml:"var_client_id" env-default:"CLIENT_ID"`
		// VarOperatorNumber string `yaml:"var_operator_number" env-default:"OPERATOR_NUMBER"`
	}

	AmiConfig struct {
		ActionTimeout     time.Duration     `yaml:"action_timeout" env-default:"10s"`
		ConnectTimeout    time.Duration     `yaml:"connect_timeout" env-default:"10s"`
		ReconnectInterval time.Duration     `yaml:"reconnect_interval" env-default:"30s"`
		ReaderBuffer      uint              `yaml:"reader_buffer" env-default:"100"`
		PSConfig          PubSubConfig      `yaml:"pubsub"`
		Servers           []AmiServerConfig `yaml:"servers"`
	}

	AmiServerConfig struct {
		Host              string        `yaml:"host"`
		Port              int           `yaml:"port" env-default:"5038"` //FIXME: env-default does not work
		Username          string        `yaml:"username"`
		Secret            string        `yaml:"secret"`
		ConnectTimeout    time.Duration `yaml:"-"`
		ActionTimeout     time.Duration `yaml:"-"`
		ReconnectInterval time.Duration `yaml:"-"`
		ReaderBuffer      uint          `yaml:"-"`
	}

	DialerConfig struct {
		CheckInterval             time.Duration `yaml:"check_interval" env-default:"30s"`
		DialToAllOperatorsTimeout time.Duration `yaml:"dial_to_all_operators_timeout" env-default:"30m"`
		DialPause                 time.Duration `yaml:"dial_pause" env-default:"20s"`
	}
)

func New(fileName string) (*Config, error) {
	var (
		cfg Config
		err error
	)

	if _, err2 := os.Stat(fileName); err2 == nil {
		err = cleanenv.ReadConfig(fileName, &cfg)
	} else {
		err = cleanenv.ReadEnv(&cfg)
	}

	if err != nil {
		return nil, err
	}

	if len(cfg.AmiConfig.Servers) < 1 {
		return nil, fmt.Errorf("you must assign at least one ami server")
	}

	for i := range cfg.AmiConfig.Servers {
		cfg.AmiConfig.Servers[i].ConnectTimeout = cfg.AmiConfig.ConnectTimeout
		cfg.AmiConfig.Servers[i].ActionTimeout = cfg.AmiConfig.ActionTimeout
		cfg.AmiConfig.Servers[i].ReconnectInterval = cfg.AmiConfig.ReconnectInterval
		cfg.AmiConfig.Servers[i].ReaderBuffer = cfg.AmiConfig.ReaderBuffer
		if cfg.AmiConfig.Servers[i].Port == 0 {
			cfg.AmiConfig.Servers[i].Port = 5038
		}
	}

	if cfg.VoipAdapterConfig.Application != "" {
		if cfg.VoipAdapterConfig.Context != "" || cfg.VoipAdapterConfig.Exten != "" || cfg.VoipAdapterConfig.Priority > 0 {
			return nil, fmt.Errorf("voip_adapter use application and data or context, exten, priority not both")
		}
	}

	if cfg.VoipAdapterConfig.Context != "" {
		if cfg.VoipAdapterConfig.Application != "" || cfg.VoipAdapterConfig.Data != "" {
			return nil, fmt.Errorf("voip_adapter use application and data or context, exten, priority not both")
		}

		if cfg.VoipAdapterConfig.Exten == "" {
			return nil, fmt.Errorf("voip_adapter if context is defined then exten must be set")
		}
	}

	if cfg.VoipAdapterConfig.Application == "" && cfg.VoipAdapterConfig.Context == "" {
		return nil, fmt.Errorf("voip_adapter at least one must be set: application [data] or context, exten, [priority]")
	}

	return &cfg, err
}
