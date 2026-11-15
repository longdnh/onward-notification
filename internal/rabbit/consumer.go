package rabbit

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/longdnh/onward-notification/internal/config"
	"github.com/longdnh/onward-notification/internal/event"
	"github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Handle(ctx context.Context, m event.TokenMail) error
}

type Dedup interface {
	Claim(ctx context.Context, id string) (bool, error) // write id if not exist, return false if id existed
	Release(ctx context.Context, id string) error // release id, use when send mail fail, to flag the resend is not duplication
}

func Run(cfg config.Config, h Handler, dd Dedup) error {
	// create a connection
	conn, err := amqp091.Dial(cfg.RabbitURL)
	if err != nil {
		return fmt.Errorf("Fail Create Rabbitmq connection: %w", err)
	}
	defer conn.Close()

	// open a channel
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Fail to create a channel: %w", err)
	}
	defer ch.Close()

	// Qos
	err = ch.Qos(cfg.Prefetch, 0, false)
	if err != nil {
		return fmt.Errorf("fail to Qos: %w", err)
	}

	// queue declaration
	queue, err := ch.QueueDeclare(
		cfg.Queue,
		true,
		false,
		false,
		false,
		amqp091.Table{
			amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
		},
	)
	if err != nil {
		return fmt.Errorf("Fail to declare a queue: %w", err)
	}

	// bind
	err = ch.QueueBind(cfg.Queue, "auth.#", cfg.Exchange, false, nil)
	if err != nil {
		return fmt.Errorf("Fail to bind: %w", err)
	}

	// consume
	messages, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Fail to consume: %w", err)
	}

	ctx := context.Background()
	for d := range messages {
		log.Printf("received messageId=%s type=%s", d.MessageId, d.Type)

		msg, err := event.Decode(d.Body)
		if err != nil {
			log.Printf("bad message %s: %v", d.MessageId, err)
			if err := d.Ack(false); err != nil {
				return fmt.Errorf("ack %s: %w", d.MessageId, err)
			}
			continue
		}
		log.Printf("email=%s expiresAt=%s", msg.Email, msg.ExpiresAt)

		// Handle duplicate
		ok, err := dd.Claim(ctx, msg.MessageID)
		if err != nil {
			return fmt.Errorf("claim %s: %w", msg.MessageID, err)
		}

		if !ok {
			log.Printf("Duplicated Message, MessageId: %s", msg.MessageID)
			if err := d.Ack(false); err != nil {
				return fmt.Errorf("ack %s: %w", d.MessageId, err)
			}
			continue
		}

		// Handle
		if err := h.Handle(ctx, msg); err != nil {
			if rerr := dd.Release(ctx, msg.MessageID); rerr != nil {
				log.Printf("release %s: %v", msg.MessageID, rerr)
			}
			return fmt.Errorf("handle %s: %w", d.MessageId, err)
		}

		if err := d.Ack(false); err != nil {
			return fmt.Errorf("ack %s: %w", d.MessageId, err)
		}
	}

	return errors.New("delivery channel closed, connection lost")
}
