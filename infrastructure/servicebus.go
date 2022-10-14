package infrastructure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/ghosts-network/news-feed/utils/logger"
	"github.com/pkg/errors"
	"time"
)

type ServiceBus struct {
	client *azservicebus.Client
}

func NewServiceBus(client *azservicebus.Client) *ServiceBus {
	return &ServiceBus{client: client}
}

func (eb ServiceBus) ListenOne(ctx context.Context, topicName string, subscriptionName string, handler func(context.Context, []byte) error) error {
	receiver, err := eb.client.NewReceiverForSubscription(topicName, subscriptionName, nil)
	if err != nil {
		return err
	}

	go func(receiver *azservicebus.Receiver) {
		for {
			messages, _ := receiver.ReceiveMessages(ctx, 1, nil)
			for _, message := range messages {
				st := time.Now()

				scope := map[string]any{
					"operationId": message.CorrelationID,
					"type":        "outgoing:servicebus",
					"messageId":   message.MessageID,
					"topic":       topicName,
				}

				logger.Info(fmt.Sprintf("Message %s processing started", message.MessageID), &scope)
				err := handler(context.WithValue(context.Background(), "operationId", message.CorrelationID), message.Body)
				scope["elapsedMilliseconds"] = time.Now().Sub(st).Milliseconds()

				if err != nil {
					_ = receiver.AbandonMessage(ctx, message, nil)
					logger.Error(errors.Wrap(err, fmt.Sprintf("Message %s abandoned", message.MessageID)), &scope)
				} else {
					_ = receiver.CompleteMessage(ctx, message, nil)
					logger.Info(fmt.Sprintf("Message %s finished", message.MessageID), &scope)
				}
			}
		}
	}(receiver)

	return nil
}
