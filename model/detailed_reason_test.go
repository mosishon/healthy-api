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
		reason := strings.Join(result.Reason, "\n")
		if !strings.Contains(reason, "Body Pattern: Expected 'UP', Matched: false") {
			t.Errorf("expected original failure message to be included, got: %s", reason)
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
		reason := strings.Join(result.Reason, "\n")
		if !strings.Contains(reason, "Status Code: Expected 200, Got 500") {
			t.Errorf("expected status code failure, got: %s", reason)
		}
		if !strings.Contains(reason, "Body Pattern: Expected 'UP', Matched: false") {
			t.Errorf("expected regex failure, got: %s", reason)
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
		reason := strings.Join(result.Reason, "\n")
		if !strings.Contains(reason, "✘ Forbidden state matched: Status Code: Expected 500, Got 500") {
			t.Errorf("expected NOT failure message, got: %s", reason)
		}
	})
}
