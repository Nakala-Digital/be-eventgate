package config

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/lib/pq"
)

func InitDB(cfg *Config) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
    )

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(15 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("gagal koneksi/ping ke database: %w", err)
    }

    log.Println("Berhasil terhubung ke database PostgreSQL")
    return db, nil
}
