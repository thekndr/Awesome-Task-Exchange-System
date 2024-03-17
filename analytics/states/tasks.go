package states

import (
	"sync"
	"time"
)

type (
	taskEntry struct {
		TaskId   string
		TaskCost uint
		// TODO: TaskReward for tracking expensive tasks from company perspective
	}

	Tasks struct {
		log   map[time.Time][]taskEntry
		guard sync.RWMutex
	}
)

func (t *Tasks) Track(timepoint time.Time, taskId string, cost uint) {
	t.guard.Lock()
	defer t.guard.Unlock()

	timepoint = TruncateTimeForLogging(timepoint)

	t.log[timepoint] = append(t.log[timepoint], taskEntry{
		TaskId: taskId, TaskCost: cost,
	})
}

func (t *Tasks) MostExpensiveUserTask(timepoint time.Time) (taskId string, taskCost uint, ok bool) {
	timepoint = TruncateTimeForLogging(timepoint)

	entries, entriesOk := t.log[timepoint]
	if !entriesOk {
		return "", 0, false
	}

	var result taskEntry
	for _, entry := range entries {
		if entry.TaskCost > result.TaskCost {
			result = entry
		}
	}

	return result.TaskId, result.TaskCost, true
}
