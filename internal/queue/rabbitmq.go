package queue

import (
	"context"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const queueName = "queue_name"

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ() (*RabbitMQ, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Ensure queue exists before consuming
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{Conn: conn, Channel: ch}, nil
}

func (r *RabbitMQ) GetConsumerChannel() (<-chan amqp.Delivery, error) {
	deliveries, err := r.Channel.Consume(queueName, "consumer", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return deliveries, nil

}

func (r *RabbitMQ) Reconnect() {
	for {
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err == nil {
			r.Conn = conn
			r.Channel, err = conn.Channel()
			if err == nil {
				log.Println("RabbitMQ reconnected successfully")
				return
			}
		}
		log.Println("RabbitMQ reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

func (e *RabbitMQ) Publish(routingKey, msgID, eventName string, payload []byte, ref map[string]string) (err error) {
	if err := e.ensureProducer(); err != nil {
		return err
	}

	msg := amqp.Publishing{
		Body: payload,
	}
	return e.Channel.Publish("exchange", "key", true, false, msg)
}

func (e *RabbitMQ) ensureProducer() (err error) {
	if !e.Conn.IsClosed() {
		return nil
	}
	if e.Conn.IsClosed() {
		e.Conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			return err
		}
	}
	e.Channel, err = e.Conn.Channel()
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) Stop(ctx context.Context) {
	log.Println("Shutting down RabbitMQ...")
	done := make(chan struct{})
	go func() {
		if r.Channel != nil {
			r.Channel.Close()
		}
		if r.Conn != nil {
			r.Conn.Close()
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("RabbitMQ shutdown complete.")
	case <-ctx.Done():
		log.Println("RabbitMQ shutdown timeout exceeded, forcing stop.")
	}
}

func (r *RabbitMQ) IsConnected() bool {
	if r.Conn == nil || r.Conn.IsClosed() {
		log.Println("RabbitMQ connection is closed")
		return false
	}
	return true
}
