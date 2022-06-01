package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"os"
	"time"
)

const subscriptionName string = "ghostnetwork.newsfeed"

var storage NewsStorage
var eventbus *EventBus

func main() {
	connectionString := os.Getenv("SERVICEBUS_CONNECTION")
	cred, err := azidentity.NewClientSecretCredential(os.Getenv("AZURE_TENANT_ID"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), nil)
	if err != nil {
		log.Fatalf("Error configuring azure credentials: %v", err.Error())
	}

	client, _ := azservicebus.NewClientFromConnectionString(connectionString, nil)
	eventbus = NewEventBus(client, log.Default(), &EventBusOptions{
		Namespace: os.Getenv("SERVICEBUS_NAMESPACE"),
		Azure: &AzureOptions{
			SubscriptionId: os.Getenv("SERVICEBUS_AZURE_SUBSCRIPTION_ID"),
			ResourceGroup:  os.Getenv("SERVICEBUS_AZURE_RESOURCE_GROUP"),
			Credential:     cred,
		},
	})

	storage = NewRedisNewsStorage(nil)

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.created", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error trying to listen ghostnetwork.content.publications.created: %v", err.Error())
	}

	err = eventbus.ListenOne(ctx, "ghostnetwork.content.publications.deleted", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model Publication
		err := json.Unmarshal(message.Body, &model)
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

	go runServer()

	_, _ = fmt.Scanln()
	defer cancel()
}

func runServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Endpoint Hit")

		body, err := json.Marshal(storage.Find())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write(body)
	})
	log.Fatal(http.ListenAndServe(":10000", nil))
}

type NewsStorage interface {
	Add(publication Publication)
	Remove(publication Publication)
	Find() []Publication
}

type RedisNewsStorage struct {
	rdb *redis.Client
}

func (storage *RedisNewsStorage) Add(publication Publication) {
}

func (storage *RedisNewsStorage) Remove(publication Publication) {
}

func (storage *RedisNewsStorage) Find() []Publication {
	return nil
}

func NewRedisNewsStorage(rdb *redis.Client) *RedisNewsStorage {
	return &RedisNewsStorage{rdb: rdb}
}

type Publication struct {
	Id      string
	Content string
	Author  *PublicationAuthor
}

type PublicationAuthor struct {
	Id        string
	FullName  string
	AvatarUrl string
}

type RequestSent struct {
	FromUser string
	ToUser   string
}

type RequestCancelled struct {
	FromUser string
	ToUser   string
}

type RequestApproved struct {
	User      string
	Requester string
}

type RequestDeclined struct {
	User      string
	Requester string
}

type Deleted struct {
	User   string
	Friend string
}
