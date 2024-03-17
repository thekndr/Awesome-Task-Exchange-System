package states

import (
	"sync"
	"time"
)

type (
	workerProfitEntry struct {
		Amount   int
		TaskId   string
		WorkerId string
	}

	WorkerBalance struct {
		log   map[time.Time][]workerProfitEntry
		guard sync.RWMutex
	}
)

func (wb *WorkerBalance) AddReward(timepoint time.Time, workerId string, taskId string, amount uint) {
	wb.guard.Lock()
	defer wb.guard.Unlock()

	timepoint = TruncateTimeForLogging(timepoint)

	wb.log[timepoint] = append(wb.log[timepoint], workerProfitEntry{
		Amount: int(amount), TaskId: taskId, WorkerId: workerId,
	})
}

func (wb *WorkerBalance) SubReward(timepoint time.Time, workerId string, taskId string, amount uint) {
	wb.guard.Lock()
	defer wb.guard.Unlock()

	timepoint = TruncateTimeForLogging(timepoint)

	wb.log[timepoint] = append(wb.log[timepoint], workerProfitEntry{
		Amount: int(-amount), TaskId: taskId, WorkerId: workerId,
	})
}

func (wb *WorkerBalance) WorkersWithNegativeBalance(timepoint time.Time) int {
	wb.guard.RLock()
	defer wb.guard.RUnlock()

	timepoint = TruncateTimeForLogging(timepoint)
	entries, ok := wb.log[timepoint]
	if !ok {
		return 0
	}

	sums := make(map[string]int)
	for _, entry := range entries {
		sums[entry.WorkerId] += entry.Amount
	}

	count := 0
	for _, amount := range sums {
		if amount < 0 {
			count++
		}
	}

	return count
}
