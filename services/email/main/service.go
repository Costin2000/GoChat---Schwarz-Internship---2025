package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type Config struct {
	RabbitMQAddr string
	SmtpHost     string
	SmtpPort     string
	SmtpUser     string
	SmtpPass     string
}

func main() {
	cfg := Config{
		RabbitMQAddr: os.Getenv("RABBITMQ_ADDR"),
		SmtpHost:     os.Getenv("SMTP_HOST"),
		SmtpPort:     os.Getenv("SMTP_PORT"),
		SmtpUser:     os.Getenv("SMTP_USER"),
		SmtpPass:     os.Getenv("SMTP_PASS"),
	}
	if cfg.RabbitMQAddr == "" || cfg.SmtpHost == "" || cfg.SmtpPort == "" || cfg.SmtpUser == "" || cfg.SmtpPass == "" {
		log.Fatal("Error: One or more environment variables (RABBITMQ_ADDR, SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS) are not set.")
	}

	conn, err := connectToRabbitMQWithRetries(cfg.RabbitMQAddr)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after several retries: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("emails_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// start consuming messages from the queue
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println(" [*] Waiting for email messages. To exit press CTRL+C")

	var forever chan struct{}

	// goroutine to process incoming messages concurrently
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var emailMsg EmailMessage
			if err := json.Unmarshal(d.Body, &emailMsg); err != nil {
				log.Printf("Error decoding JSON: %s", err)
				continue
			}
			if err := sendEmail(cfg, emailMsg); err != nil {
				log.Printf("Failed to send email to %s: %v", emailMsg.To, err)
			} else {
				log.Printf("Successfully sent email to %s", emailMsg.To)
			}
		}
	}()

	<-forever
}

func sendEmail(cfg Config, msg EmailMessage) error {
	auth := smtp.PlainAuth("", cfg.SmtpUser, cfg.SmtpPass, cfg.SmtpHost)
	smtpAddr := fmt.Sprintf("%s:%s", cfg.SmtpHost, cfg.SmtpPort)

	emailBody := "From: " + cfg.SmtpUser + "\r\n" +
		"To: " + msg.To + "\r\n" +
		"Subject: " + msg.Subject + "\r\n" +
		"\r\n" +
		msg.Body

	err := smtp.SendMail(smtpAddr, auth, cfg.SmtpUser, []string{msg.To}, []byte(emailBody))
	return err
}

// attempts to connect to RabbitMQ, retrying several times on failure
func connectToRabbitMQWithRetries(addr string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	maxRetries := 5
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(addr)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return conn, nil
		}
		log.Printf("Could not connect to RabbitMQ, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	return nil, err
}
