package main

import (
	"context"
	"github.com/ghosts-network/news-feed/news"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

func main() {
	newsFeedStorage := configureNewsStorage(os.Getenv("MONGO_CONNECTION"))
	err := migrateSourcesFromMongo(newsFeedStorage)
	if err != nil {
		log.Fatalln(err)
	}
}

type NewsStorage interface {
	AddPublication(ctx context.Context, publication *news.Publication) error
	AddUserSource(ctx context.Context, user string, source string) error
	RemoveAllSources(ctx context.Context) error
}

func configureNewsStorage(connectionString string) NewsStorage {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	return news.NewMongoNewsStorage(mongoClient)
}

func migrateSourcesFromMongo(newsFeedStorage NewsStorage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	err := newsFeedStorage.RemoveAllSources(ctx)
	if err != nil {
		return err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONNECTION")))
	if err != nil {
		return err
	}

	cursor, err := client.
		Database("ghost-network").
		Collection("friendRequests").
		Find(ctx, bson.D{{"status", bson.D{{"$in", []int{1, 2, 3}}}}})

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		result := relation{}
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		newsFeedStorage.AddUserSource(ctx, result.FromUser, result.ToUser)
		log.Printf("Added new source %s for %s\n", result.ToUser, result.FromUser)
	}
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

type relation struct {
	FromUser string `bson:"fromUser"`
	ToUser   string `bson:"toUser"`
}
