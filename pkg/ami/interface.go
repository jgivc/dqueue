package ami

import (
	"context"
	"net"
	"time"
)

type (
	Ami interface {
		Start(ctx context.Context) error
		Subscribe(filter Filter) Subscriber
		Close() error

		Answer(ctx context.Context, host string, channel string) error
		Playback(ctx context.Context, host string, channel string, fileName string) error
		StartMOH(ctx context.Context, host string, channel string) error
		StopMOH(ctx context.Context, host string, channel string) error
		Originate(host string, channel string) OriginateBuilder
		Hangup(ctx context.Context, host string, channel string, cause int) error
		Bridge(ctx context.Context, host string, channel1, channel2, tone string) error
		Setvar(ctx context.Context, host string, channel, variable, value string) error
	}

	Filter func(e *Event) bool

	Subscriber interface {
		Events() <-chan *Event
		Close()
	}

	OriginateBuilder interface {
		Exten(ext string) OriginateBuilder
		Context(ctx string) OriginateBuilder
		Priority(pri uint) OriginateBuilder
		Application(app string) OriginateBuilder
		Data(data string) OriginateBuilder
		Timeout(timeout time.Duration) OriginateBuilder
		CallerID(clid string) OriginateBuilder
		Variable(key string, value interface{}) OriginateBuilder
		AccountCode(code string) OriginateBuilder
		EarlyMedia(media bool) OriginateBuilder
		Async(async bool) OriginateBuilder
		Codecs(codecs ...string) OriginateBuilder
		ChannelID(id string) OriginateBuilder
		OtherChannelID(id string) OriginateBuilder
		Run(ctx context.Context) error
	}

	pubSubIf interface {
		Subscribe(f Filter) Subscriber
		Publish(e *Event)
		Close()
	}

	connectionFactory interface {
		Connect(ctx context.Context, addr string) (net.Conn, error)
	}
)
