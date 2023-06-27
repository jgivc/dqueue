package entity

import "sync/atomic"

type Operator struct {
	Number    string      `json:"number"`
	FirstName string      `json:"last_name"`
	LastName  string      `json:"first_name"`
	busy      atomic.Bool `json:"-"`
}

func (o *Operator) SetBusy(val bool) {
	o.busy.Store(val)
}

func (o *Operator) IsBusy() bool {
	return o.busy.Load()
}

func NewOperator(number, firstName, lastName string) *Operator {
	return &Operator{
		Number:    number,
		FirstName: firstName,
		LastName:  lastName,
	}
}
