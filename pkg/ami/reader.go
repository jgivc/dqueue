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

type amiReader struct {
	r        *bufio.Reader
	shutdown atomic.Bool
}

func (er *amiReader) Close() error {
	er.shutdown.Store(true)

	return nil
}

func (er *amiReader) Read() (Event, error) {
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

func newAmiReader(r io.Reader) *amiReader {
	return &amiReader{
		r: bufio.NewReader(r),
	}
}
