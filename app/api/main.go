package main

import (
	"context"
	"encoding/json"
	"github.com/ghosts-network/news-feed/news"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	storage := configureNewsStorage(os.Getenv("MONGO_CONNECTION"))

	r := mux.NewRouter()
	r.HandleFunc("/{user}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("Incoming http request %v\n", r.RequestURI)
		user := mux.Vars(r)["user"]
		cursor := mux.Vars(r)["cursor"]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ns, err := storage.FindNews(ctx, user, cursor)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		body, err := json.Marshal(ns)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write(body)
	}).Methods(http.MethodGet)

	log.Println("Starting http server on port 10000")
	log.Fatal(http.ListenAndServe(":10000", r))
}

type NewsStorage interface {
	FindNews(ctx context.Context, user string, cursor string) ([]news.Publication, error)
}

func configureNewsStorage(connectionString string) NewsStorage {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	return news.NewMongoNewsStorage(mongoClient)
}
