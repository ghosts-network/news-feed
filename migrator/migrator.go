package migrator

import (
	"context"
	"fmt"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/news"
	"github.com/ghosts-network/news-feed/utils"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Migrator struct {
	pc         *infrastructure.ProfilesClient
	rc         *infrastructure.RelationsClient
	pubsClient *infrastructure.PublicationsClient
	ns         *news.MongoNewsStorage
	logger     *utils.Logger
}

func NewMigrator(pc *infrastructure.ProfilesClient, rc *infrastructure.RelationsClient, pubsClient *infrastructure.PublicationsClient, ns *news.MongoNewsStorage, logger *utils.Logger) *Migrator {
	return &Migrator{pc: pc, rc: rc, pubsClient: pubsClient, ns: ns, logger: logger}
}

func (m Migrator) MigrateUsers() {
	st := time.Now()

	skip := 0
	take := 20

	for {
		ps, err := m.pc.GetProfiles(skip, take)
		if err != nil {
			m.logger.Error(err)
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

	m.logger.
		WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
		Info("Users migration finished")
}

func (m Migrator) MigrateUser(user string) {
	st := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	_ = m.ns.RemoveUserSources(ctx, user)
	m.migrateFriends(ctx, user)
	m.migrateOutgoingRequests(ctx, user)

	m.logger.
		WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
		Info(fmt.Sprintf("User %s migration finished", user))
}

func (m Migrator) MigrateUserAsync(user string, wg *sync.WaitGroup) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	_ = m.ns.RemoveUserSources(ctx, user)
	m.migrateFriends(ctx, user)
	m.migrateOutgoingRequests(ctx, user)
	wg.Done()

	m.logger.
		Debug(fmt.Sprintf("User %s migration finished", user))
}

func (m Migrator) MigratePublications() {
	st := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	_ = m.ns.RemoveAllNews(ctx)

	var cursor string
	take := 20

	for {
		ps, nextCursor, err := m.pubsClient.GetPublications(cursor, take)
		if err != nil {
			m.logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch publications with cursor: %s, count: %d", cursor, take)))
			return
		}

		if len(ps) == 0 {
			break
		}

		for _, publication := range ps {
			err = m.ns.AddPublication(ctx, &publication)

			if err != nil {
				m.logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate publication %s", publication.Id)))
			} else {
				m.logger.Debug(fmt.Sprintf("Publication %s migrated", publication.Id))
			}
		}

		if len(ps) < take {
			break
		}

		cursor = nextCursor
	}

	m.logger.
		WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
		Info("Publications migration finished")
}

func (m Migrator) migrateFriends(ctx context.Context, user string) {
	skip := 0
	take := 20

	for {
		friends, err := m.rc.GetFriends(user, skip, take)
		if err != nil {
			m.logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch friends for %s with skip: %d, count: %d", user, skip, take)))
			return
		}

		if len(friends) == 0 {
			break
		}

		for _, friend := range friends {
			err := m.ns.AddUserSource(ctx, user, friend)
			if err != nil {
				m.logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate friend %s for %s", friend, user)))
			} else {
				m.logger.Debug(fmt.Sprintf("Friend %s for %s migrated", friend, user))
			}
		}

		if len(friends) < take {
			break
		}
		skip += take
	}
}

func (m Migrator) migrateOutgoingRequests(ctx context.Context, user string) {
	skip := 0
	take := 20

	for {
		rs, err := m.rc.GetOutgoingRequests(user, skip, take)
		if err != nil {
			m.logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch outgoing requests for %s with skip: %d, count: %d", user, skip, take)))
			return
		}

		if len(rs) == 0 {
			break
		}

		for _, r := range rs {
			err := m.ns.AddUserSource(ctx, user, r)
			if err != nil {
				m.logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate outgoing request from %s to %s", user, r)))
			} else {
				m.logger.Debug(fmt.Sprintf("Outgoing request from %s to %s migrated", user, r))
			}
		}

		if len(rs) < take {
			break
		}

		skip += take
	}
}
