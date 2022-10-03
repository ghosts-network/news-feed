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
	take := 100

	for {
		ps, err := m.pc.GetProfiles(ctx, skip, take)
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
	_ = m.ns.RemovePublications(ctx)

	var cursor string
	take := 100

	for {
		ps, nextCursor, err := m.pubsClient.GetPublications(ctx, cursor, take)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch publications with cursor: %s, count: %d", cursor, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(ps) == 0 {
			break
		}

		if err = m.ns.AddPublications(ctx, ps); err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate publications batch (%s, %d)", cursor, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
		} else {
			logger.Debug(fmt.Sprintf("Publications batch (%s, %d) migrated", cursor, take), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
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
	take := 100

	for {
		friends, err := m.rc.GetFriends(ctx, user, skip, take)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch friends for %s with skip: %d, count: %d", user, skip, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(friends) == 0 {
			break
		}

		if err = m.ns.AddUserSources(ctx, user, friends); err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate friends batch (%d, %d) for %s", skip, take, user)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
		} else {
			logger.Debug(fmt.Sprintf("Friends batch (%d, %d) for %s migrated", skip, take, user), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
		}

		if len(friends) < take {
			break
		}
		skip += take
	}
}

func (m Migrator) migrateOutgoingRequests(ctx context.Context, user string) {
	skip := 0
	take := 100

	for {
		rs, err := m.rc.GetOutgoingRequests(ctx, user, skip, take)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch outgoing requests for %s with skip: %d, count: %d", user, skip, take)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
			return
		}

		if len(rs) == 0 {
			break
		}

		if err = m.ns.AddUserSources(ctx, user, rs); err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to migrate outgoing requests batch (%d, %d) for %s", skip, take, user)), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
		} else {
			logger.Debug(fmt.Sprintf("Outgoing request batch (%d, %d) for %s migrated", skip, take, user), &map[string]any{
				"operationId": ctx.Value("operationId"),
			})
		}

		if len(rs) < take {
			break
		}

		skip += take
	}
}
