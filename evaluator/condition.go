package conditionevaluator

import (
	"net/http"
)

type Evaluator interface {
	Evaluate(resp *http.Response, body []byte) bool
}
