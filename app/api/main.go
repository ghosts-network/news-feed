package main

import (
	"context"
	"encoding/json"
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/migrator"
	"github.com/ghosts-network/news-feed/news"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	profileClient := infrastructure.NewProfilesClient(os.Getenv("PROFILES_ADDRESS"))
	relationsClient := infrastructure.NewRelationsClient(os.Getenv("PROFILES_ADDRESS"))
	publicationsClient := infrastructure.NewPublicationsClient(os.Getenv("CONTENT_ADDRESS"))
	newsStorage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))

	m := migrator.NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage)

	r := mux.NewRouter()
	r.HandleFunc("/{user}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("[INFO] Incoming http request %v\n", r.RequestURI)
		user := mux.Vars(r)["user"]
		cursor := mux.Vars(r)["cursor"]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ns, err := newsStorage.FindNews(ctx, user, cursor)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ps := make([]infrastructure.Publication, 0, len(ns))
		for _, p := range ns {
			ps = append(ps, infrastructure.Publication{
				Id:        p.Id,
				Content:   p.Content,
				Author:    p.Author,
				CreatedOn: time.UnixMilli(p.CreatedOn).In(time.UTC),
				UpdatedOn: time.UnixMilli(p.UpdatedOn).In(time.UTC),
				Media:     p.Media,
			})
		}

		body, err := json.Marshal(ps)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write(body)
	}).Methods(http.MethodGet)

	r.HandleFunc("/migrator/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("[INFO] Incoming http request %v\n", r.RequestURI)
		go m.MigrateUsers()
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	r.HandleFunc("/migrator/users/{user}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("[INFO] Incoming http request %v\n", r.RequestURI)
		user := mux.Vars(r)["user"]
		go m.MigrateUser(user)
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	r.HandleFunc("/migrator/publications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("[INFO] Incoming http request %v\n", r.RequestURI)
		go m.MigratePublications()
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPost)

	log.Println("[INFO] Starting http server on port 10000")
	log.Fatal(http.ListenAndServe(":10000", r))
}
