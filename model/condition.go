package model

import (
	"fmt"
	"net/http"
	"regexp"
	"time"
)

type ConditionType string

const (
	ConditionRegex        ConditionType = "regex"
	ConditionStatusCode   ConditionType = "status_code"
	ConditionHeader       ConditionType = "header"
	ConditionAnd          ConditionType = "and"
	ConditionOr           ConditionType = "or"
	ConditionNot          ConditionType = "not"
	ConditionResponseTime ConditionType = "response_time"
)

type Condition struct {
	And          []*Condition           `yaml:"and,omitempty"`
	Or           []*Condition           `yaml:"or,omitempty"`
	Not          *Condition             `yaml:"not,omitempty"`
	Regex        *RegexCondition        `yaml:"regex,omitempty"`
	StatusCode   *StatusCodeCondition   `yaml:"status_code,omitempty"`
	Header       *[]HeaderCondition     `yaml:"header,omitempty"`
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
type NotificationType string

const (
	NotificationNetworkError    NotificationType = "network_error"
	NotificationHttpError       NotificationType = "http_error"
	NotificationSlowResponse    NotificationType = "slow_response"
	NotificationConditionFailed NotificationType = "condition_failed"
	NotificationRecovery        NotificationType = "recovery"
	NotificationDefault         NotificationType = "default"
)

type EvaluationResult struct {
	IsHealthy bool
	Reason    []string
	Type      NotificationType
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
	for i, and := range c.And {
		if err := and.Validate(fmt.Sprintf("%s.and[%d]", path, i)); err != nil {
			return err
		}
	}
	for i, or := range c.Or {
		if err := or.Validate(fmt.Sprintf("%s.or[%d]", path, i)); err != nil {
			return err
		}
	}
	if c.Not != nil {
		if err := c.Not.Validate(path + ".not"); err != nil {
			return err
		}
	}
	return nil
}

func (c *Condition) Evaluate(resp *http.Response, body []byte, duration time.Duration) EvaluationResult {
	// 1. Evaluate NOT Logic (First Priority)
	if c.Not != nil {
		res := c.Not.Evaluate(resp, body, duration)
		if res.IsHealthy {
			// Inversion: The forbidden condition matched!
			var wrappedReasons []string
			for _, r := range res.Reason {
				wrappedReasons = append(wrappedReasons, "✘ Forbidden state matched (NOT): "+r)
			}
			return EvaluationResult{
				IsHealthy: false,
				Reason:    wrappedReasons,
				Type:      NotificationConditionFailed,
			}
		}
		// If inner was NOT healthy, then NOT(Unhealthy) is Healthy.
		return EvaluationResult{IsHealthy: true}
	}

	// 2. Evaluate AND Logic
	if c.And != nil {
		var failures []string
		var firstType NotificationType
		for _, cond := range c.And {
			res := cond.Evaluate(resp, body, duration)
			if !res.IsHealthy {
				if firstType == "" {
					firstType = res.Type
				}
				failures = append(failures, res.Reason...)
			}
		}
		if len(failures) > 0 {
			return EvaluationResult{
				IsHealthy: false,
				Reason:    failures,
				Type:      firstType,
			}
		}
		return EvaluationResult{IsHealthy: true}
	}

	// 3. Evaluate OR Logic
	if c.Or != nil {
		var subFailures []string
		for _, cond := range c.Or {
			res := cond.Evaluate(resp, body, duration)
			if res.IsHealthy {
				return EvaluationResult{IsHealthy: true}
			}
			subFailures = append(subFailures, res.Reason...)
		}
		return EvaluationResult{
			IsHealthy: false,
			Reason:    subFailures,
			Type:      NotificationConditionFailed,
		}
	}

	// 4. Leaf Conditions - Always return detailed Reason even on success (for NOT inversion)

	// Regex
	if c.Regex != nil {
		matched, _ := regexp.Match(c.Regex.Regex, body)
		reason := fmt.Sprintf("Body Pattern: Expected '%s', Matched: %v", c.Regex.Regex, matched)
		if !matched {
			return EvaluationResult{IsHealthy: false, Reason: []string{reason}, Type: NotificationConditionFailed}
		}
		return EvaluationResult{IsHealthy: true, Reason: []string{reason}}
	}

	// StatusCode
	if c.StatusCode != nil {
		if resp == nil {
			return EvaluationResult{IsHealthy: false, Reason: []string{"Status Code: No response received"}, Type: NotificationHttpError}
		}
		isMatch := resp.StatusCode == c.StatusCode.Code
		reason := fmt.Sprintf("Status Code: Expected %d, Got %d", c.StatusCode.Code, resp.StatusCode)
		if !isMatch {
			return EvaluationResult{
				IsHealthy: false,
				Reason:    []string{reason},
				Type:      NotificationHttpError,
			}
		}
		return EvaluationResult{IsHealthy: true, Reason: []string{reason}}
	}

	// Header
	if c.Header != nil {
		if resp == nil {
			return EvaluationResult{IsHealthy: false, Reason: []string{"Header: No response headers available"}, Type: NotificationConditionFailed}
		}
		var failures []string
		var successes []string
		for _, h := range *c.Header {
			actual := resp.Header.Get(h.Key)
			match := actual == h.Value
			reason := fmt.Sprintf("Header [%s]: Expected '%s', Got '%s'", h.Key, h.Value, actual)
			if !match {
				failures = append(failures, reason)
			} else {
				successes = append(successes, reason)
			}
		}
		if len(failures) > 0 {
			return EvaluationResult{IsHealthy: false, Reason: failures, Type: NotificationConditionFailed}
		}
		return EvaluationResult{IsHealthy: true, Reason: successes}
	}

	// Response Time
	if c.ResponseTime != nil {
		max, _ := time.ParseDuration(c.ResponseTime.MaxDuration)
		isMatch := duration <= max
		reason := fmt.Sprintf("Latency: Max %s, Actual %v", c.ResponseTime.MaxDuration, duration.Round(time.Millisecond))
		if !isMatch {
			return EvaluationResult{IsHealthy: false, Reason: []string{reason}, Type: NotificationSlowResponse}
		}
		return EvaluationResult{IsHealthy: true, Reason: []string{reason}}
	}

	return EvaluationResult{IsHealthy: false, Reason: []string{"No valid condition defined"}, Type: NotificationDefault}
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
