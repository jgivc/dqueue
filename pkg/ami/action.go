package ami

type Ami interface {
	Subscribe(filter Filter) Subscriber
}

// type ami struct {
// }
