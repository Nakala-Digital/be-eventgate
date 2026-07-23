# EventGate Backend (`be-eventgate`)

Backend service untuk Event Management System (EventGate) yang dibangun menggunakan Golang (Go), Chi Router, dan PostgreSQL.

---

## 📚 Project Documentation

Dokumentasi lengkap proyek dapat ditemukan di folder [`docs/`](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/):

- 📄 **[EVG-34: Data Entity & Relationship Specification](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-34_Data_Entity_Relationship.md)**  
  Spesifikasi lengkap skema database, data dictionary (14 entities), aturan relasi, pertimbangan teknis backend, dan ERD.
- ⚙️ **[EVG-38: Backend Project Structure & API Foundation Setup](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-38_Backend_Setup.md)**  
  Panduan arsitektur Clean Architecture, struktur direktori, konfigurasi `.env`, PostgreSQL connection pool, middleware, dan health check API.

---

## 🚀 Quickstart Guide

### 1. Prasyarat System
- **Go**: Version 1.20+
- **PostgreSQL**: Terinstal lokal / Docker container

### 2. Jalankan Server
```powershell
# Copy environment file
Copy-Item .env.example .env

# Unduh dependensi & jalankan server
go mod tidy
go run main.go
```

### 3. Health Check Verification
Akses endpoint health check melalui browser atau API client:
`GET http://localhost:8080/api/v1/health`