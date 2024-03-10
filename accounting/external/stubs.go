package external

import (
	"log"
)

type PaymentSystem struct{}

func (ps PaymentSystem) Pay() error {
	log.Println(`payment system: pay()`)
	return nil
}
