# API Students — Dokumentasi Kontrak API

**Base URL**: `http://localhost:3000/api/v1`

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

**Contoh request**
```
GET /api/v1/students?page=1&limit=2&sort=grade&order=desc&is_active=true&min_grade=70
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


**Contoh request**
```
GET /api/v1/students/:id
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
