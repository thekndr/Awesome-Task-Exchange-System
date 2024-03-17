package states

import (
	"sync"
	"time"
)

type (
	companyProfitEntry struct {
		Amount int
		TaskId string
	}

	CompanyBalance struct {
		log   map[time.Time][]companyProfitEntry
		guard sync.RWMutex
	}
)

func (cb *CompanyBalance) ProfitFor(timepoint time.Time) int {
	cb.guard.RLock()
	defer cb.guard.RUnlock()

	timepoint = TruncateTimeForLogging(timepoint)

	entries, ok := cb.log[timepoint]
	if !ok {
		return 0
	}

	var profit int
	for _, entry := range entries {
		profit += entry.Amount
	}

	return profit
}

func (cb *CompanyBalance) AddProfit(timepoint time.Time, taskId string, amount uint) {
	cb.guard.Lock()
	defer cb.guard.Unlock()

	timepoint = TruncateTimeForLogging(timepoint)
	cb.log[timepoint] = append(cb.log[timepoint], companyProfitEntry{
		Amount: int(amount), TaskId: taskId,
	})
}

func (cb *CompanyBalance) SubProfit(timepoint time.Time, taskId string, amount uint) {
	cb.guard.Lock()
	defer cb.guard.Unlock()

	timepoint = TruncateTimeForLogging(timepoint)
	cb.log[timepoint] = append(cb.log[timepoint], companyProfitEntry{
		Amount: int(-amount), TaskId: taskId,
	})
}

func TruncateTimeForLogging(tp time.Time) time.Time {
	return time.Date(
		tp.Year(), tp.Month(), tp.Day(), 0, 0, 0, 0, tp.Location(),
	)
}
