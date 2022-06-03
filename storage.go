package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type NewsStorage interface {
	AddPublication(ctx context.Context, publication *Publication) error
	UpdatePublication(ctx context.Context, publication *Publication) error
	RemovePublication(ctx context.Context, publication *Publication) error

	AddUserSource(ctx context.Context, user string, source string) error
	RemoveUserSource(ctx context.Context, user string, source string) error

	FindNews(ctx context.Context, user string, cursor string) ([]Publication, error)
}

type MongoNewsStorage struct {
	mongo *mongo.Database
}

func NewMongoNewsStorage(mongo *mongo.Client) *MongoNewsStorage {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_ = mongo.Connect(ctx)

	return &MongoNewsStorage{
		mongo: mongo.Database("newsfeed"),
	}
}

func (storage *MongoNewsStorage) AddUserSource(ctx context.Context, user string, source string) error {
	d := bson.D{
		{"user", user},
		{"source", source},
	}

	_, err := storage.mongo.Collection("sources").InsertOne(ctx, d)

	return err
}

func (storage *MongoNewsStorage) RemoveUserSource(ctx context.Context, user string, source string) error {
	f := bson.D{{"user", user}, {"source", source}}
	_, err := storage.mongo.Collection("sources").DeleteOne(ctx, f)

	return err
}

func (storage *MongoNewsStorage) AddPublication(ctx context.Context, publication *Publication) error {
	f := bson.D{{"id", publication.Id}}
	d := bson.D{
		{"$set", bson.D{{"content", publication.Content}}},
	}

	_, err := storage.mongo.Collection("publications").UpdateOne(ctx, f, d)

	// TODO: find all users who have publication.Author.Id source

	return err
}

func (storage *MongoNewsStorage) UpdatePublication(ctx context.Context, publication *Publication) error {
	d := bson.D{
		{"_id", publication.Id},
		{"content", publication.Content},
		{"author", bson.D{
			{"id", publication.Author.Id},
			{"fullName", publication.Author.FullName},
			{"avatarUrl", publication.Author.AvatarUrl},
		}},
	}

	_, err := storage.mongo.Collection("publications").InsertOne(ctx, d)

	return err
}

func (storage *MongoNewsStorage) RemovePublication(ctx context.Context, publication *Publication) error {
	f := bson.D{{"_id", publication.Id}}
	_, err := storage.mongo.Collection("publications").DeleteOne(ctx, f)

	// TODO: remove publications from all news sources

	return err
}

func (storage *MongoNewsStorage) FindNews(ctx context.Context, user string, cursor string) ([]Publication, error) {

	return nil, nil
}
