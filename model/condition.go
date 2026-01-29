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
func (c *Condition) Evaluate(resp *http.Response, body []byte,duration time.Duration) bool {
	if c.And != nil {
		for _, cond := range c.And {
			if !cond.Evaluate(resp, body,duration) {
				return false
			}
		}
		return true
	}
	if c.Or != nil {

		for _, cond := range c.Or {
			if cond.Evaluate(resp, body,duration) {
				return true
			}
		}
		return false
	}
	if c.Not != nil {
		return !c.Not.Evaluate(resp, body,duration)
	}
	if c.Regex != nil {
		return c.Regex.Evaluate(body)
	}
	if c.StatusCode != nil {
		return c.StatusCode.Evaluate(resp)
	}
	if c.Header != nil {
		for _, h := range *c.Header {
			if matched := h.Evaluate(resp); !matched { 
                return false
            }
		}
		return true
	}
	if c.ResponseTime != nil {
		return c.ResponseTime.Evaluate(duration)
	}
	return false
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
