package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
)

type Tasks struct {
	Db *sql.DB
}

func (ts *Tasks) Get(internalId int) (Task, error) {
	query := `SELECT id, public_id, description, completed, assignment_price, reward_price FROM accounting_tasks WHERE id = $1`
	var task Task

	err := ts.Db.QueryRow(query, internalId).Scan(&task.Id, &task.PublicId, &task.Description, &task.Completed, &task.AssignmentPrice, &task.RewardPrice)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (ts *Tasks) GetByPublic(id string) (Task, error) {
	query := `SELECT id, public_id, description, completed, assignment_price, reward_price FROM accounting_tasks WHERE public_id = $1`
	var task Task

	err := ts.Db.QueryRow(query, id).Scan(&task.Id, &task.PublicId, &task.Description, &task.Completed, &task.AssignmentPrice, &task.RewardPrice)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (ts *Tasks) Add(publicId, description string) (int64, error) {
	assignmentPrice, rewardPrice := calcTaskPrices()

	result, err := ts.Db.Exec(`INSERT
INTO accounting_tasks (public_id, description, assignment_price, reward_price)
VALUES ($1, $2, $3, $4)
`, publicId, description, assignmentPrice, rewardPrice)
	if err != nil {
		return -1, fmt.Errorf(`tasks: failed to insert=%w`, err)
	}

	var id int64
	if id, err = result.LastInsertId(); err != nil {
		log.Printf(`tasks: failed to retrieve last-inserted id`)
		return -1, err
	}

	log.Printf(`tasks: task id=%s prices=(a:%d, r:%d) added with id=%d`, publicId, assignmentPrice, rewardPrice, id)
	return id, nil
}

func calcTaskPrices() (assignment, reward int) {
	assignment = rand.Intn(20-10+1) + 10
	reward = rand.Intn(40-20+1) + 20
	return
}

type Task struct {
	Id              uint64
	PublicId        string
	AssignmentPrice uint
	RewardPrice     uint
	Completed       bool
	Description     string
}
