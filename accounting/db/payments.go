package db

import (
	"database/sql"
	"fmt"
)

type Payments struct {
	Db *sql.DB
}

func (ps *Payments) Add(tx *sql.Tx, billingCycleId uint64, workerId uint64, amount uint) (Payment, error) {
	query := `INSERT INTO accounting_payments (amount, accounting_workers.id, accounting_billing_cycles.id) VALUES ($1, $2, $3) RETURNING id`
	var paymentId uint64
	err := tx.QueryRow(query, amount, workerId, billingCycleId).Scan(&paymentId)
	if err != nil {
		return Payment{}, fmt.Errorf(`failed to add payment worker=%d bid=%d amount=%d: %w`, workerId, billingCycleId, amount, err)
	}

	return Payment{Id: paymentId, WorkerId: workerId}, nil
}

type Payment struct {
	Id       uint64
	WorkerId uint64
}

func (p *Payment) Complete(tx *sql.Tx) error {
	query := `UPDATE accounting_payments SET status = "complete" WHERE id = $1`
	_, err := tx.Exec(query, p.Id)
	return err
}
