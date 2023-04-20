package web

import (
	"encoding/json"
	"errors"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/timpratim/surreal-urlshortner/repository"
	"math/rand"
	"net/http"
)

// webService is used to handle web requests via it's public methods
type webService struct {
	repository      *repository.ShortenerRepository
	redirectAddress string
}

// URL is a representation of a shortened URL
type URL struct {
	ID        string `json:"id"`
	Original  string `json:"original"`
	Shortened string `json:"shortened"`
}

// Result is the raw representation of the response from the database
type Result struct {
	URLs   []URL  `json:"result"`
	Status string `json:"status"`
	Time   string `json:"time"`
}

var log = logger.New()

func NewWebService(r *repository.ShortenerRepository, redirectAddress string) *webService {
	return &webService{
		repository:      r,
		redirectAddress: redirectAddress,
	}
}

func (ws webService) ShortenURL(writer http.ResponseWriter, request *http.Request) {
	original := request.FormValue("url")
	shortened := shortenURL(original)
	log.Tracef("shortened url: %s", shortened)

	urlMap, err := ws.repository.CreateShortUrl(original, shortened)
	if err != nil {
		log.Fatalf("failed to create short url: %+v", err) // TODO handle this better
	}
	log.Tracef("created url mapping: %+v", urlMap)

	// return json response with shortened url
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]string{"shortened": shortened})

}

func (ws webService) RedirectURL(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Path[1:]
	log.Tracef("Generating redirect URL for %s", id)

	data, err := ws.repository.FindShortenedURL(id)
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
	http.Redirect(writer, request, originalURL, http.StatusSeeOther)

}

func shortenURL(redirectUrl string) string {
	s := ""
	//rand.Intn(26) returns a random number between 0 and 25. 97 is the ascii value of 'a'. So rand.Intn(26) + 97 returns a random lowercase letter.
	for i := 0; i < 6; i++ {
		s += string(rand.Intn(26) + 97)
	}

	shortendURL := fmt.Sprintf("%s/%s", redirectUrl, s)
	return shortendURL
}
