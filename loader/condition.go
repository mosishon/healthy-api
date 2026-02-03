package loader

import (
	"healthy-api/model"
	"healthy-api/registry"
	"log/slog"
)

func LoadConditions(cfg *model.Config, reg *registry.Registry[model.Condition], logger *slog.Logger) int {
	count := 0
	for _, cond := range cfg.Conditions {
		if _, ok := reg.Get(cond.ID); ok {
			logger.Error("condition_already_exists", "id", cond.ID)
			continue
		}
		if err := cond.Condition.Validate("conditions.condition"); err != nil {
			logger.Error("invalid_condition", "id", cond.ID, "error", err)
			continue
		}
		reg.Register(cond.ID, *cond.Condition)
		logger.Info("condition_registered", "id", cond.ID)
		count++
	}
	return count
}
