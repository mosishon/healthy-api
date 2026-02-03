package loader

import (
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
	"log/slog"
	"net/http"
	"time"
)

func LoadNotifiers(cfg *model.Config, reg *registry.Registry[notifier.Notifier], logger *slog.Logger) map[string]int {
	counts := make(map[string]int)

	counts["ippanel"] = loadIPPanelNotifiers(cfg, reg, logger)
	counts["meli_payamak"] = loadPayamakPanels(cfg, reg, logger)
	counts["smtp"] = loadSMTPNotifiers(cfg, reg, logger)
	counts["webhook"] = loadWebhookNotifiers(cfg, reg, logger)

	return counts
}

func loadIPPanelNotifiers(cfg *model.Config, reg *registry.Registry[notifier.Notifier], logger *slog.Logger) int {
	count := 0
	for _, ippanel := range cfg.Notifiers.IPPanels {
		if _, ok := reg.Get(ippanel.ID); ok {
			logger.Error("notifier_already_exists", "id", ippanel.ID, "type", "ippanel")
			continue
		}
		notifierInst := &notifier.SMSNotifier{
			User:   ippanel.User,
			Pass:   ippanel.Pass,
			URL:    ippanel.Url,
			Logger: logger,
		}
		reg.Register(ippanel.ID, notifierInst)
		logger.Info("notifier_registered", "type", "ippanel", "id", ippanel.ID)
		count++
	}
	return count
}

func loadPayamakPanels(cfg *model.Config, reg *registry.Registry[notifier.Notifier], logger *slog.Logger) int {
	count := 0
	for _, pp := range cfg.Notifiers.MeliPayamakPanels {
		if _, ok := reg.Get(pp.ID); ok {
			logger.Error("notifier_already_exists", "id", pp.ID, "type", "meli_payamak")
			continue
		}
		notifierInst := &notifier.PayamakNotifier{
			Username:  pp.Username,
			Password:  pp.Password,
			Sender:    pp.Sender,
			Template:  pp.Template,
			Templates: pp.Templates,
			Logger:    logger,
		}
		reg.Register(pp.ID, notifierInst)
		logger.Info("notifier_registered", "type", "meli_payamak", "id", pp.ID)
		count++
	}
	return count
}

func loadSMTPNotifiers(cfg *model.Config, reg *registry.Registry[notifier.Notifier], logger *slog.Logger) int {
	count := 0
	for _, smtp := range cfg.Notifiers.SMTPs {
		if _, ok := reg.Get(smtp.ID); ok {
			logger.Error("notifier_already_exists", "id", smtp.ID, "type", "smtp")
			continue
		}
		notifierInst := &notifier.MailNotifier{
			Sender:    smtp.Sender,
			Server:    smtp.Server,
			Port:      smtp.Port,
			Password:  smtp.Password,
			Templates: smtp.Templates,
			Logger:    logger,
		}
		reg.Register(smtp.ID, notifierInst)
		logger.Info("notifier_registered", "type", "smtp", "id", smtp.ID)
		count++
	}
	return count
}

func loadWebhookNotifiers(cfg *model.Config, reg *registry.Registry[notifier.Notifier], logger *slog.Logger) int {
	count := 0
	for _, wh := range cfg.Notifiers.Webhook {
		if _, ok := reg.Get(wh.ID); ok {
			logger.Error("notifier_already_exists", "id", wh.ID, "type", "webhook")
			continue
		}
		if err := checkTemplate(wh.JSON); err != nil {
			logger.Error("invalid_json_template", "id", wh.ID, "error", err)
			continue
		}
		if err := checkTemplate(wh.Headers); err != nil {
			logger.Error("invalid_headers_template", "id", wh.ID, "error", err)
			continue
		}
		notifierInst := &notifier.WebhookNotifier{
			HookData: wh,
			Client:   &http.Client{Timeout: time.Second * 15},
			Logger:   logger,
		}
		reg.Register(wh.ID, notifierInst)
		logger.Info("notifier_registered", "type", "webhook", "id", wh.ID)
		count++
	}
	return count
}

func checkTemplate(templ map[string]interface{}) error {
	_, err := notifier.FillTemplate(templ, model.WebhookTemplate{
		Metadata: model.NotificationMetadata{
			ServiceName: "test",
			Timestamp:   "Test",
		},
		URL: "test",
	})
	return err
}
