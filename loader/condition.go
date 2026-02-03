package loader

import (
	"fmt"
	"healthy-api/model"
	"healthy-api/registry"
	"log/slog"
)

func LoadConditions(cfg *model.Config, reg *registry.Registry[model.Condition], logger *slog.Logger) int {
	count := 0
	for _, cond := range cfg.Conditions {
		if cond.ID == "" {
			logger.Error("invalid_condition", "error", "missing condition ID")
			continue
		}
		if _, ok := reg.Get(cond.ID); ok {
			logger.Error("condition_already_exists", "id", cond.ID)
			continue
		}
		if cond.Condition == nil {
			logger.Error("invalid_condition", "id", cond.ID, "error", "condition body is missing (check YAML structure)")
			continue
		}
		if err := cond.Condition.Validate(fmt.Sprintf("conditions[%s].condition", cond.ID)); err != nil {
			logger.Error("invalid_condition", "id", cond.ID, "error", err)
			continue
		}
		reg.Register(cond.ID, *cond.Condition)
		logger.Info("condition_registered", "id", cond.ID)
		count++
	}
	return count
}
