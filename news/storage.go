package news

import (
	"context"
	"github.com/ghosts-network/news-feed/utils/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
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
	mc, _ := mongo.Connect(ctx, options.Client().
		ApplyURI(connectionString).
		SetMonitor(&event.CommandMonitor{
			Started: func(ctx context.Context, event *event.CommandStartedEvent) {
				logger.Info("Mongodb query started", &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
			},
			Succeeded: func(ctx context.Context, event *event.CommandSucceededEvent) {
				logger.Info("Mongodb query finished", &map[string]any{
					"operationId":         ctx.Value("operationId"),
					"elapsedMilliseconds": event.DurationNanos / 1000000,
				})
			},
			Failed: func(ctx context.Context, event *event.CommandFailedEvent) {
				logger.Info("Mongodb query failed", &map[string]any{
					"operationId":         ctx.Value("operationId"),
					"elapsedMilliseconds": event.DurationNanos / 1000000,
				})
			},
		}))

	return &MongoNewsStorage{
		publications: mc.Database("newsfeed").Collection("publications"),
		sources:      mc.Database("newsfeed").Collection("sources"),
		news:         mc.Database("newsfeed").Collection("news"),
	}
}

func (storage *MongoNewsStorage) AddUserSources(ctx context.Context, user string, sources []string) error {
	documents := make([]any, 0, len(sources))
	for _, source := range sources {
		documents = append(documents, bson.D{
			{"user", user},
			{"source", source},
		})
	}

	_, err := storage.sources.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	for _, source := range sources {
		ps, err := storage.findPublications(ctx, source)
		var news []interface{}
		for _, p := range ps {
			news = append(news, newsStruct{
				PublicationId: p.Id,
				Source:        source,
				User:          user,
				Order:         p.CreatedOn,
			})
		}

		if err != nil {
			return err
		}

		if len(news) > 0 {
			_, err = storage.news.InsertMany(ctx, news)
		}
	}

	return nil
}

func (storage *MongoNewsStorage) AddUserSource(ctx context.Context, user string, source string) error {
	d := bson.D{
		{"user", user},
		{"source", source},
	}

	_, err := storage.sources.InsertOne(ctx, d)
	if err != nil {
		return err
	}

	// add publication from source to news feed
	ps, err := storage.findPublications(ctx, source)
	var news []interface{}
	for _, p := range ps {
		news = append(news, newsStruct{
			PublicationId: p.Id,
			Source:        source,
			User:          user,
			Order:         p.CreatedOn,
		})
	}

	if len(news) > 0 {
		_, err = storage.news.InsertMany(ctx, news)
	}

	return err
}

func (storage *MongoNewsStorage) RemoveUserSource(ctx context.Context, user string, source string) error {
	_, err := storage.sources.DeleteOne(ctx, bson.D{{"user", user}, {"source", source}})
	if err != nil {
		return err
	}

	_, err = storage.news.DeleteMany(ctx, bson.D{{"user", user}, {"source", source}})

	return err
}

func (storage *MongoNewsStorage) RemovePublications(ctx context.Context) error {
	_, err := storage.publications.DeleteMany(ctx, bson.D{})
	return err
}

func (storage *MongoNewsStorage) RemoveUserSources(ctx context.Context, user string) (err error) {
	_, err = storage.sources.DeleteMany(ctx, bson.D{{"user", user}})

	return
}

func (storage *MongoNewsStorage) AddPublication(ctx context.Context, p *Publication) error {
	oId, err := primitive.ObjectIDFromHex(p.Id)
	if err != nil {
		return err
	}

	_, err = storage.publications.InsertOne(ctx, publicationStruct{
		Id:        oId,
		Content:   p.Content,
		Author:    p.Author,
		CreatedOn: p.CreatedOn.UnixMilli(),
		UpdatedOn: p.UpdatedOn.UnixMilli(),
		Media:     p.Media,
	})

	f := bson.D{{"source", p.Author.Id}}
	cur, err := storage.sources.Find(ctx, f)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var news []interface{}
	for cur.Next(ctx) {
		var result sourceStruct
		err := cur.Decode(&result)
		if err != nil {
			return err
		}

		news = append(news, newsStruct{
			PublicationId: oId,
			Source:        p.Author.Id,
			User:          result.User,
			Order:         p.CreatedOn.UnixMilli(),
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

func (storage *MongoNewsStorage) AddPublications(ctx context.Context, publications []Publication) error {
	ps := make([]any, 0, len(publications))
	for _, p := range publications {
		oId, err := primitive.ObjectIDFromHex(p.Id)
		if err != nil {
			return err
		}

		ps = append(ps, publicationStruct{
			Id:        oId,
			Content:   p.Content,
			Author:    p.Author,
			CreatedOn: p.CreatedOn.UnixMilli(),
			UpdatedOn: p.UpdatedOn.UnixMilli(),
			Media:     p.Media,
		})
	}

	_, err := storage.publications.InsertMany(ctx, ps)

	//for _, p := range publications {
	//	oId, _ := primitive.ObjectIDFromHex(p.Id)
	//	f := bson.D{{"source", p.Author.Id}}
	//	cur, err := storage.sources.Find(ctx, f)
	//	if err != nil {
	//		return err
	//	}
	//
	//	var news []interface{}
	//	for cur.Next(ctx) {
	//		var result sourceStruct
	//		err := cur.Decode(&result)
	//		if err != nil {
	//			return err
	//		}
	//
	//		news = append(news, newsStruct{
	//			PublicationId: oId,
	//			Source:        p.Author.Id,
	//			User:          result.User,
	//			Order:         p.CreatedOn.UnixMilli(),
	//		})
	//	}
	//
	//	if len(news) > 0 {
	//		_, err = storage.news.InsertMany(ctx, news)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	if err := cur.Err(); err != nil {
	//		return err
	//	}
	//
	//	cur.Close(ctx)
	//}

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
	_, err := storage.publications.DeleteOne(ctx, bson.D{{"_id", publication.Id}})
	if err != nil {
		return err
	}

	_, err = storage.news.DeleteMany(ctx, bson.D{{"publicationId", publication.Id}})

	return err
}

func (storage *MongoNewsStorage) FindNews(ctx context.Context, user string, cursor string, take int) ([]Publication, error) {
	pIds, err := storage.findNews(ctx, user, cursor, take)
	if err != nil {
		return nil, err
	}

	if len(pIds) == 0 {
		return make([]Publication, 0), nil
	}

	cur, err := storage.publications.Find(ctx,
		bson.D{{"_id", bson.D{{"$in", pIds}}}},
		options.Find().SetSort(bson.D{{"createdOn", -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var publications []Publication
	for cur.Next(ctx) {
		var result publicationStruct
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}

		publications = append(publications, Publication{
			Id:        result.Id.Hex(),
			Content:   result.Content,
			Author:    result.Author,
			CreatedOn: time.UnixMilli(result.CreatedOn).In(time.UTC),
			UpdatedOn: time.UnixMilli(result.UpdatedOn).In(time.UTC),
			Media:     result.Media,
		})
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return publications, nil
}

func (storage *MongoNewsStorage) findPublications(ctx context.Context, author string) ([]publicationStruct, error) {
	cur, err := storage.publications.Find(ctx, bson.D{{"author._id", author}})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var publications []publicationStruct

	for cur.Next(ctx) {
		var result publicationStruct
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}

		publications = append(publications, result)
	}

	return publications, nil
}

func (storage *MongoNewsStorage) findNews(ctx context.Context, user string, cursor string, take int) ([]primitive.ObjectID, error) {
	filter := bson.M{"user": user}
	if oId, err := primitive.ObjectIDFromHex(cursor); err == nil {
		filter["publicationId"] = bson.M{"$lt": oId}
	}

	cur, err := storage.news.Find(ctx,
		filter,
		options.Find().
			SetSort(bson.D{{"order", -1}}).
			SetLimit(int64(take)))

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var news []primitive.ObjectID
	for cur.Next(ctx) {
		var result newsStruct
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

type publicationStruct struct {
	Id        primitive.ObjectID `bson:"_id"`
	Content   string             `bson:"content"`
	Author    *PublicationAuthor `bson:"author"`
	CreatedOn int64              `bson:"createdOn"`
	UpdatedOn int64              `bson:"updatedOn"`
	Media     []*Media           `bson:"media"`
}

type newsStruct struct {
	PublicationId primitive.ObjectID `bson:"publicationId"`
	Source        string             `bson:"source"`
	User          string             `bson:"user"`
	Order         int64              `bson:"order"`
}

type sourceStruct struct {
	User   string `bson:"user"`
	Source string `bson:"source"`
}
