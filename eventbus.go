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

func (eb EventBus) ListenOne(ctx context.Context, topic string, subscription string, handler func(*azservicebus.ReceivedMessage) error) {
	receiver, _ := eb.client.NewReceiverForSubscription(topic, subscription, nil)
	go func(receiver *azservicebus.Receiver) {
		for {
			messages, _ := receiver.ReceiveMessages(ctx, 1, nil)
			for _, message := range messages {
				eb.logger.Printf("Message %v from %v processing", message.MessageID, topic)
				if err := handler(message); err != nil {
					_ = receiver.AbandonMessage(ctx, message, nil)
					eb.logger.Printf("Message %v from %v abandoned", message.MessageID, topic)
				} else {
					_ = receiver.CompleteMessage(ctx, message, nil)
					eb.logger.Printf("Message %v from %v completed", message.MessageID, topic)
				}
			}
		}
	}(receiver)
}
