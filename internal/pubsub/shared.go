package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/gob", Body: buff.Bytes()})
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBindQueue(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
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

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func DeclareExchange(channel *amqp.Channel, name, kind string) error {
	err := channel.ExchangeDeclare(name, kind, true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(T) AckType) error {
	channel, _, err := DeclareAndBindQueue(conn, exchange, queueName, key, simpleQueueType)
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

			ack_type := handler(msg)
			switch ack_type {
			case Ack:
				err = delivery.Ack(false)
				break
			case NackRequeue:
				err = delivery.Nack(false, true)
				break
			case NackDiscard:
				err = delivery.Nack(false, false)
				break
			}
			if err != nil {
				log.Fatal("Error acknowledging message: ", err)
			}
		}
	}()

	return nil
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(T) AckType) error {
	channel, _, err := DeclareAndBindQueue(conn, exchange, queueName, key, simpleQueueType)
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
			buff := bytes.NewBuffer(delivery.Body)
			decoder := gob.NewDecoder(buff)
			err := decoder.Decode(&msg)
			if err != nil {
				log.Fatal("Error decoding message: ", err)
			}

			ack_type := handler(msg)
			switch ack_type {
			case Ack:
				err = delivery.Ack(false)
				break
			case NackRequeue:
				err = delivery.Nack(false, true)
				break
			case NackDiscard:
				err = delivery.Nack(false, false)
				break
			}
			if err != nil {
				log.Fatal("Error acknowledging message: ", err)
			}
		}
	}()

	return nil
}
