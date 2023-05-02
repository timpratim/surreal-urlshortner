package web

import (
	"encoding/json"
	"errors"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/timpratim/surreal-urlshortner/repository"
	"math/rand"
	"net/http"
	"strings"
	"time"
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

var log = func() *logger.Logger {
	log := logger.New()
	log.SetLevel(logger.TraceLevel)
	return log
}()

func NewWebService(r *repository.ShortenerRepository, redirectAddress string) *webService {
	return &webService{
		repository:      r,
		redirectAddress: redirectAddress,
	}
}

// ShortenURL is used to create a shortened URL/
func (ws webService) ShortenURL(writer http.ResponseWriter, request *http.Request) {
	original := request.FormValue("url")
	if original == "" {
		badRequest(writer, errors.New("url is required"))
		return
	}
	if !strings.HasPrefix(original, "http://") && !strings.HasPrefix(original, "https://") {
		original = "https://" + original
	}
	shortened := shortenURL(ws.redirectAddress)
	log.Tracef("created shortened url '%s' for input '%s'", shortened, original)

	urlMap, err := ws.repository.CreateShortUrl(original, shortened)
	if err != nil {
		internalError(writer, fmt.Errorf("failed to create short url: %+v", err))
		return
	}
	log.Tracef("created url mapping: %+v", urlMap)

	// return json response with shortened url
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]string{"shortened": shortened, "original": original})
}

func (ws webService) RedirectURL(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Path[1:]
	log.Tracef("Generating redirect URL for %s", id)

	data, err := ws.repository.FindShortenedURL(id)
	if err != nil {
		internalError(writer, fmt.Errorf("failed to find shortened url: %+v", err))
		return
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		internalError(writer, fmt.Errorf("failed to marshal shortened url: %+v", err))
		return
	}
	//unmarshal the data
	var results []Result
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		internalError(writer, fmt.Errorf("failed to unmarshal shortened url: %+v", err))
		return
	}

	if len(results) == 0 {
		internalError(writer, errors.New("no results found"))
		return
	}
	something := results[0].URLs
	if len(something) == 0 {
		internalError(writer, errors.New("results did not contain any URLs"))
		return
	}
	originalURL := something[0].Original
	if originalURL == "" {
		internalError(writer, errors.New("original URL is empty"))
		return
	}
	log.Tracef("Translated short '%s' to original '%s'", id, originalURL)
	//redirect to the original url
	http.Redirect(writer, request, originalURL, http.StatusSeeOther)
}

// Helper code below this line
// ----------------------------------------------------------------------

func badRequest(writer http.ResponseWriter, cause error) {
	log.Errorf("bad request: %+v", cause)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]string{"error": "bad request",
		"cause": cause.Error()})
	writer.WriteHeader(http.StatusBadRequest)
}

func internalError(writer http.ResponseWriter, cause error) {
	log.Errorf("internal error: %+v", cause)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]string{"error": "internal error",
		"cause": cause.Error()})
	writer.WriteHeader(http.StatusInternalServerError)
}

func init() {
	rand.Seed(time.Now().UnixNano())
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
