package db

import (
	"database/sql"
	"fmt"
)

type Transactions struct {
	Db *sql.DB
}

func (ts *Transactions) Get(billingCycleId uint64, taskId, userId uint64) (Transaction, error) {
	query := `SELECT id, withdrawal, enrolment FROM accounting_transactions WHERE id = $1 AND id = $2 AND id = $3`

	t := Transaction{
		BillingCycleId: billingCycleId, TaskId: taskId, WorkerId: userId,
	}

	err := ts.Db.QueryRow(query, billingCycleId, taskId, userId).Scan(
		&t.Id, &t.Withdrawal, &t.Enrolment,
	)
	if err != nil {
		return Transaction{}, err
	}

	return t, nil

}

func (ts *Transactions) New(billingCycleId uint64, taskId, userId uint64) (Transaction, error) {
	var id uint64
	query := `INSERT INTO accounting_transactions (accounting_billing_cycles.id, accounting_tasks.id, accounting_workers.id) VALUES ($1, $2, $3) RETURNING id`
	err := ts.Db.QueryRow(query, billingCycleId, taskId, userId).Scan(&id)
	if err != nil {
		return Transaction{}, err
	}

	t := Transaction{
		Id: id, BillingCycleId: billingCycleId,
		TaskId: taskId, WorkerId: userId,
	}

	return t, nil
}

type Transaction struct {
	Id             uint64
	BillingCycleId uint64
	TaskId         uint64
	WorkerId       uint64

	Withdrawal uint
	Enrolment  uint
}

func (t *Transaction) Withdraw(tx *sql.Tx, value uint) error {
	query := `UPDATE accounting_transactions SET withdrawal = $2 WHERE id = $1`
	_, err := tx.Exec(query, t.Id, value)
	if err != nil {
		fmt.Errorf(`failed to withdraw tx id=%d value=%d: %w`, t.Id, value, err)
	}
	return err
}

func (t *Transaction) Enroll(tx *sql.Tx, value uint) error {
	query := `UPDATE accounting_transactions SET enrolment = $2 WHERE id = $1`
	_, err := tx.Exec(query, t.Id, value)
	if err != nil {
		fmt.Errorf(`failed to enroll tx id=%d value=%d: %w`, t.Id, value, err)
	}
	return err
}
