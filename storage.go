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
	publications *mongo.Collection
	sources      *mongo.Collection
	news         *mongo.Collection
}

func NewMongoNewsStorage(mongo *mongo.Client) *MongoNewsStorage {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_ = mongo.Connect(ctx)

	return &MongoNewsStorage{
		publications: mongo.Database("newsfeed").Collection("publications"),
		sources:      mongo.Database("newsfeed").Collection("sources"),
		news:         mongo.Database("newsfeed").Collection("news"),
	}
}

func (storage *MongoNewsStorage) AddUserSource(ctx context.Context, user string, source string) error {
	d := bson.D{
		{"user", user},
		{"source", source},
	}

	_, err := storage.sources.InsertOne(ctx, d)

	return err
}

func (storage *MongoNewsStorage) RemoveUserSource(ctx context.Context, user string, source string) error {
	f := bson.D{{"user", user}, {"source", source}}
	_, err := storage.sources.DeleteOne(ctx, f)

	return err
}

func (storage *MongoNewsStorage) AddPublication(ctx context.Context, publication *Publication) error {
	d := bson.D{
		{"_id", publication.Id},
		{"content", publication.Content},
		{"author", bson.D{
			{"id", publication.Author.Id},
			{"fullName", publication.Author.FullName},
			{"avatarUrl", publication.Author.AvatarUrl},
		}},
	}

	_, err := storage.publications.InsertOne(ctx, d)

	// find all users who have publication.Author.Id source
	f := bson.D{{"source", publication.Author.Id}}
	cur, err := storage.sources.Find(ctx, f)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var news []bson.D
	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			return err
		}

		m := result.Map()
		news = append(news, bson.D{
			{"publicationId", publication.Id},
			{"user", m["user"]},
			{"content", publication.Content},
			{"author", bson.D{
				{"id", publication.Author.Id},
				{"fullName", publication.Author.FullName},
				{"avatarUrl", publication.Author.AvatarUrl},
			}},
		})
	}
	if err := cur.Err(); err != nil {
		return err
	}

	return err
}

func (storage *MongoNewsStorage) UpdatePublication(ctx context.Context, publication *Publication) error {
	f := bson.D{{"_id", publication.Id}}
	d := bson.D{
		{"$set", bson.D{{"content", publication.Content}}},
	}

	_, err := storage.publications.UpdateOne(ctx, f, d)

	// TODO: update publication in news

	return err
}

func (storage *MongoNewsStorage) RemovePublication(ctx context.Context, publication *Publication) error {
	f := bson.D{{"_id", publication.Id}}
	_, err := storage.publications.DeleteOne(ctx, f)
	if err != nil {
		return err
	}

	f = bson.D{{"publicationId", publication.Id}}
	_, err = storage.news.DeleteMany(ctx, f)

	return err
}

func (storage *MongoNewsStorage) FindNews(ctx context.Context, user string, cursor string) ([]Publication, error) {
	f := bson.D{{"user", user}}
	cur, err := storage.news.Find(ctx, f)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var news []Publication
	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}

		m := result.Map()
		news = append(news, Publication{
			Id:      m["publicationId"].(string),
			Content: m["content"].(string),
			Author: &PublicationAuthor{
				Id:        "",
				FullName:  "",
				AvatarUrl: "",
			},
		})
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return news, nil
}
