package main

import (
	"github.com/ghosts-network/news-feed/infrastructure"
	"github.com/ghosts-network/news-feed/migrator"
	"github.com/ghosts-network/news-feed/news"
	"log"
	"os"
	"time"
)

func main() {
	profileClient := infrastructure.NewProfilesClient("http://localhost:5000")
	relationsClient := infrastructure.NewRelationsClient("http://localhost:5000")
	publicationsClient := infrastructure.NewPublicationsClient("http://localhost:5100")

	newsStorage := news.NewMongoNewsStorage(os.Getenv("MONGO_CONNECTION"))

	start := time.Now()
	m := migrator.NewMigrator(profileClient, relationsClient, publicationsClient, newsStorage)
	//m.MigrateUsers()
	m.MigratePublications()
	log.Printf("[INFO] Migration finished in %s\n", time.Now().Sub(start))
}
