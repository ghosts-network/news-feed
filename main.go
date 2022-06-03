package main

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/gorilla/mux"
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

	storage = NewMongoNewsStorage(nil)

	//ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	//go runServer()
	//go runBackgroundSubscriptions(ctx)
	//
	//_, _ = fmt.Scanln()
	//defer cancel()
}

func runBackgroundSubscriptions(ctx context.Context) {
	err := eventbus.ListenOne(ctx, "ghostnetwork.content.publications.created", subscriptionName, func(message *azservicebus.ReceivedMessage) error {
		var model Publication
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
		var model Publication
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
		var model Publication
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
}

func runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/{user}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming http request %v\n", r.RequestURI)
		user := mux.Vars(r)["user"]
		cursor := mux.Vars(r)["cursor"]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		news, err := storage.FindNews(ctx, user, cursor)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		body, err := json.Marshal(news)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write(body)
	}).Methods(http.MethodGet)
	log.Println("Starting http server on port 10000")
	log.Fatal(http.ListenAndServe(":10000", r))
}
