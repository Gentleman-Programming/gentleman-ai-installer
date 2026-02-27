package pipeline

import "fmt"

type RollbackPolicy struct {
	OnApplyFailure bool
}

func DefaultRollbackPolicy() RollbackPolicy {
	return RollbackPolicy{OnApplyFailure: true}
}

func (p RollbackPolicy) ShouldRollback(stage Stage, err error) bool {
	if err == nil {
		return false
	}

	return stage == StageApply && p.OnApplyFailure
}

func ExecuteRollback(steps []StepResult, stepIndex map[string]Step) StageResult {
	result := StageResult{Stage: StageRollback, Success: true}

	for i := len(steps) - 1; i >= 0; i-- {
		stepResult := steps[i]
		if stepResult.Status != StepStatusSucceeded {
			continue
		}

		step, ok := stepIndex[stepResult.StepID]
		if !ok {
			continue
		}

		rollbackStep, ok := step.(RollbackStep)
		if !ok {
			continue
		}

		err := rollbackStep.Rollback()
		item := StepResult{StepID: rollbackStep.ID(), Status: StepStatusRolledBack}
		if err != nil {
			item.Status = StepStatusFailed
			item.Err = err
			result.Steps = append(result.Steps, item)
			result.Success = false
			result.Err = fmt.Errorf("rollback step %q: %w", rollbackStep.ID(), err)
			return result
		}

		result.Steps = append(result.Steps, item)
	}

	return result
}
