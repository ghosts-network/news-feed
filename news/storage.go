package news

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoNewsStorage struct {
	publications *mongo.Collection
	sources      *mongo.Collection
	news         *mongo.Collection
}

func NewMongoNewsStorage(connectionString string) *MongoNewsStorage {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc, _ := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))

	return &MongoNewsStorage{
		publications: mc.Database("newsfeed").Collection("sourcePublications"),
		sources:      mc.Database("newsfeed").Collection("sources"),
		news:         mc.Database("newsfeed").Collection("news"),
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

func (storage *MongoNewsStorage) RemoveAllNews(ctx context.Context) error {
	_, err := storage.publications.DeleteMany(ctx, bson.D{})
	if err != nil {
		return err
	}
	_, err = storage.news.DeleteMany(ctx, bson.D{})
	if err != nil {
		return err
	}

	return nil
}

func (storage *MongoNewsStorage) RemoveUserSources(ctx context.Context, user string) (err error) {
	_, err = storage.sources.DeleteMany(ctx, bson.D{{"user", user}})

	return
}

func (storage *MongoNewsStorage) AddPublication(ctx context.Context, p *Publication) error {
	_, err := storage.publications.InsertOne(ctx, p)

	f := bson.D{{"source", p.Author.Id}}
	cur, err := storage.sources.Find(ctx, f)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var news []interface{}
	for cur.Next(ctx) {
		var result SourceStruct
		err := cur.Decode(&result)
		if err != nil {
			return err
		}

		news = append(news, NewsStruct{
			PublicationId: p.Id,
			Source:        p.Author.Id,
			User:          result.User,
			Order:         p.CreatedOn,
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
	pIds, err := storage.findNews(ctx, user, cursor)
	if err != nil {
		return nil, err
	}

	if len(pIds) == 0 {
		return make([]Publication, 0), nil
	}

	cur, err := storage.publications.Find(ctx,
		bson.D{{"_id", bson.D{{"$in", pIds}}}},
		options.Find().SetSort(bson.D{{"createOn", -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var publications []Publication
	for cur.Next(ctx) {
		var result Publication
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}

		publications = append(publications, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return publications, nil
}

func (storage *MongoNewsStorage) findNews(ctx context.Context, user string, cursor string) ([]string, error) {
	cur, err := storage.news.Find(ctx,
		bson.D{{"user", user}},
		options.Find().SetSort(bson.D{{"order", -1}}))

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var news []string
	for cur.Next(ctx) {
		var result NewsStruct
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}

		news = append(news, result.PublicationId)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return news, nil
}

type NewsStruct struct {
	PublicationId string `bson:"publicationId"`
	Source        string `bson:"source"`
	User          string `bson:"user"`
	Order         int64  `bson:"order"`
}

type SourceStruct struct {
	User   string `bson:"user"`
	Source string `bson:"source"`
}
