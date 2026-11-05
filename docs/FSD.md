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

#### F-3.1 Organizer Profile Setup & Management
- **Akses**: `ORGANIZER` (setelah aktivasi akun via invite).
- **Endpoint**:
  - `GET /api/organizers/profile`: Mendapatkan data profil organisasi dari organizer yang sedang login (termasuk email dari akun `users`).
  - `PUT /api/organizers/profile`: Membuat atau memperbarui profil organisasi.
- **Content-Type**: `multipart/form-data` atau `application/json`.
- **Input**:
  - `orgName`: string (wajib, min 3, max 100).
  - `description`: string (opsional).
  - `phone`: string (opsional, min 8, max 20).
  - `socialMedia`: string (opsional, URL website / akun media sosial).
  - `logo`: file gambar (opsional, tipe: `.jpg`, `.jpeg`, `.png`, `.webp`, ukuran maksimal 2 MB).
- **Penyimpanan Gambar**: Diunggah via `StorageService` (Cloudinary jika terkonfigurasi, fallback ke local storage), URL disimpan pada kolom `logo_url`.
- **Aturan Bisnis**:
  - Email kontak resmi organisasi otomatis merujuk ke `users.email` akun organizer yang bersangkutan (tidak ada duplikasi kolom email di tabel `organizer_profiles`).
  - Tidak memerlukan persetujuan lanjutan admin (*no verification gate* / `verification_status` ditiadakan) karena kredensial organizer sudah terverifikasi saat diundang oleh `SUPER_ADMIN` atau `ADMIN`.
  - Profil organisasi wajib dilengkapi sebagai prasyarat utama sebelum diizinkan membuat event.

#### F-3.2 Staff Management (Organizer-scoped)
- **Akses**: `ORGANIZER`.
- **Prinsip**: Terisolasi per-organizer. Organizer hanya dapat mengundang, melihat, dan mengelola staf yang mereka rekrut sendiri melalui tautan undangan.
- **Mekanisme Pelacakan**: Staf terasosiasi ke organizer melalui riwayat undangan di tabel `invites` (`invites.invited_by = current_organizer_id` dan `status = 'ACCEPTED'`).
- **Endpoint**:
  - `POST /api/users/invite`: Mengundang staf baru dengan role `STAFF` via email (mengikuti F-2.2).
  - `GET /api/organizers/staff`: Daftar seluruh staf yang berada di bawah naungan organizer yang sedang login (dengan query parameters: `page`, `pageSize`, `search`).
- **Catatan Modul Event (FD-4.0)**:
  - Penugasan staf ke gerbang event tertentu (`GET /api/events/:eventId/staff` dan `POST /api/events/:eventId/staff/assign`) akan didefinisikan dan diimplementasikan secara terintegrasi pada modul Event Management.
