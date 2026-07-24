package main

import (
    "fmt"
    "log"
    "net/http"

    "be-eventgate/config"
    deliveryHTTP "be-eventgate/internal/delivery/http"

    intdb "be-eventgate/internal/database"
    introuter "be-eventgate/internal/router"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Gagal membaca konfigurasi: %v", err)
    }

    db, err := config.InitDB(cfg)
    if err != nil {
        log.Printf("Peringatan Database: %v (Aplikasi tetap berjalan)", err)
    } else {
        defer db.Close()
    }

    // EVG-41 GORM Init
    gormDB, err := intdb.Connect(*cfg)
    if err != nil {
        log.Fatalf("Gagal connect GORM: %v", err)
    }
    if err := intdb.Migrate(gormDB); err != nil {
        log.Fatalf("Gagal auto-migrate GORM: %v", err)
    }
    if err := intdb.SeedRoles(gormDB); err != nil {
        log.Fatalf("Gagal seed roles: %v", err)
    }

    router := deliveryHTTP.NewRouter()
    authRouter := introuter.New(gormDB, cfg.JWTSecret, cfg.JWTExpiryHrs)
    router.Mount("/", authRouter)

    serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
    log.Printf("Server EventGate berjalan di port %s [ENV: %s]", cfg.ServerPort, cfg.ServerEnv)

    if err := http.ListenAndServe(serverAddr, router); err != nil {
        log.Fatalf("Gagal menjalankan server: %v", err)
    }
}
