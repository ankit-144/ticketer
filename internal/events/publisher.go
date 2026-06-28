package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// GenericEventEnvelope represents the standard payload sent to the Notification Service
type GenericEventEnvelope struct {
	UserID        string                 `json:"userId"`
	EventType     string                 `json:"eventType"`
	SourceService string                 `json:"sourceService"`
	Timestamp     string                 `json:"timestamp"`
	TemplateData  map[string]interface{} `json:"templateData"`
}

type EventPublisher interface {
	PublishEvent(ctx context.Context, event GenericEventEnvelope) error
	Close() error
}

type RabbitMQPublisher struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func NewRabbitMQPublisher(amqpURI, exchange, routingKey string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQPublisher{
		conn:       conn,
		channel:    ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *RabbitMQPublisher) PublishEvent(ctx context.Context, event GenericEventEnvelope) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.PublishWithContext(ctx,
		p.exchange,   // exchange
		p.routingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		})

	if err != nil {
		// TODO (Future): Implement Transactional Outbox pattern here.
		// Instead of dropping the event, we should save it to the DB in the same 
		// transaction as the core business logic, and a background worker publishes it.
		log.Printf("ERROR: Failed to publish message to RabbitMQ: %v", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Successfully published event %s to exchange %s", event.EventType, p.exchange)
	return nil
}

func (p *RabbitMQPublisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}
