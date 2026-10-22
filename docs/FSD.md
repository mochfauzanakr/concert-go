# Functional Specification Document (FSD)

## 1. Standar Fungsional

### 1.1 Format Response API
Response HTTP wajib menggunakan struktur JSON berikut (`GlobalResponse`):
```json
{
  "message": "String deskripsi respon",
  "data": { ... },
  "pagination": {
    "currentPage": 1,
    "nextPage": 2,
    "prevPage": null,
    "totalPage": 5,
    "totalRecords": 50
  },
  "reqId": "uuid-v4-string",
  "status": "T" // "T" untuk True (sukses), "F" untuk False (gagal)
}
```

### 1.2 Konvensi Penamaan (Naming Convention)
- **JSON Payload & Response**: `camelCase` (contoh: `newPassword`, `refreshToken`, `inviteToken`)
- **Query Parameters**: `camelCase` (contoh: `pageSize`, `searchQuery`)
- **Database Column & Table**: `snake_case` (contoh: `password_hash`, `role_id`, `invites`)

### 1.3 Standar HTTP Status
- `200 OK`: Berhasil (Read, Update, Delete)
- `201 Created`: Berhasil (Insert/Create)
- `400 Bad Request`: Validasi input gagal / format data salah
- `401 Unauthorized`: Token tidak ada, expired, atau tidak valid
- `403 Forbidden`: Role tidak berhak mengakses resource
- `404 Not Found`: Data tidak ditemukan
- `422 Unprocessable Entity`: Melanggar aturan bisnis (contoh: kuota habis, email sudah terdaftar)
- `500 Internal Server Error`: Kegagalan server / database

---

## 2. Struktur Role & Hak Akses

| Role | Onboarding Method | Kewenangan Utama |
| :--- | :--- | :--- |
| **SUPER_ADMIN** | Database Seeding (Tanpa Registrasi/Invite) | Akses mutlak sistem. Mengundang `ADMIN` dan `ORGANIZER`. Tidak dapat membuat Super Admin lain. |
| **ADMIN** | Diundang oleh `SUPER_ADMIN` | Mengelola data user, memantau sistem, mengundang `ORGANIZER`. |
| **ORGANIZER** | Diundang oleh `SUPER_ADMIN` atau `ADMIN` | Membuat & mengelola event, tiket, kuota, harga, serta mengundang `STAFF`. |
| **STAFF** | Diundang oleh `ORGANIZER` | Hak akses terbatas pada event terkait (scan & validasi QR tiket di gate). |
| **BUYER** | Registrasi Mandiri (Email, OTP, Google OAuth) | Pengguna umum: jelajah event, beli tiket, riwayat transaksi, akses QR tiket. |

---

## 3. Modul Fungsional

### FD-1.0: Autentikasi & Akun Pribadi (Authentication & Account)

#### F-1.1 Register & Login (Buyer Mandiri)
- **Akses**: Publik (`BUYER`).
- **Input Register**: `name`, `email`, `password`.
- **Input Login**: `email`, `password`.
- **Validasi**: Email valid, password minimal 8 karakter.

#### F-1.2 OTP Login & Register (Passwordless)
- **Akses**: Publik (`BUYER`).
- **Input**: `email`, `otp` (6 digit numerik).
- **Aturan**: OTP dikirim via email, masa aktif (TTL) 5 menit.

#### F-1.3 OAuth2 Login (Google)
- **Akses**: Publik (`BUYER`).
- **Alur**: Redirect ke Google OAuth -> Callback menerima profil -> Buat/login user.

#### F-1.4 Change Password (Ubah Kata Sandi)
- **Akses**: Pengguna Login (Semua Role kecuali OAuth tanpa password).
- **Lokasi UI**: Halaman Profil / Pengaturan Akun.
- **Input**: `Current Password`, `New Password`, `Confirm New Password`.
- **UI & Validasi**:
  - Terdapat ikon mata silang (*eye-slash*) pada tiap field untuk sembunyikan/tampilkan teks.
  - Terdapat indikator kekuatan kata sandi (*Password Strength*: Weak / Fair / Good / Strong).
  - Kata sandi baru minimal 8 karakter dan tidak boleh sama dengan kata sandi lama.

#### F-1.5 Me (Data Pengguna Login)
- **Akses**: Pengguna Login (Bearer Token).
- **Output**: ID, nama, email, role, avatarUrl, permissions, dan metadata sesi.

#### F-1.6 Edit Profile (Self)
- **Akses**: Pengguna Login (Semua Role).
- **Endpoint**: `PUT /api/users/me`.
- **Content-Type**: `multipart/form-data` atau `application/json`.
- **Input**:
  - `name`: string (opsional, min 2, max 100).
  - `avatar`: file gambar (opsional, tipe: `.jpg`, `.jpeg`, `.png`, `.webp`, ukuran maksimal 2 MB).
- **Penyimpanan Gambar**: Auto-fallback (unggah ke Cloudinary jika API key tersedia di konfigurasi, fallback ke local storage server jika belum diset).

---

### FD-2.0: User Management & Onboarding Undangan (Invite-Only)

#### F-2.1 User Management List
- **Akses**: `SUPER_ADMIN` dan `ADMIN`.
- **Komponen UI**:
  - Tombol **"Invite User"** (membuka modal undangan).
  - Search bar (`"Search users..."`).
  - Dropdown Filter (berdasarkan Role dan Status Akun).
- **Tabel Pengguna**:
  - `FULL NAME` (avatar inisial nama).
  - `EMAIL`.
  - `ROLE` (Badge peran: Admin, Organizer, Staff, Buyer).
  - `STATUS` (Badge dot: Active, Suspended).
  - `ACTIONS` (Menu: Detail, Suspend/Reactivate).

#### F-2.2 Invite User Action
- **Akses**:
  - `SUPER_ADMIN` dapat mengundang `ADMIN` dan `ORGANIZER`.
  - `ADMIN` dapat mengundang `ORGANIZER`.
  - `ORGANIZER` dapat mengundang `STAFF`.
- **Input**: `email`, `role`.
- **Aturan Bisnis**:
  - **Tolak** jika email sudah terdaftar di sistem (sebagai role apapun).
  - Sistem membuat token undangan dengan TTL 24 jam.
  - Kirim email undangan berisi tautan penerimaan (`/accept-invite?token=...`).

#### F-2.3 Accept Invite & Account Activation
- **Akses**: Pengguna yang menerima tautan undangan via email.
- **Input**: `name`, `password`, `confirmPassword`.
- **Aturan Bisnis**:
  - Validasi token invite (apakah valid dan belum expired).
  - Jika valid, akun diaktifkan dengan role yang telah ditentukan saat diundang.
  - Token invite ditandai sebagai `ACCEPTED` (tidak bisa digunakan ulang).

#### F-2.4 User Detail & Status Management (Suspend / Reactivate)
- **Akses**: `SUPER_ADMIN` dan `ADMIN`.
- **Endpoint**:
  - `GET /api/users/:id` (Detail Pengguna)
  - `PATCH /api/users/:id/status` (Ubah Status Pengguna)
- **Input PATCH**:
  ```json
  {
    "status": "SUSPENDED" // Nilai valid: "ACTIVE", "SUSPENDED"
  }
  ```
- **Aturan Bisnis**:
  - Tidak dapat mengubah status akun diri sendiri (*self-lockout prevention*).
  - Role `SUPER_ADMIN` tidak dapat di-suspend oleh siapapun.
  - Role `ADMIN` tidak dapat men-suspend `SUPER_ADMIN` maupun sesama `ADMIN`.
  - Ketika akun di-suspend, seluruh sesi aktif dibatalkan dan pengguna langsung ditolak saat mencoba login atau refresh token.

#### F-2.5 Admin Edit User
- **Akses**: `SUPER_ADMIN` dan `ADMIN`.
- **Endpoint**: `PUT /api/users/:id`.
- **Input**:
  ```json
  {
    "name": "Updated Name",
    "roleId": 3
  }
  ```
- **Aturan Bisnis**:
  - `ADMIN` tidak dapat mengubah data maupun role sesama `ADMIN` atau `SUPER_ADMIN`.
  - Pengguna tidak dapat mengubah rolenya sendiri menjadi lebih rendah (*anti-self-demotion*).
  - Role baru yang dipilih wajib terdaftar di database.

---

### FD-3.0: Organizer Management

#### F-3.1 Organizer Profile Setup
- **Akses**: `ORGANIZER` (setelah aktivasi akun via invite).
- **Input**: `orgName`, `description`, `logoUrl`, kontak organisasi.
- **Status**: Siap membuat event setelah profil terisi.

#### F-3.2 Event Staff Management
- **Akses**: `ORGANIZER`.
- **Prinsip**: Terisolasi per-organizer (organizer hanya melihat dan mengelola staf mereka sendiri, terpisah dari `GET /api/users` platform).
- **Endpoint**:
  - `GET /api/organizers/staff`: Daftar seluruh staf yang berada di bawah naungan organizer yang sedang login (dengan pagination & search).
  - `GET /api/events/:eventId/staff`: Daftar staf yang ditugaskan khusus pada suatu event untuk gerbang scan tiket.
  - `POST /api/events/:eventId/staff/assign`: Menugaskan staf ke gerbang event tertentu.
