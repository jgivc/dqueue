package entity

import "sync/atomic"

type Operator struct {
	Number    string
	LastName  string
	FirstName string
	busy      atomic.Bool
}

func (o *Operator) SetBusy(val bool) {
	o.busy.Store(val)
}

func (o *Operator) IsBusy() bool {
	return o.busy.Load()
}
