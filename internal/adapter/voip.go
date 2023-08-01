package adapter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/entity"
	"github.com/jgivc/vapp/pkg/ami"
)

const (
	hangupCause       = 11
	chBufSize         = 1
	noTone            = "no"
	varCallerIDNumber = "CALLERID(num)"
)

var (
	errCannotConvert = errors.New("cannot convert to ClientDto")
)

type VoipAdapter struct {
	cfg   *config.VoipAdapterConfig
	ami   ami.Ami
	store *idsStore
}

func (v *VoipAdapter) Answer(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(*ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Answer(ctx, dto.Host, dto.Channel)
}

func (v *VoipAdapter) Playback(ctx context.Context, client *entity.Client, fileName string) error {
	dto, ok := client.Data.(*ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Playback(ctx, dto.Host, dto.Channel, fileName)
}

func (v *VoipAdapter) StartMOH(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(*ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.StartMOH(ctx, dto.Host, dto.Channel)
}

func (v *VoipAdapter) StopMOH(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(*ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.StopMOH(ctx, dto.Host, dto.Channel)
}

func (v *VoipAdapter) Hangup(ctx context.Context, client *entity.Client) error {
	dto, ok := client.Data.(*ClientDto)
	if !ok {
		return errCannotConvert
	}

	return v.ami.Hangup(ctx, dto.Host, dto.Channel, hangupCause)
}

func (v *VoipAdapter) Operator(id, channel string) error {
	return v.store.set(id, channel)
}

func (v *VoipAdapter) Dial(ctx context.Context, client *entity.Client, operator *entity.Operator) error {
	dto, ok := client.Data.(*ClientDto)
	if !ok {
		return errCannotConvert
	}

	ch, err := v.store.add(client.ID)
	if err != nil {
		return fmt.Errorf("cannot store id, %w", err)
	}
	defer v.store.remove(client.ID)

	ctx2, cancel := context.WithTimeout(ctx, v.cfg.DialTimeout)
	defer cancel()

	// // TODO: Watch for BridgeEnter client and operator then return.
	// sub := v.ami.Subscribe(func(e *ami.Event) bool {
	// 	if e.Host != dto.Host {
	// 		return false
	// 	}
	// 	if e.IsEvent() && e.Event() == events.BridgeEnter {
	// 		// fmt.Println(e)
	// 		if e.Get(fields.CallerIDNum) == client.Number {
	// 			return e.Get(fields.ConnectedLineNum) == operator.Number
	// 		}

	// 		if e.Get(fields.ConnectedLineNum) == client.Number {
	// 			return e.Get(fields.CallerIDNum) == operator.Number
	// 		}
	// 	}

	// 	return false
	// })
	// defer sub.Close()

	err2 := v.ami.Originate(dto.Host, fmt.Sprintf(v.cfg.DialTemplate, operator.Number)).
		Application("agi").
		Data(fmt.Sprintf("agi:async,%s", client.ID)).
		// ChannelID(client.ID).
		// Context(v.cfg.DialContext).
		// Exten(v.cfg.DialExten).
		CallerID(client.Number).
		Timeout(v.cfg.DialTimeout).
		// Variable(v.cfg.VarClientChannel, dto.Channel).
		// Variable(v.cfg.VarClientID, client.ID).
		// Variable(v.cfg.VarOperatorNumber, operator.Number).
		Async(true).
		Run(ctx2)
	if err2 != nil {
		return err
	}

	select {
	case <-ctx2.Done():
		return ctx2.Err()
	case channel := <-ch:
		if errClid := v.ami.Setvar(ctx2, dto.Host, *channel, varCallerIDNumber, operator.Number); errClid != nil {
			return fmt.Errorf("cannot set callerIDNum to operator channel: %w", errClid)
		}

		err3 := v.ami.Bridge(ctx2, dto.Host, dto.Channel, *channel, noTone)
		return err3
	}
}

func (v *VoipAdapter) Close() {
	v.store.close()
}

func NewVoipAdapter(cfg *config.VoipAdapterConfig, a ami.Ami) *VoipAdapter {
	return &VoipAdapter{
		ami:   a,
		cfg:   cfg,
		store: newIDsStore(),
	}
}

type idsStore struct {
	mux  sync.Mutex
	data map[string]chan *string
}

func (s *idsStore) add(id string) (<-chan *string, error) {
	var (
		ch  chan *string
		err error
	)
	s.mux.Lock()

	if _, exists := s.data[id]; exists {
		err = fmt.Errorf("key %s is already exists", id)
	} else {
		ch = make(chan *string, chBufSize)
		s.data[id] = ch
	}
	s.mux.Unlock()

	return ch, err
}

func (s *idsStore) set(id, channel string) error {
	var err error
	s.mux.Lock()

	if _, exists := s.data[id]; exists {
		select {
		case s.data[id] <- &channel:
		default:
			err = fmt.Errorf("cannot send channel id")
		}
	}

	s.mux.Unlock()

	return err
}

func (s *idsStore) remove(id string) {
	s.mux.Lock()
	if _, exists := s.data[id]; exists {
		close(s.data[id])
		delete(s.data, id)
	}
	s.mux.Unlock()
}

func (s *idsStore) close() {
	s.mux.Lock()
	for k := range s.data {
		close(s.data[k])
		delete(s.data, k)
	}
	s.mux.Unlock()
}

func newIDsStore() *idsStore {
	return &idsStore{
		data: make(map[string]chan *string),
	}
}
