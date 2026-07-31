package config

import (
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    ServerPort   string
    ServerEnv    string
    DBHost       string
    DBPort       string
    DBUser       string
    DBPassword   string
    DBName       string
    DBSSLMode    string
    JWTSecret    string
    JWTExpiryHrs int
}

func LoadConfig() (*Config, error) {
    _ = godotenv.Load()

    cfg := &Config{
        ServerPort:   getEnv("SERVER_PORT", "8080"),
        ServerEnv:    getEnv("SERVER_ENV", "development"),
        DBHost:       getEnv("DB_HOST", "localhost"),
        DBPort:       getEnv("DB_PORT", "5432"),
        DBUser:       getEnv("DB_USER", "postgres"),
        DBPassword:   getEnv("DB_PASSWORD", "postgres"),
        DBName:       getEnv("DB_NAME", "eventgate_db"),
        DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
        JWTSecret:    getEnv("JWT_SECRET", "CHANGE_ME_SUPER_SECRET_IN_PRODUCTION"),
        JWTExpiryHrs: getEnvInt("JWT_EXPIRY_HOURS", 24),
    }

    return cfg, nil
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
