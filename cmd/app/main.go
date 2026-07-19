package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/db"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/requests"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: movies <list|create|show|update|delete> [args]")
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal("ошибка загрузки .env file")
	}

	// Connect to database
	database := db.Connect()

	entity := os.Args[1]
	action := os.Args[2]
	if entity != "movies" {
		log.Fatal("only movies supported")
	}

	switch action {
	case "list":
		handleList(database)
	case "create":
		handleCreate(database, os.Args)
	case "show":
		handleShow(database, os.Args)
	case "update":
		handleUpdate(database, os.Args)
	case "delete":
		handleDelete(database, os.Args)
	case "unrated":
		handleUnrated(database)
	case "most_rated":
		handleMostRated(database)
	case "add_review":
		handleAddReview(database, os.Args)
	case "rating":
		handleRating(database)
	case "leaderboard":
		handleLeaderboard(database)
	default:
		log.Fatal("unknown action")
	}
}

func handleList(db *gorm.DB) {
	movies, err := requests.ListMovies(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, movie := range movies {
		log.Printf("movie: %s", movie.Title)
		if movie.Director.ID != 0 {
			log.Printf("director: %s", movie.Director.Name)
		}
	}
	log.Printf("movies: %d", len(movies))
}

func handleCreate(db *gorm.DB, args []string) {
	if len(args) < 6 {
		log.Fatal("usage: movies create <title> <genre> <released_at>")
	}

	title := args[3]
	genre := args[4]

	// Validate movie data
	if err := validateMovie(title, genre); err != nil {
		log.Fatal(err)
	}

	releasedAt, err := time.Parse("2006-01-02", args[5])
	if err != nil {
		log.Fatal("invalid date format, use YYYY-MM-DD")
	}

	movie, err := requests.CreateMovie(db, title, genre, releasedAt)
	if err != nil {
		if isDuplicateError(err) {
			log.Fatal("movie with this title already exists")
		}
		log.Fatal(err)
	}

	log.Printf("created movie id=%d", movie.ID)
}

func handleShow(db *gorm.DB, args []string) {
	if len(args) < 4 {
		log.Fatal("usage: movies show <id>")
	}

	id, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		log.Fatal("invalid movie ID")
	}

	movie, err := requests.GetMovie(db, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal("movie not found")
		}
		log.Fatal(err)
	}

	log.Printf("movie: %s (%s)", movie.Title, movie.Genre)
	if movie.Director.Name != "" {
		log.Printf("director: %s", movie.Director.Name)
	}
	for _, actor := range movie.Actors {
		log.Printf("actor: %s", actor.Name)
	}
}

func handleUpdate(db *gorm.DB, args []string) {
	if len(args) < 6 {
		log.Fatal("usage: movies update <id> <field> <value>")
	}

	id, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		log.Fatal("invalid movie ID")
	}

	rowsAffected, err := requests.UpdateMovie(db, uint(id), args[4], args[5])
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		log.Fatal("movie not found")
	}

	log.Println("movie updated")
}

func handleDelete(db *gorm.DB, args []string) {
	if len(args) < 4 {
		log.Fatal("usage: movies delete <id>")
	}

	id, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		log.Fatal("invalid movie ID")
	}

	rowsAffected, err := requests.DeleteMovie(db, uint(id))
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		log.Fatal("movie not found")
	}

	log.Println("movie deleted")
}

func handleUnrated(db *gorm.DB) {
	movies, err := requests.GetUnratedMovies(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, movie := range movies {
		log.Printf("movie: %s", movie.Title)
	}
	log.Printf("unrated movies: %d", len(movies))
}

func handleMostRated(db *gorm.DB) {
	movies, err := requests.GetMostRatedMovies(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, movie := range movies {
		log.Printf("movie: %s", movie.Title)
	}
	log.Printf("most rated movies: %d", len(movies))
}

func handleAddReview(db *gorm.DB, args []string) {
	if len(args) < 6 {
		log.Fatal("usage: movies add_review <movie_id> <score> <text>")
	}

	movieID, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		log.Fatal("invalid movie ID")
	}

	score, err := strconv.Atoi(args[4])
	if err != nil {
		log.Fatal("invalid score")
	}
	text := args[5]

	// Validate review data
	if err := validateReview(score, text); err != nil {
		log.Fatal(err)
	}

	// Check if movie exists
	exists, err := requests.MovieExists(db, uint(movieID))
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Fatal("movie not found")
	}

	// Create the review - AfterCreate hook will update movie rating automatically
	if err := requests.CreateReview(db, uint(movieID), score, text); err != nil {
		if isDuplicateError(err) {
			log.Fatal("duplicate review: a review with this text already exists for this movie")
		}
		log.Fatal(err)
	}

	log.Println("review added and movie rating updated")
}

func handleRating(db *gorm.DB) {
	movies, err := requests.GetMoviesWithRatings(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, movie := range movies {
		log.Printf("movie: %s, rating: %.2f, reviews: %d", movie.Title, movie.Rating, movie.ReviewsCount)
	}
}

func handleLeaderboard(db *gorm.DB) {
	movies, err := requests.GetMoviesLeaderboard(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, movie := range movies {
		log.Printf("rank: %d, movie: %s, rating: %.2f, reviews: %d",
			movie.Rank, movie.Title, movie.Rating, movie.ReviewsCount)
	}
}

// Validation functions
func validateMovie(title, genre string) error {
	if title == "" {
		return errors.New("title is required")
	}
	if genre == "" {
		return errors.New("genre is required")
	}
	return nil
}

func validateReview(score int, text string) error {
	if score < 1 || score > 10 {
		return errors.New("score must be between 1 and 10")
	}
	if text == "" {
		return errors.New("text is required")
	}
	if len(text) > 255 {
		return errors.New("text is too long")
	}
	return nil
}

// Helper to check for duplicate key errors
func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
