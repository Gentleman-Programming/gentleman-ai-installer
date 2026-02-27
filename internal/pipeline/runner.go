package pipeline

import "time"

type Runner struct{}

func (r Runner) Run(stage Stage, steps []Step) StageResult {
	result := StageResult{Stage: stage, Success: true, Steps: make([]StepResult, 0, len(steps))}

	for _, step := range steps {
		started := time.Now().UTC()
		err := step.Run()
		finished := time.Now().UTC()

		stepResult := StepResult{
			StepID:     step.ID(),
			StartedAt:  started,
			FinishedAt: finished,
		}

		if err != nil {
			stepResult.Status = StepStatusFailed
			stepResult.Err = err
			result.Steps = append(result.Steps, stepResult)
			result.Success = false
			result.Err = err
			return result
		}

		stepResult.Status = StepStatusSucceeded
		result.Steps = append(result.Steps, stepResult)
	}

	return result
}
