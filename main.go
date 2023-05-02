package main

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/timpratim/surreal-urlshortner/repository"
	"github.com/timpratim/surreal-urlshortner/web"
	"net/http"
)

const port = 8090

var redirectAddress = fmt.Sprintf("http://localhost:%d", port)

const url = "ws://localhost:8000/rpc"
const namespace = "surrealdb-conference-content"
const database = "urlshortner"

var log = logger.New()

func main() {
	log.SetLevel(logger.TraceLevel)
	// Create the database repository that uses SurrealDB to store information
	repository, err := repository.NewShortenerRepository(url, "root", "root", namespace, database)
	if err != nil {
		log.Fatalf("failed to create shortener repository: %+v", err)
	}
	log.Infof("Connected to database")
	// Close connections to the database at program shutdown
	defer func() {
		log.Infof("Closing database")
		repository.Close()
	}()

	// Create the web service
	ws := web.NewWebService(repository, redirectAddress)
	http.HandleFunc("/shorten", ws.ShortenURL)
	http.HandleFunc("/", ws.RedirectURL)
	log.Infof("Listening on port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("failed to listen: %+v", err)
	}
}
