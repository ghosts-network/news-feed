package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/ghosts-network/news-feed/news"
	"github.com/ghosts-network/news-feed/utils"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log *utils.Logger

const subscriptionName string = "ghostnetwork.newsfeed"

func main() {
	log = utils.NewLogger("news-feed-listener")

	storage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))
	eventbus := configureEventBus(os.Getenv("SERVICEBUS_CONNECTION"))

	ctx, cancel := context.WithCancel(context.Background())
	err := eventbus.ListenOne(ctx, "ghostnetwork.content.publications.created", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model news.Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = storage.AddPublication(ctx, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.content.publications.created"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.content.publications.created"))
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.updated", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model news.Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = storage.UpdatePublication(ctx, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.content.publications.updated"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.content.publications.updated"))
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.deleted", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model news.Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return storage.RemovePublication(ctx, &model)
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.content.publications.deleted"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.content.publications.deleted"))
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestsent", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestSent
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return storage.AddUserSource(ctx, model.FromUser, model.ToUser)
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.requestsent"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.requestsent"))
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestcancelled", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestCancelled
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return storage.RemoveUserSource(ctx, model.FromUser, model.ToUser)
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.requestcancelled"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.requestcancelled"))
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestapproved", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestApproved
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return storage.AddUserSource(ctx, model.User, model.Requester)
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.requestapproved"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.requestapproved"))
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.deleted", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model Deleted
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return storage.RemoveUserSource(ctx, model.User, model.Friend)
	})
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to subscribe on ghostnetwork.profiles.friends.deleted"))
	} else {
		log.Info(fmt.Sprintf("Successfully subscribed to topic ghostnetwork.profiles.friends.deleted"))
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
	cancel()
}

func configureEventBus(connectionString string) *EventBus {
	client, _ := azservicebus.NewClientFromConnectionString(connectionString, nil)
	return NewEventBus(client, log)
}
