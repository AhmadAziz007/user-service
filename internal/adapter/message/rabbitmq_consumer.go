package message

import (
	"context"
	"ecommerce/user-service/config"
	"ecommerce/user-service/internal/adapter/mailer"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const emailVerificationQueue = "email_verification"
const forgotPasswordQueue = "forgot-password"

type emailNotification struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

// StartEmailVerificationConsumer sends verification and password-reset messages
// from RabbitMQ through SMTP.
// Call the returned function during application shutdown.
func StartEmailVerificationConsumer(ctx context.Context, cfg *config.Config) (func() error, error) {
	conn, err := cfg.NewRabbitMQ()
	if err != nil {
		return nil, fmt.Errorf("connect RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open RabbitMQ channel: %w", err)
	}

	if err = channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("configure RabbitMQ QoS: %w", err)
	}

	queues := []struct {
		name    string
		subject string
	}{
		{name: emailVerificationQueue, subject: "Verify your account"},
		{name: forgotPasswordQueue, subject: "Reset your password"},
	}

	for _, queue := range queues {
		if _, err = channel.QueueDeclare(queue.name, true, false, false, false, nil); err != nil {
			_ = channel.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("declare %q queue: %w", queue.name, err)
		}

		consumerTag := "user-service-" + queue.name
		deliveries, err := channel.Consume(queue.name, consumerTag, false, false, false, false, nil)
		if err != nil {
			_ = channel.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("consume %q queue: %w", queue.name, err)
		}

		go consumeEmailNotifications(ctx, deliveries, cfg, queue.subject)
	}

	return func() error {
		for _, queue := range queues {
			_ = channel.Cancel("user-service-"+queue.name, false)
		}
		channelErr := channel.Close()
		connectionErr := conn.Close()
		if channelErr != nil {
			return channelErr
		}
		return connectionErr
	}, nil
}

func consumeEmailNotifications(ctx context.Context, deliveries <-chan amqp.Delivery, cfg *config.Config, subject string) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}

			var notification emailNotification
			if err := json.Unmarshal(delivery.Body, &notification); err != nil || notification.Email == "" || notification.Message == "" {
				log.Printf("[EmailConsumer] invalid message: %v", err)
				_ = delivery.Nack(false, false)
				continue
			}

			if err := mailer.SendEmail(cfg, notification.Email, subject, notification.Message); err != nil {
				log.Printf("[EmailConsumer] send email to %s: %v", notification.Email, err)
				_ = delivery.Nack(false, true)
				continue
			}

			if err := delivery.Ack(false); err != nil {
				log.Printf("[EmailConsumer] acknowledge message: %v", err)
			}
		}
	}
}
