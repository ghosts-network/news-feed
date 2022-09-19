package migrator

import (
	"context"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/news"
	"log"
	"sync"
	"time"
)

type Migrator struct {
	pc         *infrastructure.ProfilesClient
	rc         *infrastructure.RelationsClient
	pubsClient *infrastructure.PublicationsClient
	ns         *news.MongoNewsStorage
}

func NewMigrator(pc *infrastructure.ProfilesClient, rc *infrastructure.RelationsClient, pubsClient *infrastructure.PublicationsClient, ns *news.MongoNewsStorage) *Migrator {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = m.ns.AddPublication(ctx, &news.Publication{
				Id:        publication.Id,
				Content:   publication.Content,
				Author:    publication.Author,
				CreatedOn: publication.CreatedOn.UnixMilli(),
				UpdatedOn: publication.UpdatedOn.UnixMilli(),
				Media:     publication.Media,
			})

			cancel()
			log.Printf("[INFO] Publication %s migrated\n", publication.Id)
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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := m.ns.AddUserSource(ctx, user, friend)
			if err != nil {
				log.Printf("[ERR] Failed to migrate friend %s for %s\n", friend, user)
			} else {
				log.Printf("[INFO] Friend %s for %s migrated\n", friend, user)
			}

			cancel()
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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := m.ns.AddUserSource(ctx, user, r)
			if err != nil {
				log.Printf("[ERR] Failed to migrate outgoing request from %s to %s\n", user, r)
			} else {
				log.Printf("[INFO] Outgoing request from %s to %s migrated\n", user, r)
			}

			cancel()
		}

		if len(rs) < take {
			break
		}

		skip += take
	}
}
