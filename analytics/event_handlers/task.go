package event_handlers

import (
	"github.com/thekndr/ates/analytics/states"
	"time"
)

type (
	TaskAssignedEvent struct {
		Time     time.Time
		TaskId   string
		WorkerId string
		TaskCost uint
	}

	TaskAssigned struct {
		CompanyBalance *states.CompanyBalance
		WorkerBalance  *states.WorkerBalance
		Tasks          *states.Tasks
	}
)

func (h *TaskAssigned) Handle(ev TaskAssignedEvent) error {
	h.Tasks.Track(ev.Time, ev.TaskId, ev.TaskCost)
	h.CompanyBalance.AddProfit(ev.Time, ev.TaskId, ev.TaskCost)
	h.WorkerBalance.SubReward(ev.Time, ev.WorkerId, ev.TaskId, ev.TaskCost)
	return nil
}

type (
	TaskCompletedEvent struct {
		Time       time.Time
		TaskId     string
		WorkerId   string
		TaskReward uint
	}

	TaskCompleted struct {
		CompanyBalance *states.CompanyBalance
		WorkerBalance  *states.WorkerBalance
	}
)

func (h *TaskCompleted) Handle(ev TaskCompletedEvent) error {
	h.CompanyBalance.SubProfit(ev.Time, ev.TaskId, ev.TaskReward)
	h.WorkerBalance.AddReward(ev.Time, ev.WorkerId, ev.TaskId, ev.TaskReward)
	return nil
}
