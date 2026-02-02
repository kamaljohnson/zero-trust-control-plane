package engine

import (
	"context"
	"testing"
)

func TestOPAEvaluator_HealthCheck(t *testing.T) {
	// OPAEvaluator needs a policy repo for NewOPAEvaluator; HealthCheck does not use it.
	e := NewOPAEvaluator(nil)
	ctx := context.Background()
	if err := e.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}
