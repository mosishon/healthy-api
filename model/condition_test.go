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
