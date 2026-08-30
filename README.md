# API Students — Dokumentasi Kontrak API

**Base URL**: `http://localhost:3000/api/v1`


## Skema Tabel

Aplikasi ini memakai satu tabel: `students`.

```sql
CREATE TABLE IF NOT EXISTS students (
    id          SERIAL          PRIMARY KEY,
    nim         VARCHAR(20)     NOT NULL,
    name        VARCHAR(50)     NOT NULL,
    grade       NUMERIC(5,2)    NOT NULL,
    is_active   BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Keunikan NIM dijaga di level basis data, bukan lewat pengecekan manual di kode.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

-- Mempercepat pencarian nama lewat parameter ?search=, tidak membedakan huruf besar/kecil.
-- Bukan UNIQUE, karena dua mahasiswa berbeda boleh punya nama yang sama.
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
```

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `SERIAL PRIMARY KEY` | Dibuat otomatis oleh database, tidak pernah dikirim klien |
| `nim` | `VARCHAR(20)` | Kode identitas mahasiswa, wajib unik |
| `name` | `VARCHAR(50)` | Nama mahasiswa |
| `grade` | `NUMERIC(5,2)` | Nilai, rentang 0.00–100.00 |
| `is_active` | `BOOLEAN` | Status keaktifan, default `TRUE` |
| `created_at` | `TIMESTAMPTZ` | Waktu data dibuat, diisi otomatis oleh database |

---

## Cara Menyiapkan Basis Data dari Nol

Panduan ini mengasumsikan PostgreSQL sudah terpasang di komputer, tapi database untuk proyek ini belum ada.

**1. Buat database kosong**
Buka terminal, lalu jalankan (sesuaikan `postgres` dengan nama user PostgreSQL anda bila berbeda):

```bash
psql -U postgres -c "CREATE DATABASE praktikum_backend;"
```

Jika perintah `psql` tidak ditemukan, berarti PostgreSQL belum terpasang atau belum ditambahkan ke PATH sistem(instal dulu PostgreSQL sesuai OS kamu sebelum lanjut).

**2. Jalankan berkas migrasi untuk membuat tabel**

```bash
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
```

**3. Pastikan tabel dan indeksnya benar-benar terbuat**

```bash
psql -U postgres -d praktikum_backend -c "\d students"
```
Nanti akan muncul daftar kolom `id, nim, name, grade, is_active, created_at`, beserta dua indeks (`students_nim_key` dan `students_name_lower_idx`) di bagian bawah output.

**4. Siapkan berkas `.env`**

Salin `.env.example` menjadi `.env`, lalu isi nilainya sesuai PostgreSQL di komputermu:

```bash
cp .env.example .env
```

Buka `.env` dan isi minimal `DB_PASSWORD` dengan kata sandi PostgreSQL. Lihat bagian **Variabel Environment** di bawah untuk penjelasan tiap nilainya.

**5. Pasang dependensi Go dan jalankan aplikasi**

```bash
go mod tidy
go run .
```

Jika semua langkah di atas benar, akan muncul:
```
Server berjalan di http://localhost:3000
```

**6. Verifikasi database sudah tersambung**

```bash
curl -i http://localhost:3000/api/v1/health
```

Respons `200 OK` dengan pesan `"server dan database berjalan"` berarti setup sudah benar. Kalau muncul `503`, cek kembali apakah PostgreSQL sedang menyala dan nilai di `.env` sudah sesuai.

---

## Variabel Environment

Seluruh variabel berikut didefinisikan di `.env.example` (tanpa nilai) dan wajib diisi di `.env` milikmu sendiri (berkas ini tidak ikut ter-commit ke Git, sesuai `.gitignore`).

| Variabel | Wajib diisi | Nilai bawaan bila kosong | Keterangan |
|---|---|---|---|
| `APP_PORT` | Tidak | `3000` | Port tempat server Fiber berjalan |
| `DB_HOST` | Tidak | `localhost` | Alamat server PostgreSQL |
| `DB_PORT` | Tidak | `5432` | Port PostgreSQL |
| `DB_USER` | Tidak | `postgres` | Username untuk login ke PostgreSQL |
| `DB_PASSWORD` | **Ya** | kosong | Kata sandi PostgreSQL — wajib diisi sesuai instalasi masing-masing |
| `DB_NAME` | Tidak | `backend` | Nama database. |
| `DB_SSLMODE` | Tidak | `disable` | Mode SSL koneksi. `disable` untuk lokal|
| `DB_MAX_CONNS` | Tidak | `10` | Jumlah maksimum koneksi dalam connection pool |

Contoh isi `.env` yang lengkap untuk pengembangan lokal:

```env
APP_PORT=3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=kata_sandi_anda
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

---

## Entitas Student

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | int | Penanda internal, dibuat otomatis oleh server, tidak bisa diubah klien |
| `nim` | string | Penanda unik, wajib saat membuat data, tidak bisa diubah lewat PUT/PATCH |
| `name` | string | Nama mahasiswa |
| `grade` | float | Nilai, rentang 0–100 |
| `is_active` | bool | Status keaktifan mahasiswa |

---

## Daftar Endpoint

### 1. GET `/students` — Daftar mahasiswa

Mendukung paginasi, pencarian, pengurutan, dan penyaringan lewat query string.

| Parameter | Tipe | Bawaan | Keterangan |
|---|---|---|---|
| `page` | int | `1` | Halaman keberapa |
| `limit` | int | `10` | Baris per halaman, batas atas `100` |
| `search` | string | kosong | Cari pada `name`, tidak case-sensitive |
| `sort` | string | `id` | Kolom pengurutan, whitelist: `id`, `name`, `grade`, `nim` |
| `order` | string | `asc` | `asc` atau `desc` |
| `is_active` | bool | tidak menyaring | Saring berdasarkan status aktif |
| `min_grade` | float | tidak menyaring | Batas bawah nilai |
| `max_grade` | float | tidak menyaring | Batas atas nilai |

**Contoh request dengan terminal Powershell** 
```
curl.exe -i -X GET "localhost:3000/api/v1/students?page=1&limit=2&sort=grade&order=desc&is_active=true&min_grade=70"
```

**Contoh respons — 200 OK**
```json
{
  "success": true,
  "message": "daftar mahasiswa berhasil diambil",
  "data": [
    { "id": 1, "nim": "2023010001", "name": "Sari Melati", "grade": 85, "is_active": true }
  ],
  "meta": { "page": 1, "limit": 2, "total": 1, "total_pages": 1 }
}
```

---

### 2. GET `/students/:id` — Ambil satu mahasiswa

| Path Parameter | Tipe | Keterangan |
|---|---|---|
| `id` | int | ID internal mahasiswa |


**Contoh request dengan terminal Powershell**
```
curl.exe -i -X GET "localhost:3000/api/v1/students/1"
```


**Contoh respons — 200 OK**
```json
{
  "success": true,
  "message": "mahasiswa ditemukan",
  "data": { "id": 1, "nim": "2023010001", "name": "Sari Melati", "grade": 85, "is_active": true }
}
```

**Status yang mungkin**

| Status | Situasi |
|---|---|
| 200 | Data ditemukan |
| 400 | `id` bukan angka positif |
| 404 | `id` tidak ada di data |

**Contoh respons — 404 Not Found**
```json
{ "success": false, "message": "mahasiswa tidak ditemukan" }
```

---

### 3. POST `/students` — Tambah mahasiswa baru

Seluruh field wajib dikirim. `is_active` otomatis diset `true` oleh server.

**Contoh body request**
```json
{ "nim": "2023010001", "name": "Sari Melati", "grade": 85 }
```

**Contoh respons — 201 Created**

Header tambahan: `Location: /api/v1/students/1`

```json
{
  "success": true,
  "message": "mahasiswa berhasil dibuat",
  "data": { "id": 1, "nim": "2023010001", "name": "Sari Melati", "grade": 85, "is_active": true }
}
```

**Status yang mungkin**

| Status | Situasi |
|---|---|
| 201 | Berhasil dibuat |
| 400 | Body bukan JSON yang valid |
| 409 | `nim` sudah terdaftar |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Field kosong atau `grade` di luar rentang 0–100 |

**Contoh respons — 409 Conflict**
```json
{ "success": false, "message": "NIM sudah terdaftar" }
```

**Contoh respons — 422 Unprocessable Entity**
```json
{ "success": false, "message": "validasi gagal", "errors": { "grade": "harus di antara 0 dan 100" } }
```

---

### 4. PUT `/students/:id` — Ganti seluruh data

Semua field wajib dikirim. Field yang tidak dikirim dianggap dikosongkan/direset — bukan dibiarkan seperti semula. `nim` tidak bisa diubah lewat endpoint ini.

**Contoh body permintaan**
```json
{ "name": "Sari Melati Putri", "grade": 95, "is_active": false }
```

**Contoh respons — 200 OK**
```json
{
  "success": true,
  "message": "mahasiswa berhasil diganti seluruhnya",
  "data": { "id": 1, "nim": "2023010001", "name": "Sari Melati Putri", "grade": 95, "is_active": false }
}
```

**Status yang mungkin**

| Status | Situasi |
|---|---|
| 200 | Berhasil diganti |
| 400 | `id` bukan angka, atau body bukan JSON valid |
| 404 | Data tidak ditemukan |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Field wajib kosong, atau `grade` di luar rentang |

---

### 5. PATCH `/students/:id` — Ubah sebagian data

Hanya field yang dikirim yang berubah. Field lain dibiarkan seperti semula. `nim` tidak bisa diubah lewat endpoint ini.

**Contoh body permintaan**
```json
{ "is_active": true }
```

**Contoh respons — 200 OK**
```json
{
  "success": true,
  "message": "mahasiswa berhasil diperbarui sebagian",
  "data": { "id": 1, "nim": "2023010001", "name": "Sari Melati Putri", "grade": 95, "is_active": true }
}
```

**Status yang mungkin**

| Status | Situasi |
|---|---|
| 200 | Berhasil diperbarui |
| 400 | `id` bukan angka, body bukan JSON valid, atau tidak ada field dikirim sama sekali |
| 404 | Data tidak ditemukan |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Field yang dikirim tidak lolos validasi (misal `grade` di luar rentang) |

---

### 6. DELETE `/students/:id` — Hapus mahasiswa

**Contoh respons — 204 No Content**

Tanpa body.

**Status yang mungkin**

| Status | Situasi |
|---|---|
| 204 | Berhasil dihapus |
| 400 | `id` bukan angka positif |
| 404 | Data tidak ditemukan (termasuk saat mengulang DELETE ke id yang sama) |

---

## Ringkasan Status HTTP yang Dipakai

| Status | Dipakai pada |
|---|---|
| 200 OK | GET (satu/daftar), PUT, PATCH berhasil |
| 201 Created | POST berhasil, disertai header `Location` |
| 204 No Content | DELETE berhasil |
| 400 Bad Request | `id` bukan angka, body bukan JSON valid, PATCH tanpa field |
| 404 Not Found | Data dengan `id` yang diminta tidak ada |
| 409 Conflict | `nim` yang dikirim sudah dipakai mahasiswa lain |
| 415 Unsupported Media Type | `Content-Type` bukan `application/json` pada POST/PUT/PATCH |
| 422 Unprocessable Entity | Bentuk data dipahami tetapi isinya tidak lolos validasi |

