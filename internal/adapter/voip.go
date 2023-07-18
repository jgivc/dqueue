package adapter

import (
	"context"

	"github.com/jgivc/vapp/internal/entity"
	"github.com/jgivc/vapp/pkg/ami"
)

type VoipAdapter struct {
	ami ami.Ami
}

func (v *VoipAdapter) Answer(ctx context.Context, client *entity.Client) error {
	panic("not implemented")
}

func (v *VoipAdapter) Playback(ctx context.Context, client *entity.Client, fileName string) error {
	panic("not implemented")
}

func (v *VoipAdapter) StartMOH(ctx context.Context, client *entity.Client) error {
	panic("not implemented")
}

func (v *VoipAdapter) StopMOH(ctx context.Context, client *entity.Client) error {
	panic("not implemented")
}

func (v *VoipAdapter) Dial(ctx context.Context, client *entity.Client, operators ...entity.Operator) error {
	panic("not implemented")
}

func (v *VoipAdapter) Hangup(ctx context.Context, client *entity.Client) error {
	panic("not implemented")
}

func NewVoipAdapter(a ami.Ami) *VoipAdapter {
	return &VoipAdapter{
		ami: a,
	}
}
