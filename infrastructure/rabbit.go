package infrastructure

import (
	"context"
	"fmt"
	"github.com/ghosts-network/news-feed/utils/logger"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type RabbitMq struct {
	client *amqp.Connection
}

func NewRabbitMq(client *amqp.Connection) *RabbitMq {
	return &RabbitMq{client: client}
}

func (r RabbitMq) ListenOne(ctx context.Context, topicName string, subscriptionName string, handler func(context.Context, []byte) error) error {
	channel, err := r.client.Channel()
	if err != nil {
		return err
	}

	err = channel.ExchangeDeclare(topicName, "fanout", false, false, false, false, nil)
	if err != nil {
		return err
	}

	queue, err := channel.QueueDeclare(subscriptionName+"/"+topicName, false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = channel.QueueBind(queue.Name, "", topicName, false, nil)
	if err != nil {
		return err
	}

	messagesCh, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for message := range messagesCh {
			st := time.Now()

			scope := map[string]any{
				"operationId": message.CorrelationId,
				"type":        "outgoing:rabbitmq",
				"messageId":   message.MessageId,
				"topic":       topicName,
			}

			logger.Info(fmt.Sprintf("Message %s processing started", message.MessageId), &scope)
			err := handler(context.WithValue(context.Background(), "operationId", message.CorrelationId), message.Body)
			scope["elapsedMilliseconds"] = time.Now().Sub(st).Milliseconds()

			if err != nil {
				_ = channel.Reject(message.DeliveryTag, true)
				logger.Error(errors.Wrap(err, fmt.Sprintf("Message %s abandoned", message.MessageId)), &scope)
			} else {
				_ = channel.Ack(message.DeliveryTag, false)
				logger.Info(fmt.Sprintf("Message %s finished", message.MessageId), &scope)
			}
		}
	}()

	return nil
}
