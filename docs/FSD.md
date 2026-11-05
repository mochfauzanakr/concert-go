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
- **Akses**:
  - `GET /api/organizers/profile` & `PUT /api/organizers/profile`: `ORGANIZER` (setelah aktivasi akun via invite).
  - `GET /api/organizers/:id`: Publik (dapat diakses oleh `BUYER` / pengunjung untuk melihat profil promotor penyelenggara event).
- **Endpoint**:
  - `GET /api/organizers/profile`: Mendapatkan data profil organisasi dari organizer yang sedang login (termasuk email dari akun `users`).
  - `PUT /api/organizers/profile`: Membuat atau memperbarui profil organisasi.
  - `GET /api/organizers/:id`: Mendapatkan informasi publik profil organizer berdasarkan ID profil / user organizer.
- **Content-Type (`PUT /api/organizers/profile`)**: `multipart/form-data` atau `application/json`.
- **Input (`PUT /api/organizers/profile`)**:
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
- **Prinsip**: Terisolasi per-organizer. Organizer hanya dapat mengundang, melihat, dan mengelola status staf yang mereka rekrut sendiri melalui tautan undangan.
- **Mekanisme Pelacakan**: Staf terasosiasi ke organizer melalui riwayat undangan di tabel `invites` (`invites.invited_by = current_organizer_id` dan `status = 'ACCEPTED'`).
- **Endpoint**:
  - `POST /api/users/invite`: Mengundang staf baru dengan role `STAFF` via email (mengikuti F-2.2).
  - `GET /api/organizers/staff`: Daftar seluruh staf yang berada di bawah naungan organizer yang sedang login (dengan query parameters: `page`, `pageSize`, `search`, `status`).
  - `PATCH /api/organizers/staff/:id/status`: Mengubah status akun staf (`ACTIVE` / `SUSPENDED`) milik organizer yang sedang login.
- **Input PATCH (`/api/organizers/staff/:id/status`)**:
  ```json
  {
    "status": "SUSPENDED" // Nilai valid: "ACTIVE", "SUSPENDED"
  }
  ```
- **Aturan Bisnis Status Staf**:
  - Organizer hanya berhak mengubah status user dengan role `STAFF` yang memiliki catatan undangan `invites.invited_by = organizer_user_id`. Jika staf tersebut milik organizer lain, tolak dengan `403 Forbidden` / `404 Not Found`.
  - Ketika staf di-suspend (`SUSPENDED`), seluruh sesi aktif staf tersebut di tabel `sessions` otomatis dicabut (*revoked*).
- **Catatan Modul Event (FD-4.0)**:
  - Penugasan staf dan co-organizer ke event tertentu (`event_organizers`) didefinisikan pada modul Event Management (FD-4.0).

---

### FD-4.0: Event, Lineup, & Team Management

*(Catatan: Pengelolaan jenis tiket, gelombang harga, dan alokasi kuota `ONLINE`/`OFFLINE` dipisahkan ke modul **FD-5.0: Ticket, Pricing & Inventory Management**).*

#### F-4.1 Public Event Catalog & Detail
- **Akses**: Publik (Semua pengunjung / `BUYER`).
- **Endpoint**:
  - `GET /api/events`: Daftar seluruh event dengan status `PUBLISHED`.
    - **Query Parameters**: `page`, `pageSize`, `search` (nama event / artis), `city`, `category`, `startDate`, `endDate`, `sortBy`.
  - `GET /api/events/:id`: Detail informasi lengkap suatu event (mencakup deskripsi, waktu, lokasi, kota, kategori, banner, info Lead Organizer, daftar Co-Organizer, daftar Lineup Artis, serta ringkasan ketersediaan tiket).

#### F-4.2 Event Creation, Update, & Status Lifecycle (Organizer)
- **Akses**: `ORGANIZER`.
- **Prasyarat Wajib**: Akun `ORGANIZER` wajib telah melengkapi `organizer_profiles` (F-3.1). Jika belum memiliki profil organisasi, pembuatan event ditolak dengan `422 Unprocessable Entity`.
- **Endpoint**:
  - `GET /api/organizers/events`: Daftar seluruh event yang dikelola oleh organizer yang sedang login (baik sebagai Lead Organizer maupun Co-Organizer, mencakup semua status: `DRAFT`, `PUBLISHED`, `COMPLETED`, `CANCELLED`).
  - `POST /api/events`: Membuat event baru (status awal otomatis `DRAFT`).
  - `PUT /api/events/:id`: Memperbarui data dan banner event (Hanya **Lead Organizer** pemilik `events.organizer_id`).
  - `PATCH /api/events/:id/status`: Mengubah status siklus hidup event (`DRAFT`, `PUBLISHED`, `COMPLETED`, `CANCELLED`) tanpa perlu persetujuan Admin.
- **Content-Type (`POST` & `PUT /api/events`)**: `multipart/form-data` atau `application/json`.
- **Input Event**:
  - `name`: string (wajib, min 3, max 150).
  - `description`: string (wajib).
  - `category`: string (wajib, contoh: `Festival`, `Rock`, `Pop`, `Jazz`, `Indie`).
  - `city`: string (wajib, contoh: `Jakarta`, `Bandung`, `Surabaya`).
  - `location`: string (wajib, nama venue / alamat lengkap).
  - `startTime`: timestamp ISO-8601 (wajib).
  - `endTime`: timestamp ISO-8601 (wajib, harus setelah `startTime`).
  - `inventoryMode`: string (wajib: `ONLINE_ONLY` atau `DUAL_CHANNEL` untuk penjualan gabungan online & offline/on-the-spot).
  - `banner`: file gambar (opsional, `.jpg`, `.jpeg`, `.png`, `.webp`, maks 2 MB, disimpan via `StorageService` ke folder `banners` dan URL disimpan di kolom `banner_url`).

#### F-4.3 Artist & Lineup Management
- **Akses**:
  - `GET /api/artists`: Publik / Semua Role.
  - Mutasi (`POST` & `DELETE`): `ORGANIZER` (Lead Organizer pada event terkait) dan `ADMIN`/`SUPER_ADMIN`.
- **Endpoint**:
  - `GET /api/artists`: Daftar master data artis (dengan pagination & `search`).
  - `POST /api/artists`: Menambahkan profil artis baru ke master data (`name`, `bio`, serta opsional foto `avatar` via `StorageService`).
  - `POST /api/events/:eventId/artists`: Menambahkan artis ke dalam lineup suatu event (`artistId`).
  - `DELETE /api/events/:eventId/artists/:artistId`: Menghapus artis dari lineup event.

#### F-4.4 Event Team Management (Co-Organizer & Gate Staff)
- **Akses**: `ORGANIZER` (Hanya **Lead Organizer** dari event terkait).
- **Prinsip**:
  - Penugasan bersifat langsung aktif (`assigned_at`) tanpa memerlukan konfirmasi penerimaan ulang.
  - `CO_ORGANIZER` dicantumkan sebagai mitra penyelenggara pada event dan dapat melihat event di dashboard mereka, sedangkan kendali penuh pengaturan event dan tiket tetap dipegang oleh Lead Organizer (`events.organizer_id`).
  - `STAFF` ditugaskan khusus ke event terkait untuk wewenang pemindaian & validasi QR tiket di gerbang masuk.
- **Endpoint**:
  - `GET /api/events/:eventId/team`: Daftar seluruh anggota tim (`CO_ORGANIZER` dan `STAFF`) yang ditugaskan pada event tersebut.
  - `POST /api/events/:eventId/team`: Menugaskan pengguna ke dalam tim event.
    - **Input**:
      ```json
      {
        "userId": "uuid-user",
        "roleInEvent": "STAFF" // Nilai valid: "CO_ORGANIZER", "STAFF"
      }
      ```
    - **Aturan Bisnis**:
      - Jika `roleInEvent == "STAFF"`: Pengguna target wajib memiliki role `STAFF`, berstatus `ACTIVE`, dan merupakan staf yang direkrut oleh Lead Organizer tersebut (`invites.invited_by = lead_organizer_id`).
      - Jika `roleInEvent == "CO_ORGANIZER"`: Pengguna target wajib memiliki role `ORGANIZER`, berstatus `ACTIVE`, dan bukan merupakan Lead Organizer itu sendiri.
      - Tolak dengan `409 Conflict` jika pengguna sudah terdaftar di tim event tersebut.
  - `DELETE /api/events/:eventId/team/:userId`: Mencopot (unassign) `CO_ORGANIZER` atau `STAFF` dari event tersebut.
