package main

import (
    "fmt"
    "log"
    "net/http"

    "be-eventgate/config"
    deliveryHTTP "be-eventgate/internal/delivery/http"
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

    router := deliveryHTTP.NewRouter()

    serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
    log.Printf("Server EventGate berjalan di port %s [ENV: %s]", cfg.ServerPort, cfg.ServerEnv)

    if err := http.ListenAndServe(serverAddr, router); err != nil {
        log.Fatalf("Gagal menjalankan server: %v", err)
    }
}
