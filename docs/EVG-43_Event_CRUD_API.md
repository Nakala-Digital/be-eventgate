# Technical Documentation: EVG-43 [BE1] Event CRUD API Implementation

---

## Document Metadata

| Attribute         | Details                                                      |
| :---------------- | :----------------------------------------------------------- |
| **Project**       | EventGate — Event Management System                          |
| **Task Code**     | EVG-43 (Sprint 2)                                            |
| **Feature**       | Event CRUD API Implementation & Access Control               |
| **Tech Stack**    | Golang (Go), PostgreSQL, GORM, Chi Router                    |
| **Author**        | Backend Development Team                                     |
| **Status**        | Completed / Ready for Review                                 |
| **Reference Doc** | `docs/EVG-40_Database_Migration_Setup.md` & `docs/EVG-41_Auth_RBAC_Setup.md` |

---

## 1. Executive Summary

Task **EVG-43** mengimplementasikan rangkaian endpoint RESTful API untuk pengelolaan Event (Create, Read List, Read Detail, Update, dan Delete/Archive). Fitur ini dilengkapi dengan validasi field minimum (termasuk validasi harga untuk event berbayar dan rentang waktu pelaksanaan), pembuatan slug otomatis, penentuan status awal default (`draft`), serta penegakan otorisasi berbasis peran (RBAC) di mana **Super Admin** memiliki kendali penuh dan **Admin Panitia** terbatas pada pengelolaan event ciptaannya sendiri.

---

## 2. Minimum Field Event Specification

Setiap entitas Event memiliki field minimum sebagai berikut:

| Field Name | Type | Constraint | Description |
| :--- | :--- | :--- | :--- |
| `title` | String | Mandatory | Judul event. Auto-generate URL slug. |
| `description` | Text | Mandatory | Deskripsi lengkap event. |
| `banner` | String | Mandatory | URL banner / poster event. |
| `start_time` | ISO8601 | Mandatory | Waktu mulai event (`TIMESTAMPTZ`). |
| `end_time` | ISO8601 | Mandatory | Waktu selesai event (harus setelah `start_time`). |
| `location` | String | Mandatory | Lokasi fisik / platform pelaksanaan. |
| `is_paid` | Boolean | Mandatory | Status event (gratis / berbayar). |
| `price` | Numeric | Conditional | Harga tiket. Wajib > 0 jika `is_paid = true`. |
| `quota` | Integer | Mandatory | Kuota maksimum peserta (>= 0). |
| `status` | String | Default: `draft` | Status siklus event (`draft`, `pending_approval`, `approved`, dll). |
| `created_by` | Integer | System Generated | User ID pembuat event (diambil dari token JWT). |

---

## 3. Role Access Matrix for Event CRUD

| Endpoint | Method | Path | Allowed Roles | Ownership Restriction |
| :--- | :---: | :--- | :--- | :--- |
| **Create Event** | `POST` | `/api/v1/events` | `super_admin`, `admin_panitia` | None (User aktif diset sebagai `created_by` & `organizer_id`). |
| **List Events** | `GET` | `/api/v1/events` | All Authenticated Roles | None (Mendukung query parameter `search` & `status`). |
| **Get Detail** | `GET` | `/api/v1/events/{id}` | All Authenticated Roles | None. |
| **Update Event** | `PUT` | `/api/v1/events/{id}` | `super_admin`, `admin_panitia` | `admin_panitia` hanya dapat mengubah event miliknya (`created_by == userID`). `super_admin` bebas mengedit. |
| **Delete Event** | `DELETE` | `/api/v1/events/{id}` | `super_admin`, `admin_panitia` | `admin_panitia` hanya dapat menghapus event miliknya. `super_admin` bebas menghapus. |

---

## 4. API Reference

### 4.1. Create Event
* **URL:** `/api/v1/events` (atau `/api/events`)
* **Method:** `POST`
* **Auth Required:** Yes (Header `Authorization: Bearer <token>`)
* **Role:** `super_admin`, `admin_panitia`

**Request Body (JSON):**
```json
{
  "title": "Go & Microservices Workshop 2026",
  "description": "Pelatihan mendalam membangun microservices menggunakan Golang.",
  "banner": "https://cdn.eventgate.com/banners/go-workshop.jpg",
  "start_time": "2026-08-15T09:00:00Z",
  "end_time": "2026-08-15T16:00:00Z",
  "location": "Auditorium Lantai 3 JTK POLBAN",
  "is_paid": true,
  "price": 50000,
  "quota": 150,
  "status": "draft"
}
```

**Success Response (201 Created):**
```json
{
  "id": 1,
  "organizer_id": 2,
  "created_by": 2,
  "title": "Go & Microservices Workshop 2026",
  "description": "Pelatihan mendalam membangun microservices menggunakan Golang.",
  "banner": "https://cdn.eventgate.com/banners/go-workshop.jpg",
  "location": "Auditorium Lantai 3 JTK POLBAN",
  "slug": "go-microservices-workshop-2026-89123",
  "start_time": "2026-08-15T09:00:00Z",
  "end_time": "2026-08-15T16:00:00Z",
  "is_paid": true,
  "price": 50000,
  "quota": 150,
  "status": "draft",
  "created_at": "2026-07-24T16:30:00Z",
  "updated_at": "2026-07-24T16:30:00Z"
}
```

---

### 4.2. Get List Events
* **URL:** `/api/v1/events?search=workshop&status=draft`
* **Method:** `GET`
* **Auth Required:** Yes

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "title": "Go & Microservices Workshop 2026",
    "description": "Pelatihan mendalam membangun microservices menggunakan Golang.",
    "banner": "https://cdn.eventgate.com/banners/go-workshop.jpg",
    "location": "Auditorium Lantai 3 JTK POLBAN",
    "slug": "go-microservices-workshop-2026-89123",
    "start_time": "2026-08-15T09:00:00Z",
    "end_time": "2026-08-15T16:00:00Z",
    "is_paid": true,
    "price": 50000,
    "quota": 150,
    "status": "draft",
    "created_by": 2
  }
]
```

---

### 4.3. Get Event Detail
* **URL:** `/api/v1/events/1`
* **Method:** `GET`
* **Auth Required:** Yes

---

### 4.4. Update Event
* **URL:** `/api/v1/events/1`
* **Method:** `PUT`
* **Auth Required:** Yes
* **Role:** `super_admin` atau `admin_panitia` (pemilik event)

---

### 4.5. Delete Event
* **URL:** `/api/v1/events/1`
* **Method:** `DELETE`
* **Auth Required:** Yes
* **Role:** `super_admin` atau `admin_panitia` (pemilik event)

**Success Response (200 OK):**
```json
{
  "message": "event deleted successfully"
}
```
