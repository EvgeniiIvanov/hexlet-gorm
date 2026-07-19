package testutil

import (
	"time"

	"github.com/go-faker/faker/v4"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

// MovieFactory creates a movie for testing
func MovieFactory(title string) *models.Movie {
	if title == "" {
		title = faker.Word() + " " + faker.Word()
	}

	return &models.Movie{
		Title:       title,
		Genre:       "sci-fi",
		ReleasedAt:  time.Now().AddDate(-1, 0, 0),
		Description: faker.Sentence(),
	}
}

// MovieFactoryWithRandom creates a movie with random data
func MovieFactoryWithRandom() *models.Movie {
	return &models.Movie{
		Title:       faker.Word() + " " + faker.Word() + " " + faker.Word(),
		Genre:       faker.Word(),
		ReleasedAt:  time.Now().AddDate(-2, 0, 0),
		Description: faker.Sentence(),
	}
}

// ReviewFactory creates a review for testing
func ReviewFactory(movieID uint, score int, text string) *models.Review {
	if text == "" {
		text = faker.Sentence()
	}
	if score == 0 {
		score = 8
	}

	return &models.Review{
		MovieID: movieID,
		Score:   score,
		Text:    text,
	}
}

// DirectorFactory creates a director for testing
func DirectorFactory(name string) *models.Director {
	if name == "" {
		name = faker.FirstName() + " " + faker.LastName()
	}

	return &models.Director{
		Name: name,
	}
}

// ActorFactory creates an actor for testing
func ActorFactory(name string) *models.Actor {
	if name == "" {
		name = faker.FirstName() + " " + faker.LastName()
	}

	return &models.Actor{
		Name: name,
	}
}
