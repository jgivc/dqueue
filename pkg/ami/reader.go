package ami

import (
	"bufio"
	"io"
	"strings"
	"sync/atomic"
)

const (
	keyEvent       = "Event"
	keyChannel     = "Channel"
	keyCallerIDNum = "CallerIDNum"
)

type Reader struct {
	r        *bufio.Reader
	shutdown atomic.Bool
}

func (er *Reader) Close() error {
	er.shutdown.Store(true)

	return nil
}

func (er *Reader) Read() (Event, error) {
	var e Event

	e.Data = make(map[string]string)

	for !er.shutdown.Load() {
		line, err := er.r.ReadString('\n')
		if err != nil {
			return e, err
		}

		line = strings.TrimSpace(line)

		if line == "" && len(e.Data) > 0 {
			break
		}

		data := strings.Split(line, ": ")
		if len(data) > 1 {
			e.Data[data[0]] = strings.TrimSpace(data[1])
		} else {
			continue
		}
	}

	e.Name = e.Get(keyEvent)
	e.Channel = e.Get(keyChannel)
	e.CallerIDNum = e.Get(keyCallerIDNum)

	return e, nil
}

func NewAmiReader(r io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
	}
}
