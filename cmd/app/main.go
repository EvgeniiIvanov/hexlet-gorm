package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID    uint
	Name  string
	Email string
}

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

	// Auto-migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	if err := db.Create(&User{Name: "Анна", Email: "anna@example.com"}).Error; err != nil {
		log.Fatalf("ошибка вставки: %v", err)
	}

	var user User
	if err := db.First(&user).Error; err != nil {
		log.Fatalf("ошибка чтения: %v", err)
	}

	log.Printf("пользователь загружен: %s <%s>", user.Name, user.Email)
}
