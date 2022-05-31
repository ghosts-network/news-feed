package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"os"
	"time"
)

var storage NewsStorage
var eventbus *EventBus

func main() {
	connectionString := os.Getenv("SERVICEBUS_CONNECTION")
	client, _ := azservicebus.NewClientFromConnectionString(connectionString, nil)
	eventbus = NewEventBus(client, log.Default())

	storage = NewRedisNewsStorage(nil)

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)

	eventbus.ListenOne(ctx, "ghostnetwork.content.publications.created", "tests", func(message *azservicebus.ReceivedMessage) error {
		var model Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			log.Printf("%v\n", err.Error())
			return err
		}

		storage.Add(model)
		return nil
	})

	eventbus.ListenOne(ctx, "ghostnetwork.content.publications.deleted", "tests", func(message *azservicebus.ReceivedMessage) error {
		var model Publication
		err := json.Unmarshal(message.Body, &model)
		if err != nil {
			log.Printf("%v\n", err.Error())
			return err
		}

		storage.Add(model)
		return nil
	})

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
