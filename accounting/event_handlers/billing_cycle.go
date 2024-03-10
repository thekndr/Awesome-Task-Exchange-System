package event_handlers

import (
	"database/sql"
	"fmt"
	db "github.com/thekndr/ates/accounting/db"
	"log"
)

type (
	BillingCycleCompletedEvent struct{}

	BillingCycleCompleted struct {
		BillingCycles db.BillingCycles
		Workers       db.Workers
	}
)

func (h *BillingCycleCompleted) Handle(_ BillingCycleCompletedEvent) error {
	current, err := h.BillingCycles.CurrentOrNew(nil)
	if err != nil {
		return fmt.Errorf(`bc-completed: failed to get current or create a new one: %w`, err)
	}

	tx, err := h.BillingCycles.Tx()
	if err != nil {
		return fmt.Errorf(`bc-completed: failed to start db-tx: %w`, err)
	}

	var rollbackTx = true
	defer func() {
		if rollbackTx {
			log.Printf(`bc-completed: rolling back db-tx...`)
			log.Printf(`bc-completed: rollback=%w`, tx.Rollback())
		}
	}()

	if err = current.Complete(tx); err != nil {
		return fmt.Errorf(`bc-completed: failed to complete current one=%d: %w`, current.Id, err)
	}

	if current, err = h.BillingCycles.CurrentOrNew(tx); err != nil {
		return fmt.Errorf(`bc-completed: failed to create a new one: %w`, err)
	}

	if err = h.updateBalances(tx, current); err != nil {
		return fmt.Errorf(`bc-completed: failed to update balances: %w`, err)
	}

	rollbackTx = false
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(`bc-complete: failed to commit tx: %w`, err)
	}

	return nil
}

func (h *BillingCycleCompleted) updateBalances(tx *sql.Tx, bc db.BillingCycle) error {
	transactions, err := bc.Transactions()
	if err != nil {
		return err
	}

	for _, accTx := range transactions {
		workerId := accTx.WorkerId
		var (
			worker db.Worker
			err    error
		)
		if worker, err = h.Workers.Get(accTx.WorkerId); err != nil {
			return err
		}

		log.Printf(`worker id=%d withdrawal=%d`, workerId, accTx.Withdrawal)
		if err := worker.Withdraw(tx, bc.Id, accTx.Withdrawal); err != nil {
			return err
		}

		log.Printf(`worker id=%d enrolment=%d`, workerId, accTx.Enrolment)
		if err := worker.Enroll(tx, bc.Id, accTx.Enrolment); err != nil {
			return err
		}
	}

	return nil
}
