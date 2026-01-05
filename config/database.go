package config

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	// Pastikan selalu pakai .env (override env yang sudah ada)
	_ = godotenv.Overload(".env")

	host := mustEnv("DB_HOST")
	port := mustEnv("DB_PORT")
	user := mustEnv("DB_USER")
	pass := os.Getenv("DB_PASSWORD") // boleh kosong
	dbname := mustEnv("DB_NAME")

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   dbname,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	// Log tanpa password
	log.Printf("DB_URL=%s://%s@%s/%s?sslmode=disable", u.Scheme, user, u.Host, dbname)

	log.Println("🔌 Connecting to database...")
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{})
	if err != nil {
		DB = nil
		return fmt.Errorf("failed to connect database: %w", err)
	}

	DB = db
	log.Println("✅ Database connected successfully!")
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("❌ Missing required env: %s (check .env)", key)
	}
	return v
}
