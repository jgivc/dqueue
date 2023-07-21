package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/entity"
	"github.com/jgivc/vapp/pkg/ami"
)

const (
	hangupCause = 11
)

var (
	errCannotConvert = errors.New("cannot convert to ClientDto")
)

type VoipAdapter struct {
	cfg *config.VoipAdapterConfig
	ami ami.Ami
}

func (v *VoipAdapter) Answer(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Answer(ctx, dto.Host, dto.Channel)
}

func (v *VoipAdapter) Playback(ctx context.Context, client *entity.Client, fileName string) error {
	dto, ok := client.Data.(ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Playback(ctx, dto.Host, dto.Channel, fileName)
}

func (v *VoipAdapter) StartMOH(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.StartMOH(ctx, dto.Host, dto.Channel)
}

func (v *VoipAdapter) StopMOH(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.StopMOH(ctx, dto.Host, dto.Channel)
}

func (v *VoipAdapter) Hangup(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Hangup(ctx, dto.Host, dto.Channel, hangupCause)
}

func (v *VoipAdapter) Dial(ctx context.Context, client *entity.Client, operator *entity.Operator) error {
	dto, ok := client.Data.(ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Originate(dto.Host, fmt.Sprintf(v.cfg.DialTemplate, operator.Number)).
		Context(v.cfg.DialContext).
		CallerID(client.Number).
		Variable(v.cfg.VarClientChannel, dto.Channel).
		Variable(v.cfg.VarClientID, client.ID).
		Variable(v.cfg.VarOperatorNumber, operator.Number).
		Async(true).
		Run(ctx)
}

func NewVoipAdapter(a ami.Ami) *VoipAdapter {
	return &VoipAdapter{
		ami: a,
	}
}
