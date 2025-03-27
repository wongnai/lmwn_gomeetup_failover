package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
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
	ctx, cancelConsumer := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	StartConsumer(ctx, wg, rmq, svc)

	// Step 3: Start Health Check Server
	health.RunHealthCheck(mongo, rmq)

	// Step 4: Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cancelConsumer() // Cancel the context to stop consumers
	wg.Wait()        // Wait for all goroutines to finish

	rmq.Stop(shutdownCtx)
	svc.Shutdown(shutdownCtx)
	mongo.Close(shutdownCtx)

	log.Println("Consumer shutdown complete.")
}

func StartConsumer(ctx context.Context, wg *sync.WaitGroup, rmq *queue.RabbitMQ, svc *service.Service) {
	deliveries, err := rmq.GetConsumerChannel()
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in consumer: %v", r)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping consumer...")
				return
			case msg, ok := <-deliveries:
				if !ok {
					log.Println("Delivery channel closed, attempting reconnect...")
					// Try reconnecting only if context is not done
					select {
					case <-ctx.Done():
						log.Println("Shutdown in progress, skipping reconnect.")
						return
					default:
						rmq.Reconnect()
						StartConsumer(ctx, wg, rmq, svc)
						return
					}
				}

				log.Printf("Received message: %s", msg.Body)

				if svc.IsMessageProcessed(string(msg.Body)) {
					log.Println("Skipping duplicate message")
					msg.Ack(false)
					continue
				}

				if err := svc.ProcessMessage(ctx, string(msg.Body)); err != nil {
					log.Printf("Error processing message: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	log.Print("Consumer started successfully")
}
