package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

// Connect establishes a database connection and runs migrations
func Connect() *gorm.DB {
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

	// Run migrations
	if err := db.AutoMigrate(&models.Movie{}, &models.Actor{}, &models.Director{}, &models.Review{}); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	return db
}
