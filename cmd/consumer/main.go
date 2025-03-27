package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lmwn_gomeetup_failover/internal/db"
	"lmwn_gomeetup_failover/internal/health"
	"lmwn_gomeetup_failover/internal/queue"
	"lmwn_gomeetup_failover/internal/service"
)

func main() {
	// Step 1: Initialize dependencies
	mongo, err := db.NewMongoDB()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	svc := service.NewService(mongo)
	rmq, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}

	// Step 2: Start main business logic (HTTP Server)
	StartConsumer(rmq, svc)

	// Step 3: Start Health Check Server
	health.RunHealthCheck(mongo, rmq)

	// Step 4: Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rmq.Stop(shutdownCtx)
	svc.Shutdown(shutdownCtx)
	mongo.Close(shutdownCtx)

	log.Println("Consumer shutdown complete.")
}

func StartConsumer(rmq *queue.RabbitMQ, svc *service.Service) {
	deliveries, err := rmq.GetConsumerChannel()
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in consumer: %v", r)
			}
		}()
		for msg := range deliveries {
			log.Printf("Received message: %s", msg.Body)
			if svc.IsMessageProcessed(string(msg.Body)) { // Check if already processed
				log.Println("Skipping duplicate message")
				msg.Ack(false)
				continue
			}

			if err := svc.ProcessMessage(context.Background(), string(msg.Body)); err != nil {
				log.Printf("Error processing message: %v", err)
				msg.Nack(false, true) // Requeue the message in case of failure
			} else {
				msg.Ack(false) // Acknowledge successful processing
			}
		}

		log.Println("Consumer stopped receiving messages, attempting reconnect...")
		rmq.Reconnect()
		StartConsumer(rmq, svc)
	}()

	log.Print("start consumer successful")
}
