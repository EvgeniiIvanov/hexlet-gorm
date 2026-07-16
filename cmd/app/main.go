package main

import (
	"log"
	"os"
	"strconv"
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
	releasedAt, err := time.Parse("2006-01-02", args[5])
	if err != nil {
		log.Fatal(err)
	}
	movie := models.Movie{
		Title:      args[3],
		Genre:      args[4],
		ReleasedAt: releasedAt,
	}
	if err := db.Create(&movie).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("created movie id=%d", movie.ID)
}

func handleShow(db *gorm.DB, args []string) {
	var movie models.Movie
	if err := db.Preload("Director").Preload("Actors").First(&movie, args[3]).Error; err != nil {
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
	if err := db.Model(&models.Movie{}).
		Where("id = ?", args[3]).
		Update(args[4], args[5]).Error; err != nil {
		log.Fatal(err)
	}
	log.Println("movie updated")
}

func handleDelete(db *gorm.DB, args []string) {
	if err := db.Delete(&models.Movie{}, args[3]).Error; err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	score, err := strconv.Atoi(args[4])
	if err != nil {
		log.Fatal(err)
	}
	text := args[5]

	err = db.Transaction(func(tx *gorm.DB) error {
		review := models.Review{
			MovieID: uint(movieID),
			Score:   score,
			Text:    text,
		}
		if err := tx.Create(&review).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Movie{}).
			Where("id = ?", movieID).
			Update("reviews_count", gorm.Expr("reviews_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
