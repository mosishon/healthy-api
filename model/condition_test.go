package model_test

import (
	"healthy-api/model"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// تابع کمکی برای شبیه‌سازی پاسخ HTTP
func newMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestCondition_EvaluateComplexNested(t *testing.T) {
	complexCondition := &model.Condition{
		Or: []*model.Condition{
			{
				And: []*model.Condition{
					{StatusCode: &model.StatusCodeCondition{Code: 200}},
					{Regex: &model.RegexCondition{Regex: "OK"}},
				},
			},
			{
				StatusCode: &model.StatusCodeCondition{Code: 404},
			},
		},
	}

	dummyDuration := 100 * time.Millisecond

	t.Run("should pass with status 200 and correct body", func(t *testing.T) {
		resp := newMockResponse(200, "System status is OK")
		result := complexCondition.Evaluate(resp, []byte("System status is OK"), dummyDuration)
		// اصلاح: استفاده از .IsHealthy
		if !result.IsHealthy {
			t.Errorf("Expected true, but got false. Reason: %s", result.Reason)
		}
	})

	t.Run("should pass with status 404", func(t *testing.T) {
		resp := newMockResponse(404, "Not Found")
		result := complexCondition.Evaluate(resp, []byte("Not Found"), dummyDuration)
		// اصلاح: استفاده از .IsHealthy
		if !result.IsHealthy {
			t.Errorf("Expected true, but got false. Reason: %s", result.Reason)
		}
	})

	t.Run("should fail with status 200 and wrong body", func(t *testing.T) {
		resp := newMockResponse(200, "System status is Error")
		result := complexCondition.Evaluate(resp, []byte("System status is Error"), dummyDuration)
		// اصلاح: استفاده از .IsHealthy
		if result.IsHealthy {
			t.Error("Expected false, but got true")
		}
	})
}

func TestResponseTimeCondition(t *testing.T) {
	cond := &model.Condition{
		ResponseTime: &model.ResponseTimeCondition{
			MaxDuration: "500ms",
		},
	}

	// حالت موفق
	resultOk := cond.Evaluate(nil, nil, 200*time.Millisecond)
	if !resultOk.IsHealthy {
		t.Errorf("Expected true for 200ms, got false. Reason: %s", resultOk.Reason)
	}

	// حالت شکست
	resultFail := cond.Evaluate(nil, nil, 600*time.Millisecond)
	if resultFail.IsHealthy {
		t.Errorf("Expected false for 600ms, but got true")
	}
}

func TestCombinedCondition(t *testing.T) {
	cond := &model.Condition{
		And: []*model.Condition{
			{StatusCode: &model.StatusCodeCondition{Code: 200}},
			{ResponseTime: &model.ResponseTimeCondition{MaxDuration: "500ms"}},
		},
	}

	resp200 := &http.Response{StatusCode: 200}
	resp500 := &http.Response{StatusCode: 500}

	// موفق
	res1 := cond.Evaluate(resp200, nil, 100*time.Millisecond)
	if !res1.IsHealthy {
		t.Errorf("Should be healthy, but got: %s", res1.Reason)
	}

	// شکست (سرعت پایین)
	res2 := cond.Evaluate(resp200, nil, 700*time.Millisecond)
	if res2.IsHealthy {
		t.Error("Should fail because response is too slow")
	}

	// شکست (کد وضعیت غلط)
	res3 := cond.Evaluate(resp500, nil, 100*time.Millisecond)
	if res3.IsHealthy {
		t.Error("Should fail because status code is 500")
	}
}

func TestValidation(t *testing.T) {
	invalidCond := &model.Condition{
		ResponseTime: &model.ResponseTimeCondition{
			MaxDuration: "invalid-time",
		},
	}

	if err := invalidCond.Validate("test"); err == nil {
		t.Error("Validation should fail for invalid duration format")
	}

	validCond := &model.Condition{
		ResponseTime: &model.ResponseTimeCondition{
			MaxDuration: "1.5s",
		},
	}

	if err := validCond.Validate("test"); err != nil {
		t.Errorf("Validation should pass for '1.5s', got: %v", err)
	}
}