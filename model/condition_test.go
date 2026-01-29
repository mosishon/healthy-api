package model_test

import (
	"healthy-api/model"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func newMockResponse(statusCode int, body string, headers http.Header) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		Header:     headers,
	}
}

func TestCondition_EvaluateComplexNested(t *testing.T) {
	complexCondition := &model.Condition{
		Or: []*model.Condition{
			{
				And: []*model.Condition{
					{
						StatusCode: &model.StatusCodeCondition{Code: 200},
					},
					{
						Regex: &model.RegexCondition{Regex: "OK"},
					},
				},
			},
			{
				StatusCode: &model.StatusCodeCondition{Code: 404},
			},
		},
	}

	// Test Case 1: Should pass because of the first AND block
	t.Run("should pass with status 200 and correct body", func(t *testing.T) {
		// Arrange
		resp := newMockResponse(200, "System status is OK", nil)

		// Act
		result := complexCondition.Evaluate(resp, []byte("System status is OK"))

		// Assert
		if !result {
			t.Error("Expected condition to evaluate to true, but got false")
		}
	})

	// Test Case 2: Should pass because of the second OR block
	t.Run("should pass with status 404", func(t *testing.T) {
		// Arrange
		resp := newMockResponse(404, "Not Found", nil)

		// Act
		result := complexCondition.Evaluate(resp, []byte("Not Found"))

		// Assert
		if !result {
			t.Error("Expected condition to evaluate to true, but got false")
		}
	})

	// Test Case 3: Should fail because neither condition is met
	t.Run("should fail with status 200 and wrong body", func(t *testing.T) {
		// Arrange
		resp := newMockResponse(200, "System status is Error", nil)

		// Act
		result := complexCondition.Evaluate(resp, []byte("System status is Error"))

		// Assert
		if result {
			t.Error("Expected condition to evaluate to false, but got true")
		}
	})

	// Test Case 4: Should fail with a completely different status
	t.Run("should fail with status 500", func(t *testing.T) {
		// Arrange
		resp := newMockResponse(500, "Server Error", nil)

		// Act
		result := complexCondition.Evaluate(resp, []byte("Server Error"))

		// Assert
		if result {
			t.Error("Expected condition to evaluate to false, but got true")
		}
	})
}


func TestResponseTimeCondition(t *testing.T) {
	// ۱. تست شرط Response Time به تنهایی
	cond := &Condition{
		ResponseTime: &ResponseTimeCondition{
			MaxDuration: "500ms",
		},
	}

	// حالت موفق: زمان واقعی ۲۰۰ میلی ثانیه (کمتر از ۵۰۰)
	if !cond.Evaluate(nil, nil, 200*time.Millisecond) {
		t.Errorf("Expected true for 200ms when limit is 500ms")
	}

	// حالت شکست: زمان واقعی ۶۰۰ میلی ثانیه (بیشتر از ۵۰۰)
	if cond.Evaluate(nil, nil, 600*time.Millisecond) {
		t.Errorf("Expected false for 600ms when limit is 500ms")
	}
}

func TestCombinedCondition(t *testing.T) {
	// ۲. تست ترکیبی: کد ۲٠٠ باشد AND زمان پاسخ زیر ۵۰۰ میلی ثانیه باشد
	cond := &Condition{
		And: []*Condition{
			{StatusCode: &StatusCodeCondition{Code: 200}},
			{ResponseTime: &ResponseTimeCondition{MaxDuration: "500ms"}},
		},
	}

	resp200 := &http.Response{StatusCode: 200}
	resp500 := &http.Response{StatusCode: 500}

	// موفق: هر دو شرط برقرار است
	if !cond.Evaluate(resp200, nil, 100*time.Millisecond) {
		t.Errorf("Should be healthy with status 200 and fast response")
	}

	// شکست: کد ۲٠٠ است اما سرعت پایین است
	if cond.Evaluate(resp200, nil, 700*time.Millisecond) {
		t.Errorf("Should fail because response is too slow")
	}

	// شکست: سرعت بالاست اما کد ۵٠٠ است
	if cond.Evaluate(resp500, nil, 100*time.Millisecond) {
		t.Errorf("Should fail because status code is 500")
	}
}

func TestValidation(t *testing.T) {
	// ۳. تست اعتبارسنجی فرمت زمان
	invalidCond := &Condition{
		ResponseTime: &ResponseTimeCondition{
			MaxDuration: "invalid-time",
		},
	}

	if err := invalidCond.Validate("test"); err == nil {
		t.Error("Validation should fail for invalid duration format")
	}

	validCond := &Condition{
		ResponseTime: &ResponseTimeCondition{
			MaxDuration: "1.5s",
		},
	}

	if err := validCond.Validate("test"); err != nil {
		t.Errorf("Validation should pass for '1.5s', got: %v", err)
	}
}