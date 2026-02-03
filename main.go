package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"healthy-api/config"
	"healthy-api/healthcheck"
	"healthy-api/loader"
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
)

var (
	configPath string
	verbose    bool
)

func init() {
	flag.StringVar(&configPath, "config", "", "Path to the configurations file.")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging.")
}

func main() {
	flag.Parse()

	if configPath == "" {
		fmt.Println("🚨 Missing required flag: -config")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, logFile := setupLogger(verbose)
	if logFile != nil {
		defer logFile.Close()
	}
	slog.SetDefault(logger)

	notifierRegistry := registry.NewRegistry[notifier.Notifier]()
	conditionRegistry := registry.NewRegistry[model.Condition]()

	notifierCounts := loader.LoadNotifiers(cfg, notifierRegistry, logger)
	conditionCount := loader.LoadConditions(cfg, conditionRegistry, logger)

	printSummary(notifierCounts, conditionCount, len(cfg.Services))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	for _, svc := range cfg.Services {
		if !validateService(svc, notifierRegistry, conditionRegistry, logger) {
			continue
		}

		hc := &healthcheck.HealthChecker{
			Service:           svc,
			NotifierRegistry:  notifierRegistry,
			ConditionRegistry: conditionRegistry,
			Client: &http.Client{
				Timeout: 15 * time.Second,
			},
			Logger: logger,
		}

		wg.Add(1)
		go func(s model.Service) {
			defer wg.Done()
			runHealthCheck(ctx, hc)
			logger.Info("checker_stopped", "service", s.Name, "url", s.URL)
		}(svc)
	}

	<-ctx.Done()
	logger.Info("shutting_down", "message", "waiting for workers to finish...")

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("shutdown_complete")
	case <-time.After(10 * time.Second):
		logger.Warn("shutdown_timeout", "message", "some workers did not stop in time")
	}
}

func setupLogger(verbose bool) (*slog.Logger, *os.File) {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	var logOutput io.Writer
	if verbose {
		logOutput = io.MultiWriter(os.Stdout, logFile)
	} else {
		logOutput = logFile
	}

	handler := slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler), logFile
}

func validateService(svc model.Service, nr *registry.Registry[notifier.Notifier], cr *registry.Registry[model.Condition], logger *slog.Logger) bool {
	if _, ok := cr.Get(svc.ConditionName); !ok {
		logger.Error("condition_not_found", "service", svc.Name, "condition_id", svc.ConditionName)
		return false
	}

	for _, target := range svc.Targets {
		if _, ok := nr.Get(target.NotifierID); !ok {
			logger.Error("notifier_not_found", "service", svc.Name, "notifier_id", target.NotifierID)
			return false
		}
	}
	return true
}

func printSummary(notifierCounts map[string]int, conditionCount int, serviceCount int) {
	fmt.Println()
	fmt.Println("--------- CONFIG SUMMARY -----------")
	for t, c := range notifierCounts {
		fmt.Printf("%d %s registered.\n", c, t)
	}
	fmt.Printf("%d conditions found.\n", conditionCount)
	fmt.Printf("%d services found.\n", serviceCount)
	fmt.Println("------------------------------------")
	fmt.Println()
}

func runHealthCheck(ctx context.Context, hc *healthcheck.HealthChecker) {
	hc.Start(ctx)
}
