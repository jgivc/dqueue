package ami

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
