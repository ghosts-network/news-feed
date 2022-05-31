package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"log"
	"os"
	"time"
)

var client *azservicebus.Client

func main() {
	connectionString := os.Getenv("SERVICEBUS_CONNECTION")
	client, _ = azservicebus.NewClientFromConnectionString(connectionString, nil)
	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)

	go subscribeOnPublicationCreated(ctx)
	go subscribeOnPublicationDeleted(ctx)

	fmt.Scanln()
	defer cancel()
}

func subscribeOnPublicationCreated(ctx context.Context) {
	receiver, _ := client.NewReceiverForSubscription("ghostnetwork.content.publications.created", "tests", nil)
	for {
		messages, _ := receiver.ReceiveMessages(ctx, 1, nil)
		for _, message := range messages {
			var model PublicationCreated
			_ = json.Unmarshal(message.Body, &model)
			log.Printf("%+v\n", model)
		}
	}
}

func subscribeOnPublicationDeleted(ctx context.Context) {
	receiver, _ := client.NewReceiverForSubscription("ghostnetwork.content.publications.deleted", "tests", nil)
	for {
		messages, _ := receiver.ReceiveMessages(ctx, 1, nil)
		for _, message := range messages {
			var model PublicationDeleted
			_ = json.Unmarshal(message.Body, &model)
			log.Printf("%+v\n", model)
		}
	}
}

//  {"Id":"6295f3db68297744ef421447","Content":"22","Author":{"Id":"5e35fa4e-6cde-4be8-8178-e7afa036256b","FullName":"Vladimir Borodin","AvatarUrl":"https://ghostnetwork.blob.core.windows.net/photos/5e35fa4e-6cde-4be8-8178-e7afa036256b/6e9f8721-1ba1-4f84-a91b-fffddada5bb1.png"},"CreatedOn":"2022-05-31T10:54:19.1542376+00:00","TrackerId":"419a2ff6-f8f8-4ea3-9df8-cec9e29f1ab0"}
type PublicationCreated struct {
	Id      string
	Content string
	Author  *PublicationAuthor
}

type PublicationDeleted struct {
	Id      string
	Content string
	Author  *PublicationAuthor
}

type PublicationAuthor struct {
	Id        string
	FullName  string
	AvatarUrl string
}
