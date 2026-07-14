package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

func main() {
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

	if err := db.AutoMigrate(&models.Movie{}); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	var movie models.Movie
	if err := db.First(&movie, 2).Error; err != nil {
		log.Fatalf("ошибка чтения: %v", err)
	}

	log.Printf("фильм загружен: %s <%s>, рейтинг: %f", movie.Title, movie.Genre, movie.Rating)
}
