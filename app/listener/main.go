package main

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/ghosts-network/news-feed/news"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const subscriptionName string = "ghostnetwork.newsfeed"

func main() {
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
		log.Fatalf("Error trying to listen ghostnetwork.content.publications.created: %v", err.Error())
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
		log.Fatalf("Error trying to listen ghostnetwork.content.publications.updated: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.deleted", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model news.Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = storage.RemovePublication(ctx, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.content.publications.deleted: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestsent", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestSent
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.profiles.friends.requestsent: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestcancelled", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestCancelled
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.profiles.friends.requestcancelled: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestapproved", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestApproved
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.profiles.friends.requestapproved: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.requestdeclined", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model RequestDeclined
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.profiles.friends.requestdeclined: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.profiles.friends.deleted", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model Deleted
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.profiles.friends.deleted: %v", err.Error())
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
	return NewEventBus(client, log.Default())
}
