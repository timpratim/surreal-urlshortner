package repository

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	surreal "github.com/surrealdb/surrealdb.go"
)

var log = logger.New()

type ShortenerRepository struct {
	db *surreal.DB
}

func NewShortenerRepository(address, user, password, namespace, database string) (*ShortenerRepository, error) {
	db, err := surreal.New(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}
	_, err = db.Signin(map[string]interface{}{
		"user": user,
		"pass": password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign in: %w", err)
	}

	_, err = db.Use(namespace, database)
	if err != nil {
		return nil, err
	}

	return &ShortenerRepository{db}, nil
}

func (r ShortenerRepository) Close() {
	r.db.Close()
}

func (r ShortenerRepository) CreateShortUrl(original string, shortened string) (interface{}, error) {
	return r.db.Create("urls", map[string]interface{}{
		"original":  original,
		"shortened": shortened,
	})
}

func (r ShortenerRepository) FindShortenedURL(id string) (interface{}, error) {
	return r.db.Query("SELECT * FROM urls WHERE shortened = $shortened limit 1", map[string]interface{}{
		"shortened": "http://localhost:8090/" + id,
	})
}
