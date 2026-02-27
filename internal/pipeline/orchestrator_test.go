package pipeline

import (
	"errors"
	"reflect"
	"testing"
)

func TestOrchestratorRunsPrepareThenApply(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(DefaultRollbackPolicy())

	result := orchestrator.Execute(StagePlan{
		Prepare: []Step{
			newTestStep("prepare-1", &order),
		},
		Apply: []Step{
			newTestStep("apply-1", &order),
		},
	})

	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !reflect.DeepEqual(order, []string{"run:prepare-1", "run:apply-1"}) {
		t.Fatalf("execution order = %v", order)
	}

	if !result.Prepare.Success || !result.Apply.Success {
		t.Fatalf("stage result = prepare:%v apply:%v", result.Prepare.Success, result.Apply.Success)
	}
}

func TestOrchestratorRollsBackApplyStepsOnFailure(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(DefaultRollbackPolicy())

	result := orchestrator.Execute(StagePlan{
		Apply: []Step{
			newRollbackStep("apply-1", &order, nil),
			newRollbackStep("apply-2", &order, errors.New("boom")),
		},
	})

	if result.Err == nil {
		t.Fatalf("Execute() expected apply error")
	}

	wantOrder := []string{"run:apply-1", "run:apply-2", "rollback:apply-1"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("execution order = %v, want %v", order, wantOrder)
	}

	if result.Rollback.Stage != StageRollback {
		t.Fatalf("rollback stage = %q", result.Rollback.Stage)
	}

	if !result.Rollback.Success {
		t.Fatalf("rollback expected success, got err = %v", result.Rollback.Err)
	}
}

func TestOrchestratorSkipsRollbackWhenPolicyDisabled(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(RollbackPolicy{OnApplyFailure: false})

	result := orchestrator.Execute(StagePlan{
		Apply: []Step{
			newRollbackStep("apply-1", &order, errors.New("boom")),
		},
	})

	if result.Err == nil {
		t.Fatalf("Execute() expected apply error")
	}

	if len(result.Rollback.Steps) != 0 {
		t.Fatalf("rollback steps = %d, want 0", len(result.Rollback.Steps))
	}

	if !reflect.DeepEqual(order, []string{"run:apply-1"}) {
		t.Fatalf("execution order = %v", order)
	}
}

type testStep struct {
	id      string
	order   *[]string
	runErr  error
	rollErr error
}

func newTestStep(id string, order *[]string) *testStep {
	return &testStep{id: id, order: order}
}

func newRollbackStep(id string, order *[]string, runErr error) *testStep {
	return &testStep{id: id, order: order, runErr: runErr}
}

func (s *testStep) ID() string {
	return s.id
}

func (s *testStep) Run() error {
	*s.order = append(*s.order, "run:"+s.id)
	return s.runErr
}

func (s *testStep) Rollback() error {
	*s.order = append(*s.order, "rollback:"+s.id)
	return s.rollErr
}
