package db

import (
	"database/sql"
	"fmt"
)

const (
	billingCyclesTableName = `accounting_billing_cycles`
)

type BillingCycles struct {
	Db *sql.DB
}

func (bs *BillingCycles) Tx() (*sql.Tx, error) {
	tx, err := bs.Db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = bs.Db.Exec(fmt.Sprintf(`LOCK TABLE %s IN ACCESS EXCLUSIVE MODE`, billingCyclesTableName))
	return tx, err
}

func (bs *BillingCycles) Current(tx *sql.Tx) (BillingCycle, error) {
	query := `SELECT id FROM accounting_billing_cycles WHERE status = 'active'`
	rows, err := tx.Query(query)
	if err != nil {
		return BillingCycle{}, err
	}
	defer rows.Close()

	activeCycles := make([]BillingCycle, 0)
	for rows.Next() {
		cycle := BillingCycle{Db: bs.Db}
		if err := rows.Scan(&cycle.Id); err != nil {
			return BillingCycle{}, err
		}
		activeCycles = append(activeCycles, cycle)
	}

	if len(activeCycles) > 1 {
		return BillingCycle{}, fmt.Errorf("error: found more than one active billing cycle")
	}

	if len(activeCycles) == 0 {
		return BillingCycle{}, fmt.Errorf(`no active billing cycle found`)
	}

	return activeCycles[0], nil
}

func (bs *BillingCycles) CurrentOrNew(tx *sql.Tx) (BillingCycle, error) {
	queryCheck := `SELECT id FROM accounting_billing_cycles WHERE status = 'active' FOR UPDATE`
	row := tx.QueryRow(queryCheck)

	cycle := BillingCycle{Db: bs.Db}
	if err := row.Scan(&cycle.Id); err != nil {
		if err == sql.ErrNoRows {
			queryInsert := `INSERT INTO accounting_billing_cycles (status) VALUES ('active') RETURNING id, created_at, status`
			err = tx.QueryRow(queryInsert).Scan(&cycle.Id)
			if err != nil {
				return BillingCycle{}, err
			}
		} else {
			return BillingCycle{}, err
		}
	}

	return cycle, nil
}

type BillingCycle struct {
	Db *sql.DB
	Id uint64
}

func (b *BillingCycle) Complete(tx *sql.Tx) error {
	query := `UPDATE accounting_billing_cycles SET status = 'complete' WHERE id = $1`
	_, err := tx.Exec(query, b.Id)
	return err
}

func (b *BillingCycle) Transactions() ([]Transaction, error) {
	query := `SELECT id, accounting_tasks.id, accounting_workers.id, withdrawal, enrolment FROM accounting_transactions WHERE id = $1`
	rows, err := b.Db.Query(query, b.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		t := Transaction{BillingCycleId: b.Id}
		if err := rows.Scan(&t.Id, &t.TaskId, &t.WorkerId, &t.Withdrawal, &t.Enrolment); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil

}
