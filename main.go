package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"healthy-api/config"
	"healthy-api/healthcheck"
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
)

var configPath string
var verbose bool

func init() {
	flag.StringVar(&configPath, "config", "", "Path to the configurations file.")
	flag.BoolVar(&verbose, "verbose", false, "showing logs or no.")
}

func loadPayamakPanels(cfg *model.Config, notifierRegistry *registry.Registry[notifier.Notifier], logger *log.Logger) int {
	payamakCount := 0
	for _, pp := range cfg.Notifiers.MeliPayamakPanels {
		_, ok := notifierRegistry.Get(pp.ID)
		if ok {
			log.Fatalf("notifier with name %s already exists", pp.ID)
		}

		notifierInst := &notifier.PayamakNotifier{
			Username: pp.Username,
			Password: pp.Password,
			Sender:   pp.Sender,
			Template: pp.Template,
			Logger:   logger,
		}
		
		payamakCount++
		notifierRegistry.Register(pp.ID, notifierInst)
	}
	return payamakCount
}

func loadIPPanelNotifiers(cfg *model.Config, notifierRegistry *registry.Registry[notifier.Notifier], logger *log.Logger) int {
	ippanelCount := 0
	for _, ippanel := range cfg.Notifiers.IPPanels {
		_, ok := notifierRegistry.Get(ippanel.ID)
		if ok == true {
			log.Fatalf("notifier with name %s already exists", ippanel.ID)
		}
		notifierInst := &notifier.SMSNotifier{
			User:   ippanel.User,
			Pass:   ippanel.Pass,
			URL:    ippanel.Url,
			Logger: logger,
		}
		ippanelCount++

		notifierRegistry.Register(ippanel.ID, notifierInst)
		logger.Printf("new notifier registered. type:ippanel -> %v\n", notifierInst)

	}
	return ippanelCount
}

func loadSMTPNotifiers(cfg *model.Config, notifierRegistry *registry.Registry[notifier.Notifier], logger *log.Logger) int {
	smtpCount := 0

	for _, smtp := range cfg.Notifiers.SMTPs {
		_, ok := notifierRegistry.Get(smtp.ID)
		if ok == true {
			log.Fatalf("notifier with name %s already exists", smtp.ID)
		}
		notifierInst := &notifier.MailNotifier{
			Sender:   smtp.Sender,
			Server:   smtp.Server,
			Port:     smtp.Port,
			Password: smtp.Password,
			Logger:   logger,
		}
		smtpCount++

		notifierRegistry.Register(smtp.ID, notifierInst)
		logger.Printf("new notifier registered. type:smtp -> %v\n", notifierInst)

	}
	return smtpCount
}

func checkTemplate(templ map[string]interface{}) error {
	_, err := notifier.FillTemplate(templ, model.WebhookTemplate{
		ServiceName: "test",
		TimeStamp:   "Test",
		URL:         "test",
	})
	return err

}
func loadWebhookNotifiers(cfg *model.Config, notifierRegistry *registry.Registry[notifier.Notifier], logger *log.Logger) int {
	whCount := 0
	for _, wh := range cfg.Notifiers.Webhook {
		_, ok := notifierRegistry.Get(wh.ID)
		if ok == true {
			log.Fatalf("notifier with name %s already exists", wh.ID)
		}
		err := checkTemplate(wh.JSON)
		if err != nil {
			logger.Fatalf("[INVALID TEMPLATE] json template for webhook '%s' is not valid. details:\n%v", wh.ID, err)
		}
		err = checkTemplate(wh.Headers)
		if err != nil {
			logger.Fatalf("[INVALID TEMPLATE] headers template for webhook '%s' is not valid. details:\n%v", wh.ID, err)
		}
		notifierInst := &notifier.WebhookNotifier{
			HookData: wh,
			Client:   &http.Client{Timeout: time.Second * 15},
			Logger:   logger,
		}
		whCount++
		notifierRegistry.Register(wh.ID, notifierInst)
		logger.Printf("new notifier registered. type:webhook -> %v\n", notifierInst)

	}
	return whCount
}

func PrintCondition(cond *model.Condition) {
	bytes, err := json.MarshalIndent(cond, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling condition:", err)
		return
	}
	fmt.Println(string(bytes))
}

func loadConditions(cfg *model.Config, conditionRegistry *registry.Registry[model.Condition], logger *log.Logger) int {
	cCound := 0
	for _, cond := range cfg.Conditions {

		_, ok := conditionRegistry.Get(cond.ID)
		if ok == true {
			logger.Fatalf("condition with id %s already exists", cond.ID)
		}
		if err := cond.Condition.Validate("conditions.condition"); err != nil {
			logger.Fatalf("error condition is not valid %s", err)
		}
		cCound++
		conditionRegistry.Register(cond.ID, *cond.Condition)

	}
	return cCound
}

// TODO We need gracefull shutdown for goroutines.
func main() {

	flag.Parse()
	if configPath == "" {
		fmt.Println("🚨 Missing required flag: -config")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
	var wg sync.WaitGroup
	println("Reading config file.")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	notifierRegistry := registry.NewRegistry[notifier.Notifier]()
	conditionRegistry := registry.NewRegistry[model.Condition]()
	logger := log.Default()
	if verbose == false {
		logger.SetOutput(io.Discard)
	}
	ippanelCount := loadIPPanelNotifiers(cfg, notifierRegistry, logger)
	meliPayamakCount := loadPayamakPanels(cfg, notifierRegistry, logger) 
	smtpCount := loadSMTPNotifiers(cfg, notifierRegistry, logger)
	whCount := loadWebhookNotifiers(cfg, notifierRegistry, logger)
	fmt.Println()
	fmt.Println("---------NOTIFIERS-----------")
	fmt.Printf("%d ippanel regisered.\n", ippanelCount)
	fmt.Printf("%d meli_payamak_panel registered.\n", meliPayamakCount)
	fmt.Printf("%d smtp regisered.\n", smtpCount)
	fmt.Printf("%d webhook regisered.\n", whCount)
	fmt.Println("---------NOTIFIERS-----------")
	fmt.Println()
	cCount := loadConditions(cfg, conditionRegistry, logger)
	fmt.Printf("%d condition found.\n\n", cCount)

	fmt.Printf("%d service found.\n\n", len(cfg.Services))
	for n, svc := range cfg.Services {
		n++
		fmt.Printf("Service [%d]: %s\n", n, svc.Name)
		fmt.Println("  URL:", svc.URL)
		fmt.Println("  Period:", svc.CheckPeriod)
		fmt.Println("  Condition id:", svc.ConditionName)
		fmt.Println("  SleepOnFail:", svc.SleepOnFail)
		fmt.Println("  Targets count:", len(svc.Targets))
		fmt.Println("  User-Agent:", svc.UserAgent)
		fmt.Println("  Threshold:", svc.Threshold)
	
		fmt.Println("----")
		for _, v := range svc.Targets {
			_, ok := notifierRegistry.Get(v.NotifierID)
			if ok == false {
				fmt.Printf("\n\n[ERROR] notifier with id: '%s' not found.for service: `%s`\n\n\n", v.NotifierID, svc.Name)
				os.Exit(1)
			}

		}

		hc := healthcheck.HealthChecker{
			Service:           svc,
			NotifierRegistry:  notifierRegistry,
			ConditionRegistry: conditionRegistry,
			Client: &http.Client{
				Timeout: time.Duration(15) * time.Second,
			},
			Logger: logger,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			hc.Start()
			fmt.Printf("chcker for %s[%s] stopped", svc.Name, svc.URL)
		}()
	}

	println("Wating for all workers to finish their work.")
	wg.Wait()
}
