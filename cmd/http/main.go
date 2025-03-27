package main

import (
	"context"
	"lmwn_gomeetup_failover/internal/db"
	"lmwn_gomeetup_failover/internal/health"
	"lmwn_gomeetup_failover/internal/http"
	"lmwn_gomeetup_failover/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Step 1: Initialize dependencies
	mongo, err := db.NewMongoDB()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	svc := service.NewService(mongo)
	httpServer := http.NewHTTPServer(svc)

	// Step 2: Start main business logic (HTTP Server)
	go httpServer.Start()

	// Step 3: Start Health Check Server
	health.RunHealthCheck(mongo, nil)

	// Step 4: Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done() // Wait for termination signal
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpServer.Stop(shutdownCtx)
	svc.Shutdown(shutdownCtx)
	mongo.Close(shutdownCtx)
}
