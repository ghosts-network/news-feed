package migrator

import (
	"context"
	"fmt"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/news"
	"github.com/ghosts-network/news-feed/utils/logger"
	"github.com/pkg/errors"
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

func (m Migrator) MigrateUsers(ctx context.Context) {
	st := time.Now()

	skip := 0
	take := 20

	for {
		ps, err := m.pc.GetProfiles(skip, take)
		if err != nil {
			logger.Error(err, &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(ps) == 0 {
			break
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(ps))
		for _, profile := range ps {
			go m.MigrateUserAsync(ctx, profile.Id, wg)
		}
		wg.Wait()

		if len(ps) < take {
			break
		}

		skip += take
	}

	logger.Info("Users migration finished", &map[string]any{
		"operationId":         ctx.Value("operationId"),
		"elapsedMilliseconds": time.Now().Sub(st).Milliseconds(),
	})
}

func (m Migrator) MigrateUser(ctx context.Context, user string) {
	st := time.Now()

	_ = m.ns.RemoveUserSources(ctx, user)
	m.migrateFriends(ctx, user)
	m.migrateOutgoingRequests(ctx, user)

	logger.Info(fmt.Sprintf("User %s migration finished", user), &map[string]any{
		"operationId":         ctx.Value("operationId"),
		"elapsedMilliseconds": time.Now().Sub(st).Milliseconds(),
	})
}

func (m Migrator) MigrateUserAsync(ctx context.Context, user string, wg *sync.WaitGroup) {
	_ = m.ns.RemoveUserSources(ctx, user)
	m.migrateFriends(ctx, user)
	m.migrateOutgoingRequests(ctx, user)
	wg.Done()

	logger.Debug(fmt.Sprintf("User %s migration finished", user), &map[string]any{
		"operationId": ctx.Value("operationId"),
	})
}

func (m Migrator) MigratePublications(ctx context.Context) {
	st := time.Now()
	_ = m.ns.RemoveAllNews(ctx)

	var cursor string
	take := 20

	for {
		ps, nextCursor, err := m.pubsClient.GetPublications(cursor, take)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch publications with cursor: %s, count: %d", cursor, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(ps) == 0 {
			break
		}

		for _, publication := range ps {
			err = m.ns.AddPublication(ctx, &publication)

			if err != nil {
				logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate publication %s", publication.Id)), &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
			} else {
				logger.Debug(fmt.Sprintf("Publication %s migrated", publication.Id), &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
			}
		}

		if len(ps) < take {
			break
		}

		cursor = nextCursor
	}

	logger.Info("Publications migration finished", &map[string]any{
		"operationId":         ctx.Value("operationId"),
		"elapsedMilliseconds": time.Now().Sub(st).Milliseconds(),
	})
}

func (m Migrator) migrateFriends(ctx context.Context, user string) {
	skip := 0
	take := 20

	for {
		friends, err := m.rc.GetFriends(user, skip, take)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch friends for %s with skip: %d, count: %d", user, skip, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(friends) == 0 {
			break
		}

		for _, friend := range friends {
			err := m.ns.AddUserSource(ctx, user, friend)
			if err != nil {
				logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate friend %s for %s", friend, user)), &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
			} else {
				logger.Debug(fmt.Sprintf("Friend %s for %s migrated", friend, user), &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
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
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch outgoing requests for %s with skip: %d, count: %d", user, skip, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(rs) == 0 {
			break
		}

		for _, r := range rs {
			err := m.ns.AddUserSource(ctx, user, r)
			if err != nil {
				logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate outgoing request from %s to %s", user, r)), &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
			} else {
				logger.Debug(fmt.Sprintf("Outgoing request from %s to %s migrated", user, r), &map[string]any{
					"operationId": ctx.Value("operationId"),
				})
			}
		}

		if len(rs) < take {
			break
		}

		skip += take
	}
}
