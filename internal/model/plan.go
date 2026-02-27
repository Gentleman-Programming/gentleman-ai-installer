package model

type PlanStatus string

const (
	PlanStatusPending   PlanStatus = "pending"
	PlanStatusRunning   PlanStatus = "running"
	PlanStatusSucceeded PlanStatus = "succeeded"
	PlanStatusFailed    PlanStatus = "failed"
)

type RunResult string

const (
	RunResultSkipped RunResult = "skipped"
	RunResultSuccess RunResult = "success"
	RunResultFailed  RunResult = "failed"
)

type PlanStep struct {
	ID     string
	Name   string
	Status PlanStatus
	Result RunResult
	Error  string
}

type Plan struct {
	ID        string
	Selection Selection
	Status    PlanStatus
	Steps     []PlanStep
}
