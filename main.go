package main

import (
	"encoding/json"
	"errors"
	"fmt"
	logger "github.com/sirupsen/logrus"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"math/rand"
	"net/http"
	"os"
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

func shortenURL(url string) string {
	s := ""
	//rand.Intn(26) returns a random number between 0 and 25. 97 is the ascii value of 'a'. So rand.Intn(26) + 97 returns a random lowercase letter.
	for i := 0; i < 6; i++ {
		s += string(rand.Intn(26) + 97)
	}

	shortendURL := fmt.Sprintf("http://localhost:8090/%s", s)
	return shortendURL
}

func Connect() (*surrealdb.DB, error) {
	url := os.Getenv("SURREALDB_URL")
	if url == "" {
		url = "ws://localhost:8000/rpc"
	}

	db, err := surrealdb.New(url)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func signin(db *surrealdb.DB) (interface{}, error) {
	signin, err := db.Signin(map[string]interface{}{
		"user": "root",
		"pass": "root",
	})

	if err != nil {
		log.Errorf("failed to signin: %+v", err)
		return nil, err
	}

	return signin, nil
}

func redirectURL(db *surrealdb.DB, w http.ResponseWriter, r *http.Request) {

	id := r.URL.Path[1:]
	log.Tracef("Generating redirect URL for %s", id)

	db.Use("test", "test")
	_, err := signin(db)
	if err != nil {
		panic("failed to signin")
	}
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
	db, err := Connect()
	if err != nil {
		panic("failed to connect database")
	}
	defer func() {
		log.Infof("closing database")
		db.Close()
	}()
	log.Infof("Connected to database")

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		original := r.FormValue("url")
		shortened := shortenURL(original)
		log.Tracef("shortened url: %s", shortened)
		//db.Create(&URL{Original: original, Shortened: shortened})
		db.Use("test", "test")
		_, err := signin(db)
		if err != nil {
			panic("failed to signin")
		}

		urlMap, err := db.Create("urls", map[string]interface{}{
			"original":  original,
			"shortened": shortened,
		})
		log.Tracef("created url mapping: %+v", urlMap)

		if err != nil || urlMap == nil {
			panic("failed to create user")
		}
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
