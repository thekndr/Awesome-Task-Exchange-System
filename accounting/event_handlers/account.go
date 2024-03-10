package event_handlers

import (
	"fmt"
	"github.com/thekndr/ates/accounting/db"
	"log"
)

type (
	WorkerAccountCreatedEvent struct {
		Email string `json:"email"`
		Id    string `json:"id"`
	}

	AccountCreated struct {
		Workers db.Workers
	}
)

func (h *AccountCreated) Handle(ev WorkerAccountCreatedEvent) error {
	internalId, err := h.Workers.Add(ev.Id, ev.Email)
	if err != nil {
		log.Printf(`account-created: failed to add new account: %s`, err)
		return err
	}

	log.Printf(`account-created: user=%s:%s added with id=%d`, ev.Email, ev.Id, internalId)
	return nil
}

type (
	AccountRoleChangedEvent struct {
		Email   string `json:"email"`
		Id      string `json:"id"`
		NewRole string `json:"new_role"`
	}

	AccountRoleChanged struct{}
)

func (h *AccountRoleChanged) Handle(interface{}) error {
	return fmt.Errorf(`TODO: not supported yet`)
}
