package main

import (
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
	"healthy-api/notifier"
)

var configPath string
var verbose bool

func init() {
	flag.StringVar(&configPath, "config", "", "Path to the configurations file.")
	flag.BoolVar(&verbose, "verbose", false, "showing logs or no.")
}
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
	smsNotifier := notifier.SMSNotifier{
		User: cfg.User,
		Pass: cfg.Pass,
	}
	fmt.Printf("%d service found.\n\n", len(cfg.Services))
	for n, svc := range cfg.Services {
		n++
		fmt.Printf("Service [%d]: %s\n", n, svc.Name)
		fmt.Println("  URL:", svc.URL)
		fmt.Println("  Phones:", svc.Phones)
		fmt.Println("  Period:", svc.CheckPeriod)
		fmt.Println("  SleepOnFail:", svc.SleepOnFail)
		fmt.Println("----")
		logger := log.Default()

		if verbose == false {
			logger.SetOutput(io.Discard)
		}
		hc := healthcheck.HealthChecker{
			Service:  svc,
			Notifier: smsNotifier,
			Client: &http.Client{
				Timeout: time.Duration(10) * time.Second,
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

	// err = smsNotifier.Notif(model.Notification{
	// 	ServiceName: "TEST",
	// 	Recipients:  []string{"+989164230882"},
	// })
	// fmt.Println(err)
	println("Wating for all workers to finish their work.")
	wg.Wait()
}
