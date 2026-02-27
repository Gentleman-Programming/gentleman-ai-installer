package pipeline

import "time"

type StepStatus string

const (
	StepStatusSucceeded  StepStatus = "succeeded"
	StepStatusFailed     StepStatus = "failed"
	StepStatusRolledBack StepStatus = "rolled-back"
)

type StepResult struct {
	StepID     string
	Status     StepStatus
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error
}

type StageResult struct {
	Stage   Stage
	Steps   []StepResult
	Success bool
	Err     error
}

type ExecutionResult struct {
	Prepare  StageResult
	Apply    StageResult
	Rollback StageResult
	Err      error
}
