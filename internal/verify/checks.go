package verify

import "context"

type CheckStatus string

const (
	CheckStatusPassed  CheckStatus = "passed"
	CheckStatusFailed  CheckStatus = "failed"
	CheckStatusSkipped CheckStatus = "skipped"
)

type Check struct {
	ID          string
	Description string
	Run         func(context.Context) error
}

type CheckResult struct {
	ID          string
	Description string
	Status      CheckStatus
	Error       string
}

func RunChecks(ctx context.Context, checks []Check) []CheckResult {
	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		result := CheckResult{ID: check.ID, Description: check.Description}
		if check.Run == nil {
			result.Status = CheckStatusSkipped
			result.Error = "check not implemented"
			results = append(results, result)
			continue
		}

		if err := check.Run(ctx); err != nil {
			result.Status = CheckStatusFailed
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		result.Status = CheckStatusPassed
		results = append(results, result)
	}

	return results
}
