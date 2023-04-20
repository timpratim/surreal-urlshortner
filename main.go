package main

import (
	"encoding/json"
	"errors"
	"fmt"
	logger "github.com/sirupsen/logrus"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/timpratim/surreal-urlshortner/repository"
	"math/rand"
	"net/http"
)

type Result struct {
	URLs   []URL  `json:"result"`
	Status string `json:"status"`
	Time   string `json:"time"`
}

type URL struct {
	ID        string `json:"id"`
	Original  string `json:"original"`
	Shortened string `json:"shortened"`
}

var log = logger.New()

const port = 8090
const url = "ws://localhost:8000/rpc"
const namespace = "surrealdb-conference-content"
const database = "urlshortner"

func shortenURL(url string) string {
	s := ""
	//rand.Intn(26) returns a random number between 0 and 25. 97 is the ascii value of 'a'. So rand.Intn(26) + 97 returns a random lowercase letter.
	for i := 0; i < 6; i++ {
		s += string(rand.Intn(26) + 97)
	}

	shortendURL := fmt.Sprintf("http://localhost:8090/%s", s)
	return shortendURL
}

func redirectURL(db *surrealdb.DB, w http.ResponseWriter, r *http.Request) {

	id := r.URL.Path[1:]
	log.Tracef("Generating redirect URL for %s", id)

	data, err := db.Query("SELECT * FROM urls WHERE shortened = $shortened limit 1", map[string]interface{}{
		"shortened": "http://localhost:8090/" + id,
	})
	if err != nil {
		log.Errorf("failed to query shortened: %+v", err)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	//unmarshal the data
	var results []Result
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		panic(err)
	}

	if len(results) == 0 {
		panic(errors.New("no results found"))
	}
	something := results[0].URLs
	if len(something) == 0 {
		panic(errors.New("no results found"))
	}
	originalURL := something[0].Original
	log.Tracef("Original URL: %s", originalURL)
	//redirect to the original url
	http.Redirect(w, r, originalURL, http.StatusSeeOther)

}

func main() {
	repository, err := repository.NewShortenerRepository(url, "root", "root", namespace, database)
	if err != nil {
		log.Fatalf("failed to create shortener repository: %+v", err)
	}
	log.Infof("Connected to database")
	defer func() {
		log.Infof("Closing database")
		repository.Close()
	}()

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) error {
		original := r.FormValue("url")
		shortened := shortenURL(original)
		log.Tracef("shortened url: %s", shortened)
		//db.Create(&URL{Original: original, Shortened: shortened})

		urlMap, err := repository.CreateShortUrl(original, shortened)
		if err != nil {
			log.Errorf("failed to create short url: %+v", err)
			return err
		}
		log.Tracef("created url mapping: %+v", urlMap)

		// return json response with shortened url
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"shortened": shortened})

	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		redirectURL(db, w, r)
	})
	log.Infof("Listening on port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("failed to listen: %+v", err)
	}
}

//https://www.youtube.com/watch?v=4KfuQwB5rIs&t=1s
//curl -X POST -d "url=https://www.youtube.com/watch?v=4KfuQwB5rIs&t=1s" http://localhost:8080/shorten
