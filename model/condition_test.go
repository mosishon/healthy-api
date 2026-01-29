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

	// آرگومان سوم (زمان) به تمام Evaluate ها اضافه شد
	dummyDuration := 100 * time.Millisecond

	t.Run("should pass with status 200 and correct body", func(t *testing.T) {
		resp := newMockResponse(200, "System status is OK")
		result := complexCondition.Evaluate(resp, []byte("System status is OK"), dummyDuration)
		if !result {
			t.Error("Expected true, but got false")
		}
	})

	t.Run("should pass with status 404", func(t *testing.T) {
		resp := newMockResponse(404, "Not Found")
		result := complexCondition.Evaluate(resp, []byte("Not Found"), dummyDuration)
		if !result {
			t.Error("Expected true, but got false")
		}
	})

	t.Run("should fail with status 200 and wrong body", func(t *testing.T) {
		resp := newMockResponse(200, "System status is Error")
		result := complexCondition.Evaluate(resp, []byte("System status is Error"), dummyDuration)
		if result {
			t.Error("Expected false, but got true")
		}
	})
}

func TestResponseTimeCondition(t *testing.T) {
	// استفاده از model. قبل از نام استراکت‌ها به دلیل پکیج model_test
	cond := &model.Condition{
		ResponseTime: &model.ResponseTimeCondition{
			MaxDuration: "500ms",
		},
	}

	// حالت موفق: ۲۰۰ میلی ثانیه < ۵۰۰
	if !cond.Evaluate(nil, nil, 200*time.Millisecond) {
		t.Errorf("Expected true for 200ms when limit is 500ms")
	}

	// حالت شکست: ۶۰۰ میلی ثانیه > ۵۰۰
	if cond.Evaluate(nil, nil, 600*time.Millisecond) {
		t.Errorf("Expected false for 600ms when limit is 500ms")
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

	// موفق: هر دو شرط برقرار است
	if !cond.Evaluate(resp200, nil, 100*time.Millisecond) {
		t.Errorf("Should be healthy with status 200 and fast response")
	}

	// شکست: کد ۲۰۰ است اما سرعت پایین است (۷۰۰ میلی ثانیه)
	if cond.Evaluate(resp200, nil, 700*time.Millisecond) {
		t.Errorf("Should fail because response is too slow")
	}

	// شکست: سرعت بالاست اما کد ۵۰۰ است
	if cond.Evaluate(resp500, nil, 100*time.Millisecond) {
		t.Errorf("Should fail because status code is 500")
	}
}

func TestValidation(t *testing.T) {
	// تست اعتبارسنجی فرمت زمان
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