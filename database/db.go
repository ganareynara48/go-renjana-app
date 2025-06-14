package database

import (
	"fmt"
	"os"

	"renjana-app/models" // ganti dengan path sebenarnya

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Tambahkan logger
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // atau logger.Warn / logger.Error
	})
	if err != nil {
		panic("Failed to connect to PostgreSQL")
	}

	// Logging migration mulai
	fmt.Println("Running auto migration...")

	if err := models.MigrateAll(db); err != nil {
		panic("Failed to auto migrate models: " + err.Error())
	}

	DB = db
}
