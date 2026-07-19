package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: movies <list|create|show|update|delete> [args]")
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("ошибка загрузки .env file")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=gorm password=dev_password_123 dbname=gorm_dev port=5432 sslmode=disable"
	}

	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatalf("ошибка подключения: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("ошибка доступа к пулу: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("ошибка пинга базы: %v", err)
	}

	if err := db.AutoMigrate(&models.Movie{}, &models.Actor{}, &models.Director{}, &models.Review{}); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	entity := os.Args[1]
	action := os.Args[2]
	if entity != "movies" {
		log.Fatal("only movies supported")
	}

	switch action {
	case "list":
		handleList(db)
	case "create":
		handleCreate(db, os.Args)
	case "show":
		handleShow(db, os.Args)
	case "update":
		handleUpdate(db, os.Args)
	case "delete":
		handleDelete(db, os.Args)
	case "unrated":
		handleUnrated(db)
	case "most_rated":
		handleMostRated(db)
	case "add_review":
		handleAddReview(db, os.Args)
	case "rating":
		handleRating(db)
	case "leaderboard":
		handleLeaderboard(db)
	default:
		log.Fatal("unknown action")
	}
}

func handleList(db *gorm.DB) {
	var movies []models.Movie
	if err := db.Preload("Director").Find(&movies).Error; err != nil {
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

	movie := models.Movie{
		Title:      title,
		Genre:      genre,
		ReleasedAt: releasedAt,
	}

	if err := db.Create(&movie).Error; err != nil {
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

	var movie models.Movie
	if err := db.Preload("Director").Preload("Actors").First(&movie, args[3]).Error; err != nil {
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

	result := db.Model(&models.Movie{}).
		Where("id = ?", args[3]).
		Update(args[4], args[5])

	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if result.RowsAffected == 0 {
		log.Fatal("movie not found")
	}

	log.Println("movie updated")
}

func handleDelete(db *gorm.DB, args []string) {
	if len(args) < 4 {
		log.Fatal("usage: movies delete <id>")
	}

	result := db.Delete(&models.Movie{}, args[3])
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if result.RowsAffected == 0 {
		log.Fatal("movie not found")
	}

	log.Println("movie deleted")
}

func handleUnrated(db *gorm.DB) {
	var movies []models.Movie
	if err := db.Where("rating IS NULL").Find(&movies).Error; err != nil {
		log.Fatal(err)
	}
	for _, movie := range movies {
		log.Printf("movie: %s", movie.Title)
	}
	log.Printf("unrated movies: %d", len(movies))
}

func handleMostRated(db *gorm.DB) {
	var movies []models.Movie
	if err := db.Where("rating > ?", 8.5).Find(&movies).Error; err != nil {
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
	var movie models.Movie
	if err := db.First(&movie, movieID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal("movie not found")
		}
		log.Fatal(err)
	}

	// Create the review - AfterCreate hook will update movie rating automatically
	review := models.Review{
		MovieID: uint(movieID),
		Score:   score,
		Text:    text,
	}
	if err := db.Create(&review).Error; err != nil {
		if isDuplicateError(err) {
			log.Fatal("duplicate review: a review with this text already exists for this movie")
		}
		log.Fatal(err)
	}

	log.Println("review added and movie rating updated")
}

func handleRating(db *gorm.DB) {
	var movies []models.MovieRatingDTO
	if err := db.Raw(`
		SELECT 
			m.title, 
			AVG(r.score) as rating, 
			COUNT(r.id) as reviews_count 
		FROM movies m
		LEFT JOIN reviews r ON r.movie_id = m.id
		GROUP BY m.id, m.title
		ORDER BY rating DESC NULLS LAST, m.title;
	`).Scan(&movies).Error; err != nil {
		log.Fatal(err)
	}
	for _, movie := range movies {
		log.Printf("movie: %s, rating: %.2f, reviews: %d", movie.Title, movie.Rating, movie.ReviewsCount)
	}
}

func handleLeaderboard(db *gorm.DB) {
	var movies []models.MovieLeaderboardDTO
	if err := db.Raw(`
		SELECT
			m.title,
			COALESCE(AVG(r.score), 0) as rating,
			COUNT(r.id) as reviews_count,
			DENSE_RANK() OVER (ORDER BY COALESCE(AVG(r.score), 0) DESC) as rank
		FROM movies m
		LEFT JOIN reviews r ON r.movie_id = m.id
		GROUP BY m.id, m.title
		ORDER BY rating DESC, m.title;
	`).Scan(&movies).Error; err != nil {
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
