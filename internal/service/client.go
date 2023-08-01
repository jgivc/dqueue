package service

import (
	"context"
	"errors"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/entity"
)

const (
	NewClient = iota
	Hangup
)

var (
	errCannotHandle = errors.New("cannot handle")
)

type clientDto struct {
	reqType int
	number  string
	data    interface{}
}

type ClientService struct {
	voip                VoipAdapter
	queue               Queue
	repo                ClientRepo
	dialer              Dialer
	logger              Logger
	handleClientTimeout time.Duration
	ch                  chan *clientDto
}

func (s *ClientService) NewClient(number string, data interface{}) error {
	select {
	case s.ch <- &clientDto{
		reqType: NewClient,
		number:  number,
		data:    data,
	}:
	default:
		return errCannotHandle
	}

	return nil
}

func (s *ClientService) Hangup(number string, data interface{}) error {
	select {
	case s.ch <- &clientDto{
		reqType: Hangup,
		number:  number,
		data:    data,
	}:
	default:
		return errCannotHandle
	}

	return nil
}

func (s *ClientService) Start(ctx context.Context) {
	started := make(chan struct{})
	defer close(started)

	go func() {
		started <- struct{}{}

		for {
			select {
			case <-ctx.Done():
				return
			case d := <-s.ch:
				switch d.reqType {
				case NewClient:
					client, err := s.repo.New(d.number, d.data)
					if err != nil {
						s.logger.Error("msg", "Cannot create client", "error", err)
						continue
					}
					go s.handleClient(ctx, client)
				case Hangup:
					if err := s.repo.Remove(d.number, d.data); err != nil {
						s.logger.Error("msg", "Cannot remove client", "error", err)
					}
				}
			}
		}
	}()

	<-started
}

func (s *ClientService) handleClient(ctx context.Context, client *entity.Client) {
	clientCtx, cancel := context.WithTimeout(ctx, s.handleClientTimeout)
	defer cancel()

	go func() {
		select {
		case <-clientCtx.Done():
			return
		case <-client.Lost():
			cancel()
		}
	}()

	select {
	case <-ctx.Done():
		return
	case <-client.Lost():
		return
	default:
		if err := s.voip.Answer(clientCtx, client); err != nil {
			s.logger.Error("msg", "Cannot answer to client", "client", client, "error", err)

			return
		}

		if err := s.voip.StartMOH(clientCtx, client); err != nil {
			s.logger.Error("msg", "Cannot start MOH", "client", client, "error", err)

			return
		}
	}

	if err := s.queue.Push(client); err != nil {
		s.logger.Error("msg", "Cannot push client to queue", "client", client, "error", err)
		return
	}

	s.dialer.Notify()
}

func (s *ClientService) Operator(id, channel string) error {
	return s.voip.Operator(id, channel)
}

func NewClientService(cfg *config.ClientService, voip VoipAdapter, queue Queue, repo ClientRepo,
	dialer Dialer, logger Logger) *ClientService {
	return &ClientService{
		voip:                voip,
		queue:               queue,
		repo:                repo,
		dialer:              dialer,
		logger:              logger,
		handleClientTimeout: cfg.HandleClientTimeout,
		ch:                  make(chan *clientDto, cfg.ChannelBufferSize),
	}
}
