package main

import (
	"context"
	"lmwn_gomeetup_failover/internal/db"
	"lmwn_gomeetup_failover/internal/health"
	"lmwn_gomeetup_failover/internal/service"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func runCronJob(ctx context.Context, wg *sync.WaitGroup, interval time.Duration, svc *service.Service) {
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in cron job: %v", r)
			}
		}()
		defer wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Cron job shutting down")
				return
			case <-ticker.C:
				log.Println("Executing scheduled task...")
				if svc.ShouldProcessTask() { // ✅ Ensure idempotency before execution
					performTask(svc)
				} else {
					log.Println("Skipping task: Already processed or in progress")
				}

			}
		}
	}()
}

func performTask(svc *service.Service) {
	// Simulate the actual cron job task
	log.Println("Running cron task logic...")

	// Add actual business logic here
	err := svc.DoSomething(context.Background())
	if err != nil {
		log.Printf("DoSomething error: %v\n", err)
	}
	log.Println("Cron task completed")
}

func main() {
	// Step 1: Initialize dependencies
	mongo, err := db.NewMongoDB()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	svc := service.NewService(mongo)

	// Step 2: Start main business logic (HTTP Server)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	wg := &sync.WaitGroup{}
	interval := 10 * time.Second // Adjust interval as needed
	runCronJob(ctx, wg, interval, svc)

	// Step 3: Start Health Check Server
	health.RunHealthCheck(mongo, nil)

	// Step 4: Handle graceful shutdown
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Cron job shutdown complete.")
	case <-shutdownCtx.Done():
		log.Println("Cron shutdown timeout exceeded, forcing stop.")
	}
}
