package service

import (
	"context"

	"github.com/jgivc/vapp/internal/entity"
)

type VoipService struct {
}

func (s *VoipService) Answer(ctx context.Context, client *entity.Client) error {
	return nil
}

func (s *VoipService) Playback(ctx context.Context, client *entity.Client) error {
	return nil
}

func (s *VoipService) StartMOH(ctx context.Context, client *entity.Client) error {
	return nil
}

func (s *VoipService) StopMOH(ctx context.Context, client *entity.Client) error {
	return nil
}

// func (s *VoipService) Redirect(ctx context.Context, client *entity.Client, voipContext, exten string, priority int) error {
// 	return nil
// }

// func (s *VoipService) Dial(ctx context.Context, client *entity.Client, ext string) error {
// 	return nil
// }

// func (s *VoipService) Originate(ctx context.Context, host, channel, ext string) error {
// 	return nil
// }
