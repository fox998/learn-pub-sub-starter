package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub/result"
	amqp "github.com/rabbitmq/amqp091-go"
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) result.Type,
	unmarshaller func([]byte) (T, error),
) error {
	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	// channel.Qos(10, 0, false)

	delivery, err := channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range delivery {
			val, err := unmarshaller(d.Body)
			if err != nil {
				continue
			}

			res := handler(val)
			switch res {
			case result.Success:
				log.Printf("Successfully handled message: %v", val)
				err = d.Ack(false)
			case result.RetryRequire:
				log.Printf("Retry required for message: %v", val)
				err = d.Nack(false, true)
			case result.Discard:
				log.Printf("Discarding message: %v", val)
				err = d.Nack(false, false)
			}

			if err != nil {
				log.Printf("Failed to ack/nack delivery: %s", err)
			}

		}
	}()

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) result.Type,
) error {

	subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(b []byte) (T, error) {
			var val T
			err := json.Unmarshal(b, &val)
			return val, err
		},
	)

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) result.Type,
) error {

	subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(b []byte) (T, error) {
			var val T
			err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&val)
			return val, err
		},
	)

	return nil
}
