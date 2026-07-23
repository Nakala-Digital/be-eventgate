package config

import (
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    ServerPort string
    ServerEnv  string
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    DBSSLMode  string
}

func LoadConfig() (*Config, error) {
    _ = godotenv.Load()

    cfg := &Config{
        ServerPort: getEnv("SERVER_PORT", "8080"),
        ServerEnv:  getEnv("SERVER_ENV", "development"),
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", "postgres"),
        DBName:     getEnv("DB_NAME", "eventgate_db"),
        DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
    }

    return cfg, nil
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}
