package db

import (
	"database/sql"
	"fmt"
)

type Workers struct {
	Db *sql.DB
}

func (us *Workers) Add(publicId, email string) (uint64, error) {
	query := `INSERT INTO accounting_workers (public_id, email, balance) VALUES ($1, $2, 0) ON CONFLICT DO NOTHING RETURNING id`

	var workerInternalId uint64
	err := us.Db.QueryRow(query, publicId, email).Scan(&workerInternalId)
	if err != nil {
		return 0, fmt.Errorf(`workers: insertion to workers failed=%w`, err)
	}

	return workerInternalId, nil
}

func (us *Workers) Get(id uint64) (Worker, error) {
	query := `SELECT id, public_id, email FROM accounting_workers WHERE id = $1`
	var worker Worker

	err := us.Db.QueryRow(query, id).Scan(&worker.Id, &worker.PublicId, &worker.Email)
	if err != nil {
		return Worker{}, err
	}

	return worker, nil
}

func (us *Workers) GetByPublic(id string) (Worker, error) {
	query := `SELECT id, public_id, email FROM accounting_workers WHERE public_id = $1`
	var worker Worker

	err := us.Db.QueryRow(query, id).Scan(&worker.Id, &worker.PublicId, &worker.Email)
	if err != nil {
		return Worker{}, err
	}

	return worker, nil
}

type Worker struct {
	Id       uint64
	PublicId string
	Email    string
}

func (w *Worker) Withdraw(tx *sql.Tx, billingCycleId uint64, amount uint) (int, error) {
	updateBalanceQuery := `UPDATE accounting_workers SET balance = balance - $1 WHERE id = $2 RETURNING balance`
	var newBalance int
	err := tx.QueryRow(updateBalanceQuery, amount, w.Id).Scan(&newBalance)
	if err != nil {
		return 0, fmt.Errorf(`failed to withdraw user=%d amount=%d bid=%d: %w`, w.Id, amount, billingCycleId, err)
	}
	return newBalance, nil
}

func (w *Worker) Enroll(tx *sql.Tx, billingCycleId uint64, amount uint) (int, error) {
	updateBalanceQuery := `UPDATE accounting_workers SET balance = balance + $1 WHERE id = $2 RETURNING balance`
	var newBalance int
	err := tx.QueryRow(updateBalanceQuery, amount, w.Id).Scan(&newBalance)
	if err != nil {
		return 0, fmt.Errorf(`failed to enroll user=%d amount=%d bid=%d: %w`, w.Id, amount, billingCycleId, err)
	}
	return newBalance, nil
}
