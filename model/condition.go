package model

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

)

type ConditionType string

const (
	ConditionRegex      ConditionType = "regex"
	ConditionStatusCode ConditionType = "status_code"
	ConditionHeader     ConditionType = "header"
	ConditionAnd        ConditionType = "and"
	ConditionOr         ConditionType = "or"
	ConditionNot        ConditionType = "not"
	ConditionResponseTime ConditionType = "response_time"
)

type Condition struct {
	And        []*Condition         `yaml:"and,omitempty"`
	Or         []*Condition         `yaml:"or,omitempty"`
	Not        *Condition           `yaml:"not,omitempty"`
	Regex      *RegexCondition      `yaml:"regex,omitempty"`
	StatusCode *StatusCodeCondition `yaml:"status_code,omitempty"`
	Header     *[]HeaderCondition   `yaml:"header,omitempty"`
	ResponseTime *ResponseTimeCondition `yaml:"response_time,omitempty"`
}

type NamedCondition struct {
	ID        string     `yaml:"id"`
	Condition *Condition `yaml:"condition"`
}

type RegexCondition struct {
	Regex string `yaml:"pattern"`
}

type StatusCodeCondition struct {
	Code int `yaml:"code"`
}

type HeaderCondition struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}
type ResponseTimeCondition struct {
	MaxDuration string `yaml:"max_duration"`
}
type EvaluationResult struct {
	IsHealthy bool
	Reason    string
}

func (c *Condition) Validate(path string) error {
	count := 0
	if c.Regex != nil {
		count++
	}
	if c.StatusCode != nil {
		count++
	}
	if c.Header != nil {
		count++
	}
	if c.And != nil {
		count++
	}
	if c.Or != nil {
		count++
	}
	if c.Not != nil {
		count++
	}
	if c.ResponseTime != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("a condition node must contain exactly one field (got %d) at %s", count, path)
	}
	if c.ResponseTime != nil {
		if _, err := time.ParseDuration(c.ResponseTime.MaxDuration); err != nil {
			return fmt.Errorf("invalid duration format '%s' at %s: %v", c.ResponseTime.MaxDuration, path, err)
		}
	}
	for _, and := range c.And {
		path = path + "." + "and"
		if err := and.Validate(path); err != nil {
			return err
		}
	}
	for _, or := range c.And {
		path = path + "." + "or"
		if err := or.Validate(path); err != nil {
			return err
		}
	}
	return nil
}
func (c *Condition) Evaluate(resp *http.Response, body []byte, duration time.Duration) EvaluationResult {
	// 1. منطق AND
	if c.And != nil {
		for _, cond := range c.And {
			res := cond.Evaluate(resp, body, duration)
			if !res.IsHealthy {
				return res
			}
		}
		return EvaluationResult{IsHealthy: true}
	}

	// 2. منطق OR
	if c.Or != nil {
		var reasons []string
		for i, cond := range c.Or {
			res := cond.Evaluate(resp, body, duration)
			if res.IsHealthy {
				return EvaluationResult{IsHealthy: true}
			}
			reasons = append(reasons, fmt.Sprintf("OR[%d]: %s", i, res.Reason))
		}
		return EvaluationResult{
			IsHealthy: false,
			Reason:    fmt.Sprintf("All OR conditions failed: %v", reasons),
		}
	}

	// 3. منطق NOT
	if c.Not != nil {
    res := c.Not.Evaluate(resp, body, duration)
    if res.IsHealthy {
        return EvaluationResult{
            IsHealthy: false,
            Reason:    "Forbidden condition matched (Service should not have met this condition)",
        }
    }
    return EvaluationResult{IsHealthy: true}
}

	// 4. بررسی Regex
	if c.Regex != nil {
		matched, _ := regexp.Match(c.Regex.Regex, body)
		if !matched {
			return EvaluationResult{
				IsHealthy: false,
				Reason:    fmt.Sprintf("Regex pattern '%s' not found in body", c.Regex.Regex),
			}
		}
		return EvaluationResult{IsHealthy: true}
	}

	// 5. بررسی StatusCode
	if c.StatusCode != nil {
		if resp == nil {
			return EvaluationResult{IsHealthy: false, Reason: "No response received"}
		}
		if resp.StatusCode != c.StatusCode.Code {
			return EvaluationResult{
				IsHealthy: false,
				Reason:    fmt.Sprintf("Expected status %d, but got %d", c.StatusCode.Code, resp.StatusCode),
			}
		}
		return EvaluationResult{IsHealthy: true}
	}

	// 6. بررسی Headers
	if c.Header != nil {
		if resp == nil {
			return EvaluationResult{IsHealthy: false, Reason: "No response headers available"}
		}
		for _, h := range *c.Header {
			actual := resp.Header.Get(h.Key)
			if actual != h.Value {
				return EvaluationResult{
					IsHealthy: false,
					Reason:    fmt.Sprintf("Header '%s' expected '%s', got '%s'", h.Key, h.Value, actual),
				}
			}
		}
		return EvaluationResult{IsHealthy: true}
	}

	// 7. بررسی Response Time
	if c.ResponseTime != nil {
		max, _ := time.ParseDuration(c.ResponseTime.MaxDuration)
		if duration > max {
			return EvaluationResult{
				IsHealthy: false,
				Reason:    fmt.Sprintf("Response time %v exceeded limit %v", duration, max),
			}
		}
		return EvaluationResult{IsHealthy: true}
	}

	return EvaluationResult{IsHealthy: false, Reason: "No valid condition defined"}
}

func (r *RegexCondition) Evaluate(body []byte) bool {
	matched, err := regexp.Match(r.Regex, body)
	return err == nil && matched
}

func (s *StatusCodeCondition) Evaluate(resp *http.Response) bool {

	return resp.StatusCode == s.Code
}

func (h *HeaderCondition) Evaluate(resp *http.Response) bool {

	return resp.Header.Get(h.Key) == h.Value
}

func (rt *ResponseTimeCondition) Evaluate(actual time.Duration) bool {
	max, err := time.ParseDuration(rt.MaxDuration)
	if err != nil {
		return false 
	}
	return actual <= max
}

// TODO: jsonpath condition
