package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

type LoginRegisterMessage struct {
	UserID uuid.UUID `json:"user_id"`
	JWT    string    `json:"jwt"`
}

func NewRabbitMQ(rabbitMQURL string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (r *RabbitMQ) PublishLoginRegisterMessage(queueName string, userID uuid.UUID, jwt string) error {
	message := LoginRegisterMessage{
		UserID: userID,
		JWT:    jwt,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = r.Channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to queue %s: %s", queueName, body)
	return nil
}

func (r *RabbitMQ) ConsumeLoginRegisterMessages(queueName string, handler func(LoginRegisterMessage)) error {
	_, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := r.Channel.Consume(
		queueName,
		"",
		true,  
		false, 
		false, 
		false, 
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			var message LoginRegisterMessage
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			handler(message)
		}
	}()

	log.Printf("Started consuming messages from queue %s", queueName)
	return nil
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.Connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

