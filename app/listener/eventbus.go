package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/ghosts-network/news-feed/utils"
	"github.com/pkg/errors"
	"time"
)

type EventBus struct {
	client *azservicebus.Client
	logger *utils.Logger
}

func NewEventBus(client *azservicebus.Client, logger *utils.Logger) *EventBus {
	return &EventBus{
		client: client,
		logger: logger,
	}
}

func (eb EventBus) ListenOne(ctx context.Context, topicName string, subscriptionName string, handler func(*azservicebus.ReceivedMessage) error) error {
	receiver, err := eb.client.NewReceiverForSubscription(topicName, subscriptionName, nil)
	if err != nil {
		return err
	}

	go func(receiver *azservicebus.Receiver) {
		for {
			messages, _ := receiver.ReceiveMessages(ctx, 1, nil)
			for _, message := range messages {
				st := time.Now()

				scopedLogger := eb.logger.
					WithValue("messageId", message.MessageID).
					WithValue("topic", topicName).
					Info(fmt.Sprintf("Message %s processing started", message.MessageID))

				if err := handler(message); err != nil {
					_ = receiver.AbandonMessage(ctx, message, nil)
					scopedLogger.
						WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
						Error(errors.Wrap(err, fmt.Sprintf("Message %s abandoned", message.MessageID)))
				} else {
					_ = receiver.CompleteMessage(ctx, message, nil)
					scopedLogger.
						WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
						Info(fmt.Sprintf("Message %s abandoned", message.MessageID))
				}
			}
		}
	}(receiver)

	return nil
}
