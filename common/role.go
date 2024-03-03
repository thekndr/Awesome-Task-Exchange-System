package common

type Role string

const (
	WorkerRole  Role = "worker"
	ManagerRole Role = "manager"
	AdminRole   Role = "admin"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return r == WorkerRole || r == ManagerRole || r == AdminRole
}
