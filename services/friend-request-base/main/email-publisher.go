package main

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const emailsQueueName = "emails_queue"

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailPublisher interface {
	Publish(ctx context.Context, msg EmailMessage) error
	Close() error
}

type amqpEmailPublisher struct {
	ch    *amqp.Channel
	queue string
}

// building AMQP publisher from connection
func newAmqpEmailPublisher(conn *amqp.Connection) (EmailPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare(emailsQueueName, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return nil, err
	}
	return &amqpEmailPublisher{
		ch:    ch,
		queue: emailsQueueName,
	}, nil
}

// publish email message to queue
func (p *amqpEmailPublisher) Publish(ctx context.Context, msg EmailMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(
		ctx,
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (p *amqpEmailPublisher) Close() error {
	if p.ch != nil {
		return p.ch.Close()
	}
	return nil
}
