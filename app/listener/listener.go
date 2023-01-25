package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/news"
	"github.com/ghosts-network/news-feed/utils/logger"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strings"
	"time"
)

const subscriptionName string = "ghostnetwork.newsfeed"

type Listener struct {
	exit <-chan os.Signal
}

func NewListener(exit <-chan os.Signal) *Listener {
	return &Listener{exit: exit}
}

func (l Listener) Run() {
	log.SetFlags(0)

	storage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))

	ctx := context.Background()

	eventbus, err := getEventBus()
	if err != nil {
		logger.Error(err, &map[string]any{})
		return
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.created", subscriptionName, func(ctx context.Context, message []byte) error {
		var model news.Publication
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return storage.AddPublication(ctx, &model)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.content.publications.created"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.content.publications.created"), &map[string]any{})
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.updated", subscriptionName, func(ctx context.Context, message []byte) error {
		var model news.Publication
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return storage.UpdatePublication(ctx, &model)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.content.publications.updated"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.content.publications.updated"), &map[string]any{})
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.deleted", subscriptionName, func(ctx context.Context, message []byte) error {
		var model news.Publication
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return storage.RemovePublication(ctx, &model)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.content.publications.deleted"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.content.publications.deleted"), &map[string]any{})
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestsent", subscriptionName, func(ctx context.Context, message []byte) error {
		var model RequestSent
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return storage.AddUserSource(ctx, model.FromUser, model.ToUser)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.requestsent"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.requestsent"), &map[string]any{})
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestcancelled", subscriptionName, func(ctx context.Context, message []byte) error {
		var model RequestCancelled
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return storage.RemoveUserSource(ctx, model.FromUser, model.ToUser)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.requestcancelled"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.requestcancelled"), &map[string]any{})
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestapproved", subscriptionName, func(ctx context.Context, message []byte) error {
		var model RequestApproved
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return storage.AddUserSource(ctx, model.User, model.Requester)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.requestapproved"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.requestapproved"), &map[string]any{})
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.deleted", subscriptionName, func(ctx context.Context, message []byte) error {
		var model Deleted
		err := json.Unmarshal(message, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return storage.RemoveUserSource(ctx, model.User, model.Friend)
	})
	if err != nil {
		logger.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.deleted"), &map[string]any{})
	} else {
		logger.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.deleted"), &map[string]any{})
	}

	<-l.exit
}

func getEventBus() (EventListener, error) {
	if strings.ToLower(os.Getenv("EVENTHUB_TYPE")) == "servicebus" {
		eventbus, err := configureServiceBus(os.Getenv("SERVICEBUS_CONNECTION"))
		return eventbus, err
	} else if strings.ToLower(os.Getenv("EVENTHUB_TYPE")) == "rabbit" {
		eventbus, err := configureRabbit(os.Getenv("RABBIT_CONNECTION"))
		return eventbus, err
	} else {
		return &NullEventListener{}, nil
	}
}

func configureServiceBus(connectionString string) (*infrastructure.ServiceBus, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	return infrastructure.NewServiceBus(client), err
}

func configureRabbit(connectionString string) (*infrastructure.RabbitMq, error) {
	conn, err := amqp.Dial(connectionString)
	return infrastructure.NewRabbitMq(conn), err
}

type EventListener interface {
	ListenOne(ctx context.Context, topicName string, subscriptionName string, handler func(context.Context, []byte) error) error
}

type NullEventListener struct {
}

func (n NullEventListener) ListenOne(ctx context.Context, topicName string, subscriptionName string, handler func(context.Context, []byte) error) error {
	return nil
}
