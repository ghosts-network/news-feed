package news

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoNewsStorage struct {
	publications *mongo.Collection
	sources      *mongo.Collection
	news         *mongo.Collection
}

func NewMongoNewsStorage(mongo *mongo.Client) *MongoNewsStorage {
	ctx := context.Background()
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

	// add publication from source to news feed

	return err
}

func (storage *MongoNewsStorage) RemoveUserSource(ctx context.Context, user string, source string) error {
	f := bson.D{{"user", user}, {"source", source}}
	_, err := storage.sources.DeleteOne(ctx, f)
	if err != nil {
		return err
	}

	// remove publications from source from news feed
	f = bson.D{{"user", user}, {"author.id", source}}
	_, err = storage.news.DeleteMany(ctx, f)

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

	f := bson.D{{"source", publication.Author.Id}}
	cur, err := storage.sources.Find(ctx, f)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var news []interface{}
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

	if len(news) > 0 {
		_, _ = storage.news.InsertMany(ctx, news)
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

	if err != nil {
		return err
	}

	f = bson.D{{"publicationId", publication.Id}}
	_, err = storage.news.UpdateMany(ctx, f, d)

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
		am := m["author"].(primitive.D).Map()
		news = append(news, Publication{
			Id:      m["publicationId"].(string),
			Content: m["content"].(string),
			Author: &PublicationAuthor{
				Id:        am["id"].(string),
				FullName:  am["fullName"].(string),
				AvatarUrl: am["avatarUrl"].(string),
			},
		})
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return news, nil
}
