package entity

import (
	"fmt"
	"sync/atomic"
)

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

func (o *Operator) String() string {
	return fmt.Sprintf("<Operator: %s>", o.Number)
}

func NewOperator(number, firstName, lastName string) *Operator {
	return &Operator{
		Number:    number,
		FirstName: firstName,
		LastName:  lastName,
	}
}
