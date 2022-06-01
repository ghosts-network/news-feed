package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
	"log"
)

type AzureOptions struct {
	SubscriptionId string
	ResourceGroup  string
	Credential     azcore.TokenCredential
}

type EventBusOptions struct {
	Azure     *AzureOptions
	Namespace string
}

type EventBus struct {
	client  *azservicebus.Client
	logger  *log.Logger
	options *EventBusOptions
}

func NewEventBus(client *azservicebus.Client, logger *log.Logger, options *EventBusOptions) *EventBus {
	return &EventBus{
		client:  client,
		logger:  logger,
		options: options,
	}
}

func (eb EventBus) ListenOne(ctx context.Context, topicName string, subscriptionName string, handler func(*azservicebus.ReceivedMessage) error) error {
	if err := eb.ensureSubscriptionExists(topicName, subscriptionName); err != nil {
		return err
	}

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
					eb.logger.Printf("Message %v from %v abandoned", message.MessageID, topicName)
				} else {
					_ = receiver.CompleteMessage(ctx, message, nil)
					eb.logger.Printf("Message %v from %v completed", message.MessageID, topicName)
				}
			}
		}
	}(receiver)

	return nil
}

func (eb EventBus) ensureSubscriptionExists(topicName string, subscriptionName string) error {
	subscriptionsClient, err := armservicebus.NewSubscriptionsClient(eb.options.Azure.SubscriptionId, eb.options.Azure.Credential, nil)
	if err != nil {
		return err
	}

	_, err = subscriptionsClient.CreateOrUpdate(
		context.Background(),
		eb.options.Azure.ResourceGroup,
		eb.options.Namespace,
		topicName,
		subscriptionName,
		armservicebus.SBSubscription{},
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}
