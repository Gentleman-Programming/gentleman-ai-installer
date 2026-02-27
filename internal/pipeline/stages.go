package pipeline

type Stage string

const (
	StagePrepare  Stage = "prepare"
	StageApply    Stage = "apply"
	StageRollback Stage = "rollback"
)

type Step interface {
	ID() string
	Run() error
}

type RollbackStep interface {
	Step
	Rollback() error
}

type StagePlan struct {
	Prepare []Step
	Apply   []Step
}
