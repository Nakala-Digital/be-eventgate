# Panduan Menjalankan Sistem EventGate (Backend & Frontend)

Panduan ini ditujukan bagi anggota tim pengembang atau pengguna baru yang ingin menginstal, mengonfigurasi, dan menjalankan aplikasi **EventGate** secara lokal di komputernya masing-masing.

---

## 1. Prasyarat Sistem

Pastikan perangkat Anda telah terinstal perkakas berikut:
- **Golang**: Versi `1.20` atau lebih baru ([Download Go](https://go.dev/dl/))
- **PostgreSQL**: Versi `14` ke atas (Bisa via instalasi lokal / pgAdmin, atau menggunakan Docker)
- **Node.js**: Versi `18` ke atas & **npm** ([Download Node.js](https://nodejs.org/))
- **Git**: Untuk clone repositori

---

## 2. Persiapan Database PostgreSQL

Sebelum menjalankan Backend, pastikan service database PostgreSQL Anda telah aktif.

1. **Buat Database Baru**:
   Buka terminal/pgAdmin dan buat database baru bernama `eventgate_db`:
   ```sql
   CREATE DATABASE eventgate_db;
   ```

2. **Konfigurasi Environment Backend (`.env`)**:
   Masuk ke folder `be-eventgate`, salin `.env.example` menjadi `.env`:
   ```powershell
   Copy-Item .env.example .env
   ```
   Buka file `.env` dan sesuaikan kredensial dengan akun PostgreSQL Anda:
   ```env
   SERVER_PORT=8080
   SERVER_ENV=development

   DB_HOST=127.0.0.1
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=PasswordPostgresAnda
   DB_NAME=eventgate_db
   DB_SSLMODE=disable
   ```

---

## 3. Langkah Menjalankan Backend (`be-eventgate`)

Buka terminal di dalam folder repository **`be-eventgate`**, lalu jalankan perintah-perintah berikut secara berurutan:

### **Langkah 3.1: Eksekusi Migrasi Database**
Perintah ini akan membuat seluruh tabel, constraint, foreign key, dan enum status di database PostgreSQL:
```powershell
go run cmd/migrate/main.go up
```

### **Langkah 3.2: Seed Akun Pengguna Testing**
Perintah ini akan membuat 5 akun uji coba default dengan berbagai peran (RBAC):
```powershell
go run cmd/seeduser/main.go
```

> **Daftar Akun Pengujian yang Dihasilkan:**
> | Role | Email | Password | Hak Akses |
> | :--- | :--- | :--- | :--- |
> | **Super Admin** | `superadmin@eventgate.com` | `Rahasia123!` | Akses penuh, Approval, Publish, Kelola User |
> | **Admin Panitia** | `panitia@eventgate.com` | `Rahasia123!` | Buat Event, Edit Event Milik Sendiri, Submit Approval |
> | **School Reviewer** | `reviewer@eventgate.com` | `Rahasia123!` | Review, Approve, Reject, Request Revisi Event |
> | **Staf Lapangan** | `staf@eventgate.com` | `Rahasia123!` | Scan QR tiket & presensi peserta |

### **Langkah 3.3: Jalankan Server Backend**
```powershell
go run main.go
```
- Server Backend akan aktif di port: **`http://localhost:8080`**
- Verifikasi API Health: Buka browser di **`http://localhost:8080/api/v1/health`**

---

## 4. Langkah Menjalankan Frontend (`fe-eventgate`)

Buka terminal baru di folder repository **`fe-eventgate`**:

### **Langkah 4.1: Konfigurasi Environment Frontend (`.env`)**
Pastikan file `.env` mengarah ke URL Backend:
```env
VITE_API_BASE_URL=http://localhost:8080
```

### **Langkah 4.2: Install Dependencies & Jalankan Server Dev**
```powershell
npm install
npm run dev
```

Buka URL yang ditampilkan di terminal (biasanya **`http://localhost:5173`** atau **`http://localhost:3000`**) pada browser Anda, lalu login menggunakan salah satu akun di atas.

---

## 5. Panduan Kolaborasi Antar Rekan Tim

Jika ada rekan tim lain yang ingin menjalankan aplikasi:

### **Skenario A: Rekan Menjalankan Database Sendiri (Mandiri)**
Rekan Anda **tidak perlu meng-copy database dari komputer Anda**. Rekan Anda cukup:
1. Clone repositori `be-eventgate` & `fe-eventgate`.
2. Buat database lokal `eventgate_db` dan atur file `.env` di komputernya.
3. Jalankan perintah migrasi & seed:
   ```powershell
   go run cmd/migrate/main.go up
   go run cmd/seeduser/main.go
   ```
4. Jalankan `go run main.go` dan `npm run dev`.

### **Skenario B: Rekan Hanya Menjalankan Frontend (Tembak BE Anda via Ngrok)**
Jika rekan hanya memegang Frontend dan tidak ingin install database / Go:
1. Anda jalankan Backend di port 8080.
2. Expose port Anda menggunakan Ngrok:
   ```powershell
   ngrok http 8080
   ```
3. Berikan URL Ngrok publik yang muncul (contoh: `https://abcd-123.ngrok-free.app`) ke rekan Anda.
4. Rekan Anda cukup mengisi `.env` di Frontend-nya:
   ```env
   VITE_API_BASE_URL=https://abcd-123.ngrok-free.app
   ```
5. Rekan Anda tinggal menjalankan `npm run dev`.

---

## 6. Troubleshooting Umum

- **Status `404 Not Found` pada Endpoint API:**
  Pastikan base URL di Frontend memuat `/api/v1` atau mengarah ke root `http://localhost:8080` (Backend sudah mendukung routing `/api/v1/*` dan `/api/*`).
- **Database Connection Error:**
  Pastikan password dan port PostgreSQL di `.env` sudah benar serta service PostgreSQL Windows / Docker dalam status *Running*.
- **Port Conflict (Port 8080 sudah terpakai):**
  Ubah nilai `SERVER_PORT` di `.env` (misal ke `8081`) dan sesuaikan `VITE_API_BASE_URL` di Frontend menjadi `http://localhost:8081`.
