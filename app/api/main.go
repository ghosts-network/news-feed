package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/migrator"
	"github.com/ghosts-network/news-feed/news"
	"github.com/ghosts-network/news-feed/utils/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	log.SetFlags(0)

	newsStorage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))

	r := mux.NewRouter()
	r.HandleFunc("/{user}", func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]
		cursor := r.URL.Query().Get("cursor")
		take, _ := strconv.Atoi(r.URL.Query().Get("take"))
		if 0 < take || take > 100 {
			take = 20
		}

		ps, err := newsStorage.FindNews(r.Context(), user, cursor, take)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch news")), &map[string]any{
				"correlationId": r.Context().Value("correlationId"),
			})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		body, err := json.Marshal(ps)
		if err != nil {
			logger.Error(errors.Wrap(err, fmt.Sprintf("Failed to marshal news")), &map[string]any{
				"correlationId": r.Context().Value("correlationId"),
			})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(ps) > 0 {
			c := ps[len(ps)-1].Id
			w.Header().Set("X-Cursor", c)
		}
		_, _ = w.Write(body)
	}).Methods(http.MethodGet)

	r.HandleFunc("/migrator/users", func(w http.ResponseWriter, r *http.Request) {
		getMigrator().
			MigrateUsers(r.Context())

		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	r.HandleFunc("/migrator/users/{user}", func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]

		getMigrator().
			MigrateUser(r.Context(), user)

		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost)

	r.HandleFunc("/migrator/publications", func(w http.ResponseWriter, r *http.Request) {
		getMigrator().
			MigratePublications(r.Context())

		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost)

	r.Use(scopedLoggerMiddleware)
	r.Use(loggingMiddleware)
	r.Use(setJsonContentType)

	logger.Info("Starting http server on port 10000", &map[string]any{})
	logger.Error(http.ListenAndServe(":10000", r), &map[string]any{})
}

func scopedLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)

		newContext := context.WithValue(r.Context(), "correlationId", id)
		next.ServeHTTP(w, r.WithContext(newContext))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()

		logger.Info(fmt.Sprintf("%s %s request started", r.Method, r.RequestURI), &map[string]any{
			"correlationId": r.Context().Value("correlationId"),
			"calledId":      r.Header.Get("X-Called-ID"),
			"type":          "incoming:http",
		})

		sw := NewStatusWriter(w)
		next.ServeHTTP(sw, r)

		logger.Info(fmt.Sprintf("%s %s request finished", r.Method, r.RequestURI), &map[string]any{
			"correlationId":       r.Context().Value("correlationId"),
			"type":                "incoming:http",
			"statusCode":          sw.Status,
			"elapsedMilliseconds": time.Now().Sub(st).Milliseconds(),
		})
	})
}

func setJsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type StatusWriter struct {
	http.ResponseWriter
	Status int
}

func NewStatusWriter(w http.ResponseWriter) *StatusWriter {
	return &StatusWriter{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
}

func (w *StatusWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func getMigrator() *migrator.Migrator {
	httpClient := infrastructure.NewScopedClient()

	profileClient := infrastructure.NewProfilesClient(os.Getenv("PROFILES_ADDRESS"), httpClient)
	relationsClient := infrastructure.NewRelationsClient(os.Getenv("PROFILES_ADDRESS"), httpClient)
	publicationsClient := infrastructure.NewPublicationsClient(os.Getenv("CONTENT_ADDRESS"), httpClient)
	newsStorage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))

	return migrator.NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage)
}
