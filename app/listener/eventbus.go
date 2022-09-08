package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"log"
)

type EventBus struct {
	client *azservicebus.Client
	logger *log.Logger
}

func NewEventBus(client *azservicebus.Client, logger *log.Logger) *EventBus {
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
				eb.logger.Printf("Message %v from %v processing", message.MessageID, topicName)
				if err := handler(message); err != nil {
					_ = receiver.AbandonMessage(ctx, message, nil)
					eb.logger.Printf("Message %v from %v abandoned with error: %v", message.MessageID, topicName, err.Error())
				} else {
					_ = receiver.CompleteMessage(ctx, message, nil)
					eb.logger.Printf("Message %v from %v completed", message.MessageID, topicName)
				}
			}
		}
	}(receiver)

	return nil
}
