package model_test

import (
	"healthy-api/model"
	"strings"
	"testing"
	"time"
)

func TestEvaluate_DetailedReasons(t *testing.T) {
	dummyDuration := 100 * time.Millisecond

	t.Run("AND failure detail", func(t *testing.T) {
		cond := &model.Condition{
			And: []*model.Condition{
				{StatusCode: &model.StatusCodeCondition{Code: 200}},
				{Regex: &model.RegexCondition{Regex: "UP"}},
			},
		}
		resp := newMockResponse(200, "DOWN")
		result := cond.Evaluate(resp, []byte("DOWN"), dummyDuration)

		if result.IsHealthy {
			t.Fatal("expected failure")
		}
		if !strings.Contains(result.Reason, "✘ Body Match: Pattern 'UP' not found") {
			t.Errorf("expected original failure message to be included, got: %s", result.Reason)
		}
	})

	t.Run("OR failure detail aggregation", func(t *testing.T) {
		cond := &model.Condition{
			Or: []*model.Condition{
				{StatusCode: &model.StatusCodeCondition{Code: 200}},
				{Regex: &model.RegexCondition{Regex: "UP"}},
			},
		}
		resp := newMockResponse(500, "DOWN")
		result := cond.Evaluate(resp, []byte("DOWN"), dummyDuration)

		if result.IsHealthy {
			t.Fatal("expected failure")
		}
		if !strings.Contains(result.Reason, "✘ All OR conditions failed") {
			t.Errorf("expected OR failure message, got: %s", result.Reason)
		}
		if !strings.Contains(result.Reason, "Status Code: Expected 200, Got 500") {
			t.Errorf("expected status code failure, got: %s", result.Reason)
		}
		if !strings.Contains(result.Reason, "Body Match: Pattern 'UP' not found in response") {
			t.Errorf("expected regex failure, got: %s", result.Reason)
		}
	})

	t.Run("NOT failure detail", func(t *testing.T) {
		cond := &model.Condition{
			Not: &model.Condition{
				StatusCode: &model.StatusCodeCondition{Code: 500},
			},
		}
		resp := newMockResponse(500, "")
		result := cond.Evaluate(resp, nil, dummyDuration)

		if result.IsHealthy {
			t.Fatal("expected failure")
		}
		if !strings.Contains(result.Reason, "✘ Forbidden Status Code: Received 500") {
			t.Errorf("expected NOT failure message, got: %s", result.Reason)
		}
	})
}
