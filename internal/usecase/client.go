package usecase

import (
	"context"
	"fmt"

	"github.com/jgivc/vapp/internal/entity"
	"github.com/jgivc/vapp/pkg/logger"
)

type ClientUseCase struct {
	clientRepo    clientRepo
	queueService  queueService
	dialerService dialerService
	voipService   voipService
	logger        logger.Logger
}

/*
If client hangup it may happen faster than client handle goroutine starts. So in handler channel synchronization may be needed.
*/
func (u *ClientUseCase) NewClient(ctx context.Context, host, uniqueID, channel, number string) error {
	client, err := u.clientRepo.New(host, uniqueID, channel, number)
	if err != nil {
		return err
	}

	if u.queueService.IsFull() {
		return fmt.Errorf("queue is full")
	}

	go u.handleClient(ctx, client)

	return nil
}

func (u *ClientUseCase) handleClient(ctx context.Context, client *entity.Client) {
out:
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return
		case <-client.Hangup():
			return
		default:
			switch i {
			case 0:
				if err := u.voipService.Answer(ctx, client); err != nil {
					u.logger.Error("msg", "Cannot answer to client", "client", client)
					break out
				}
			case 1:
				if err := u.voipService.StartMOH(ctx, client); err != nil {
					u.logger.Error("msg", "Cannot start MOH", "client", client)
					break out
				}
			case 2:
				if err := u.queueService.Push(client); err != nil {
					u.logger.Error("msg", "Cannot push client to queue", "client", client)
					break out
				}

				u.dialerService.Notify()
			}
		}
	}

}

func (u *ClientUseCase) Hangup(host, uniqueID, channel string) error {
	return u.clientRepo.Remove(host, uniqueID, channel)
}

func NewClientUseCase(clientRepo clientRepo, queueService queueService, dialerService dialerService, voipService voipService, logger logger.Logger) *ClientUseCase {
	return &ClientUseCase{
		clientRepo:    clientRepo,
		queueService:  queueService,
		dialerService: dialerService,
		voipService:   voipService,
	}
}
