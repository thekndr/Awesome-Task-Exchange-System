package common

type TaskStatus int

const (
	TaskActiveStatus TaskStatus = 0
	TaskDoneStatus   TaskStatus = 1
)

func (t TaskStatus) String() string {
	switch t {
	case TaskActiveStatus:
		return "active"
	case TaskDoneStatus:
		return "done"
	default:
		return "unknown"
	}
}
