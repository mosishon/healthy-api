package model_test

import (
	"healthy-api/model"
	"testing"
)

func TestCondition_ValidateNested(t *testing.T) {
	t.Run("valid nested AND/OR", func(t *testing.T) {
		cond := &model.Condition{
			And: []*model.Condition{
				{StatusCode: &model.StatusCodeCondition{Code: 200}},
				{
					Or: []*model.Condition{
						{Regex: &model.RegexCondition{Regex: "A"}},
						{Regex: &model.RegexCondition{Regex: "B"}},
					},
				},
			},
		}
		if err := cond.Validate("root"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid nested OR (bug check)", func(t *testing.T) {
		// This used to pass incorrectly if it only checked c.And
		cond := &model.Condition{
			Or: []*model.Condition{
				{
					// Invalid: has both StatusCode and Regex
					StatusCode: &model.StatusCodeCondition{Code: 200},
					Regex:      &model.RegexCondition{Regex: "A"},
				},
			},
		}
		if err := cond.Validate("root"); err == nil {
			t.Error("expected validation error for invalid nested condition in OR branch")
		}
	})

	t.Run("valid NOT", func(t *testing.T) {
		cond := &model.Condition{
			Not: &model.Condition{
				StatusCode: &model.StatusCodeCondition{Code: 500},
			},
		}
		if err := cond.Validate("root"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}
