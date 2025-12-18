package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("⚠️  .env file not found, using default values")
    }
    
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
        log.Fatal("❌ Failed to connect to database:", err)
    }

    DB = database
    log.Println("✅ Database connected successfully!")
}

// Helper function to get environment variable with default value
func getEnv(key string, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}
