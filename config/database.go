package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env file not found, using environment variables")
	} else {
		log.Println("✅ .env file loaded")
	}

	// ===== CEK MODE DATABASE =====
	useDB := getEnv("USE_DATABASE", "true")

	// Skip database jika USE_DATABASE = false
	if useDB == "false" {
		log.Println("⚠️  Database DISABLED - Running in Frontend Only Mode")
		log.Println("💡 Set USE_DATABASE=true di .env untuk enable database")
		DB = nil // Set DB ke nil supaya bisa dicek di tempat lain
		return nil
	}

	// ===== CONNECT TO DATABASE =====
	log.Println("🔌 Connecting to database...")

	// Get configuration from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "hris_db")

	// Build connection string
	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("❌ Failed to connect to database:", err)
		log.Println("⚠️  Server will run WITHOUT database")
		DB = nil
		return err // Return error tapi jangan Fatal
	}

	DB = database
	log.Println("✅ Database connected successfully!")
	return nil
}

// Helper function to get environment variable with default value
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
