package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: data})
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	var durable, autoDelete, exclusive bool
	if simpleQueueType == 0 { // "durable"
		durable = true
		autoDelete = false
		exclusive = false
	} else { // "transient"
		durable = false
		autoDelete = true
		exclusive = true
	}

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(T)) error {
	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			var msg T
			err = json.Unmarshal(delivery.Body, &msg)
			if err != nil {
				log.Fatal("Error unmarshaling message: ", err)
			}

			handler(msg)
			err = delivery.Ack(false)
			if err != nil {
				log.Fatal("Error acknowledging message: ", err)
			}
		}
	}()

	return nil
}
