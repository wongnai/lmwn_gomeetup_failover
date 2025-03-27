package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"lmwn_gomeetup_failover/internal/circuitbreaker"
	"lmwn_gomeetup_failover/internal/db"
	"lmwn_gomeetup_failover/internal/queue"
	"lmwn_gomeetup_failover/internal/retry"
	"lmwn_gomeetup_failover/internal/safe"
	"lmwn_gomeetup_failover/internal/workerpool"
)

type Service struct {
	wg         sync.WaitGroup
	shutdown   chan struct{}
	mongo      *db.MongoDB
	cb         *circuitbreaker.CircuitBreaker
	rmq        *queue.RabbitMQ
	workerPool *workerpool.WorkerPool
}

func NewService(mongo *db.MongoDB) *Service {
	mq, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("cannot init rabbitMQ %v", err)
	}
	return &Service{
		shutdown:   make(chan struct{}),
		mongo:      mongo,
		cb:         circuitbreaker.NewCircuitBreaker(),
		rmq:        mq,
		workerPool: workerpool.NewWorkerPool(5, 10), // 5 workers, queue size 10
	}
}

func (s *Service) sendNotification(orderID string) {
	defer s.wg.Done()
	log.Printf("Sending notification for order %s", orderID)
	time.Sleep(2 * time.Second)
	log.Printf("Notification sent for order %s", orderID)
}

func (s *Service) doPayment(orderID string) {
	defer s.wg.Done()
	log.Printf("Processing payment for order %s", orderID)
	time.Sleep(3 * time.Second)
	log.Printf("Payment completed for order %s", orderID)
}

func (s *Service) ProcessMessage(ctx context.Context, message string) error {
	log.Printf("Processing message: %s", message)
	// Example MongoDB interaction
	collection := s.mongo.Client.Database("exampleDB").Collection("orders")
	_, err := collection.InsertOne(ctx, map[string]string{"message": message})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) IsMessageProcessed(message string) bool {
	// Ensure idempotency before execution (e.g. check data state for the corresponding message)
	return true
}

func (s *Service) ShouldProcessTask() bool {
	// Ensure idempotency before execution (e.g. check data state, other ongoing cron process)
	return true
}

func (s *Service) DoSomething(ctx context.Context) error {
	return nil
}

func (s *Service) CallExternalAPIWithCircuitBreaker() error {
	_, err := s.cb.Execute(func() (interface{}, error) {
		// Simulating failure (Replace with actual API call)

		failed := time.Now().Unix()%2 == 0 // Simulate 50% failure rate
		if failed {
			return nil, errors.New("API request failed")
		}
		return "Success", nil
	})
	return err
}

func (s *Service) CallExternalAPIWithRetry() error {
	retry.RetryWithExponentialBackoff(func() error {
		// Simulating failure (Replace with actual API call)

		failed := time.Now().Unix()%2 == 0 // Simulate 50% failure rate
		if failed {
			return errors.New("API request failed")
		}
		return nil

	}, 3, 100*time.Millisecond)

	return nil
}

func (s *Service) BulkSendOrdersReminder(orderIDs []string) {
	for _, orderID := range orderIDs {
		s.workerPool.Submit(func() {
			s.sendNotification(orderID)
		})
	}
}

func (s *Service) BulkSendOrdersReminderWithSemaphore(orderIDs []string) {
	const maxConcurrentJobs = 5 // Limit concurrency to 5
	semaphore := make(chan struct{}, maxConcurrentJobs)
	var wg sync.WaitGroup

	for i, orderID := range orderIDs {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a slot

		go func(jobID int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the slot

			fmt.Printf("Processing job %d\n", jobID)
			s.sendNotification(orderID)
			fmt.Printf("Job %d done\n", jobID)
		}(i)
	}

	wg.Wait()
}

func (s *Service) CreateOrder(param string) (orderID string, err error) {
	log.Printf("Order %s created", param)

	// create order implementation
	orderID = uuid.NewString()

	s.wg.Add(1)
	safe.GoFunc(func() {
		s.doPayment(param)
	})

	s.wg.Add(1)
	safe.GoFunc(func() {
		s.sendNotification(param)
	})

	safe.GoFunc(func() {
		s.rmq.Publish("routingKey", "msgID", "eventName", []byte(fmt.Sprintf(`{"orderId":%s, "status":"created"}`, orderID)), nil)
	})

	return orderID, nil
}

func (s *Service) Shutdown(ctx context.Context) {
	log.Println("Shutting down service...")
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		s.workerPool.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Service shutdown complete.")
	case <-ctx.Done():
		log.Println("Service shutdown timeout exceeded, forcing exit.")
	}
}
