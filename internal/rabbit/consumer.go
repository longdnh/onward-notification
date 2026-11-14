package rabbit

import (
	"errors"
	"fmt"
	"log"

	"github.com/longdnh/onward-notification/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

func Run(cfg config.Config) error {

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

	for d := range messages {
		log.Printf("Received a message: %s", d.Body)
		log.Printf("received messageId=%s type=%s", d.MessageId, d.Type)
		err := d.Ack(false)
		if err != nil {
			log.Printf("ack failed: %v", err)
		}
	}

	return errors.New("delivery channel closed, connection lost")
}
