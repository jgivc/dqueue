package ami

import (
	"bytes"
	"fmt"
	"io"
)

const (
	emptyString = ""
)

type Event struct {
	Host        string
	Name        string
	Channel     string
	CallerIDNum string
	Data        map[string]string
}

func (e *Event) write(w io.Writer) error {
	b := bytes.NewBuffer(nil)
	b.WriteString(fmt.Sprintf("%s: %s\r\n", e.Name, e.Data[e.Name]))
	for key, val := range e.Data {
		if key == e.Name {
			continue
		}

		b.WriteString(fmt.Sprintf("%s: %s\r\n", key, val))
	}
	b.WriteString("\r\n")

	n, err := w.Write(b.Bytes())
	if n != b.Len() {
		return fmt.Errorf("not all data writen: expected: %d, actual: %d", b.Len(), n)
	}

	return err
}

func (e *Event) Get(key string) string {
	if _, exists := e.Data[key]; exists {
		return e.Data[key]
	}

	return emptyString
}

func (e *Event) Copy() *Event {
	var eCopy Event
	eCopy.Data = make(map[string]string)

	eCopy.Host = e.Host
	eCopy.Name = e.Name
	eCopy.Channel = e.Channel
	eCopy.CallerIDNum = e.CallerIDNum

	for k := range e.Data {
		eCopy.Data[k] = e.Data[k]
	}

	return &eCopy
}
