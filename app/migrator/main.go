package main

import (
	"context"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/news"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	profileClient := infrastructure.NewProfilesClient("http://localhost:5000")
	relationsClient := infrastructure.NewRelationsClient("http://localhost:5000")
	publicationsClient := infrastructure.NewPublicationsClient("http://localhost:5100")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONNECTION")))
	newsStorage := news.NewMongoNewsStorage(mongoClient)
	start := time.Now()

	m := NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage)
	//m.MigrateUsers()
	m.MigratePublications()
	log.Printf("[INFO] Migration finished in %s\n", time.Now().Sub(start))
}

type Migrator struct {
	pc         *infrastructure.ProfilesClient
	rc         *infrastructure.RelationsClient
	pubsClient *infrastructure.PublicationsClient
	ns         NewsStorage
}

func NewMigrator(pc *infrastructure.ProfilesClient, rc *infrastructure.RelationsClient, pubsClient *infrastructure.PublicationsClient, ns NewsStorage) *Migrator {
	return &Migrator{pc: pc, rc: rc, pubsClient: pubsClient, ns: ns}
}

func (m Migrator) MigrateUsers() {
	skip := 0
	take := 20

	for {
		ps, err := m.pc.GetProfiles(skip, take)
		if err != nil {
			log.Println(err)
			return
		}

		if len(ps) == 0 {
			break
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(ps))
		for _, profile := range ps {
			go m.MigrateUserAsync(profile.Id, wg)
		}
		wg.Wait()

		if len(ps) < take {
			break
		}

		skip += take
	}
}

func (m Migrator) MigrateUser(user string) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_ = m.ns.RemoveUserSources(ctx, user)
	m.migrateFriends(user)
	m.migrateOutgoingRequests(user)
}

func (m Migrator) MigrateUserAsync(user string, wg *sync.WaitGroup) {
	m.MigrateUser(user)
	log.Printf("[INFO] Migration for %s finished\n", user)
	wg.Done()
}

func (m Migrator) MigratePublications() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_ = m.ns.RemoveAllNews(ctx)

	var cursor string
	take := 20

	for {
		ps, nextCursor, err := m.pubsClient.GetPublications(cursor, take)
		if err != nil {
			log.Println(err)
			return
		}

		if len(ps) == 0 {
			break
		}

		for _, publication := range ps {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			m.ns.AddPublication(ctx, &news.Publication{
				Id:        publication.Id,
				Content:   publication.Content,
				Author:    publication.Author,
				CreatedOn: publication.CreatedOn.UnixMilli(),
				UpdatedOn: publication.UpdatedOn.UnixMilli(),
				Media:     publication.Media,
			})
		}

		if len(ps) < take {
			break
		}

		cursor = nextCursor
	}
}

func (m Migrator) migrateFriends(user string) {
	skip := 0
	take := 20

	for {
		friends, err := m.rc.GetFriends(user, skip, take)
		if err != nil {
			log.Println(err)
			return
		}

		if len(friends) == 0 {
			break
		}

		for _, friend := range friends {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			err := m.ns.AddUserSource(ctx, user, friend)
			if err != nil {
				log.Printf("[ERR] Failed to migrate friend %s for %s\n", friend, user)
			} else {
				log.Printf("[INFO] Friend %s for %s migrated\n", friend, user)
			}
		}

		if len(friends) < take {
			break
		}
		skip += take
	}
}

func (m Migrator) migrateOutgoingRequests(user string) {
	skip := 0
	take := 20

	for {
		rs, err := m.rc.GetOutgoingRequests(user, skip, take)
		if err != nil {
			log.Println(err)
			return
		}

		if len(rs) == 0 {
			break
		}

		for _, r := range rs {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			err := m.ns.AddUserSource(ctx, user, r)
			if err != nil {
				log.Printf("[ERR] Failed to migrate outgoing request from %s to %s\n", user, r)
			} else {
				log.Printf("[INFO] Outgoing request from %s to %s migrated\n", user, r)
			}
		}

		if len(rs) < take {
			break
		}

		skip += take
	}
}

type NewsStorage interface {
	AddPublication(ctx context.Context, publication *news.Publication) error
	AddUserSource(ctx context.Context, user string, source string) error
	RemoveAllNews(ctx context.Context) error
	RemoveUserSources(ctx context.Context, user string) error
}
