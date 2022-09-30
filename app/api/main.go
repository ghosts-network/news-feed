package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/migrator"
	"github.com/ghosts-network/news-feed/news"
	"github.com/ghosts-network/news-feed/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"strconv"
	"time"
)

var log *utils.Logger

func main() {
	log = utils.NewLogger("news-feed-api")

	profileClient := infrastructure.NewProfilesClient(os.Getenv("PROFILES_ADDRESS"))
	relationsClient := infrastructure.NewRelationsClient(os.Getenv("PROFILES_ADDRESS"))
	publicationsClient := infrastructure.NewPublicationsClient(os.Getenv("CONTENT_ADDRESS"))
	newsStorage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))

	r := mux.NewRouter()
	r.HandleFunc("/{user}", func(w http.ResponseWriter, r *http.Request) {
		l := r.Context().Value("logger").(*utils.Logger)
		user := mux.Vars(r)["user"]
		cursor := r.URL.Query().Get("cursor")
		take, _ := strconv.Atoi(r.URL.Query().Get("take"))
		if 0 < take || take > 100 {
			take = 20
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ps, err := newsStorage.FindNews(ctx, user, cursor, take)
		if err != nil {
			l.Error(errors.Wrap(err, fmt.Sprintf("Failed to fetch news")))
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		body, err := json.Marshal(ps)
		if err != nil {
			l.Error(errors.Wrap(err, fmt.Sprintf("Failed to marshal news")))
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
		go migrator.
			NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage,
				r.Context().Value("logger").(*utils.Logger)).
			MigrateUsers()
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	r.HandleFunc("/migrator/users/{user}", func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]
		go migrator.
			NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage,
				r.Context().Value("logger").(*utils.Logger)).
			MigrateUser(user)
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	r.HandleFunc("/migrator/publications", func(w http.ResponseWriter, r *http.Request) {
		go migrator.
			NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage,
				r.Context().Value("logger").(*utils.Logger)).
			MigratePublications()
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	r.Use(scopedLoggerMiddleware)
	r.Use(loggingMiddleware)
	r.Use(setJsonContentType)

	log.Info("Starting http server on port 10000")
	log.Error(http.ListenAndServe(":10000", r))
}

func scopedLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}

		newContext := context.WithValue(r.Context(), "logger", log.WithValue("operationId", id))
		next.ServeHTTP(w, r.WithContext(newContext))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		scopedLog := r.Context().Value("logger").(*utils.Logger).
			Info(fmt.Sprintf("%s %s request started", r.Method, r.RequestURI))

		sw := NewStatusWriter(w)
		next.ServeHTTP(sw, r)

		scopedLog.
			WithValue("statusCode", sw.Status).
			WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
			Info(fmt.Sprintf("%s %s request finished", r.Method, r.RequestURI))
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
