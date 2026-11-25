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
- `409 Conflict`: Konflik state / duplikasi data (contoh: email sudah terdaftar, pengajuan sudah direview)
- `422 Unprocessable Entity`: Melanggar aturan bisnis (contoh: kuota habis, profil organisasi belum dilengkapi)
- `500 Internal Server Error`: Kegagalan server / database

### 1.4 Standar Global Query Parameters (`BaseListParams`)
Seluruh endpoint `GET` berbasis daftar/tabel wajib mendukung parameter kueri standar berikut (diekstrak melalui `util.ParseBaseListParams` ke `model.BaseListParams`):
- `page`: Nomor halaman (default: `1`).
- `pageSize` (atau `limit`): Jumlah item per halaman (default: `10`, maksimal: `20`).
- `search` (atau `searchQuery`): Kata kunci pencarian teks bebas sesuai konteks tabel.
- `sortBy`: Kolom pengurutan (di-whitelist per tabel untuk mencegah SQL Injection, default: `created_at`).
- `sortOrder`: Arah pengurutan `ASC` atau `DESC` (default: `DESC`).
- `startDate`: Filter batas bawah waktu pembuatan (`created_at >= startDate`, format ISO-8601 / `YYYY-MM-DD`).
- `endDate`: Filter batas atas waktu pembuatan (`created_at <= endDate`, format ISO-8601 / `YYYY-MM-DD`).

---

## 2. Struktur Role & Hak Akses

| Role | Onboarding Method | Kewenangan Utama |
| :--- | :--- | :--- |
| **SUPER_ADMIN** | Database Seeding (Tanpa Registrasi/Invite) | Akses mutlak sistem. Mengundang `ADMIN` dan `ORGANIZER`. Tidak dapat membuat Super Admin lain. |
| **ADMIN** | Diundang oleh `SUPER_ADMIN` | Mengelola data user, memantau sistem, mengundang `ORGANIZER`. |
| **ORGANIZER** | Diundang oleh `SUPER_ADMIN` / `ADMIN` (Direct Invite atau via Pengajuan Kemitraan Publik) | Membuat & mengelola event, tiket, kuota, harga, serta mengundang `STAFF`. |
| **STAFF** | Diundang oleh `ORGANIZER` | Hak akses terbatas pada event terkait (scan & validasi QR tiket di gate). |
| **USER** | Registrasi Mandiri (Email, OTP, Google OAuth) | Pengguna umum: jelajah event, beli tiket, riwayat transaksi, akses QR tiket. |

---

## 3. Modul Fungsional

### FD-1.0: Autentikasi & Akun Pribadi (Authentication & Account)

#### F-1.1 Register & Login (Buyer Mandiri)
- **Akses**: Publik (`USER`).
- **Input Register**: `name`, `email`, `password`.
- **Input Login**: `email`, `password`.
- **Validasi**: Email valid, password minimal 8 karakter.

#### F-1.2 OTP Login & Register (Passwordless)
- **Akses**: Publik (`USER`).
- **Input**: `email`, `otp` (6 digit numerik).
- **Aturan**: OTP dikirim via email, masa aktif (TTL) 5 menit.

#### F-1.3 OAuth2 Login (Google)
- **Akses**: Publik (`USER`).
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
  - Khusus untuk undangan dengan role `STAFF`, sistem otomatis mencatat relasi antara organizer pengundang (`invites.invited_by`) dan staf baru (`users.id`) ke tabel `organizer_staffs`.
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
  - `GET /api/organizers/:id`: Publik (dapat diakses oleh `USER` / pengunjung untuk melihat profil promotor penyelenggara event).
- **Endpoint**:
  - `GET /api/organizers/profile`: Mendapatkan data profil organisasi dari organizer yang sedang login (termasuk email dari akun `users`).
  - `PUT /api/organizers/profile`: Membuat atau memperbarui profil organisasi.
  - `GET /api/organizers/:id`: Mendapatkan informasi publik profil organizer berdasarkan ID profil / user organizer. Respons publik **tidak menyertakan email login** organizer (`email` hanya tersedia pada `GET /api/organizers/profile` milik organizer yang bersangkutan).
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
- **Mekanisme Pelacakan**: Staf terasosiasi ke organizer melalui tabel `organizer_staffs` (`organizer_staffs.organizer_id = current_organizer_id` dan `organizer_staffs.user_id = staff_user_id`) yang otomatis tercatat saat staf menyelesaikan `AcceptInvite`.
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
  - Organizer hanya berhak mengubah status user dengan role `STAFF` yang terdaftar di bawah organisasinya pada tabel `organizer_staffs` (`organizer_staffs.organizer_id = organizer_user_id`). Jika staf tersebut milik organizer lain, tolak dengan `403 Forbidden` / `404 Not Found`.
  - Ketika staf di-suspend (`SUSPENDED`), seluruh sesi aktif staf tersebut di tabel `sessions` otomatis dicabut (*revoked*).
- **Catatan Modul Event (FD-4.0)**:
  - Penugasan staf dan co-organizer ke event tertentu (`event_organizers`) didefinisikan pada modul Event Management (FD-4.0).

#### F-3.3 Public Organizer Partnership Application (Inbound Lead / Become an Organizer)
- **Akses**:
  - `POST /api/organizers/apply`: Publik (Tanpa Login, dilindungi Rate Limiter ketat per IP/Email).
  - `GET /api/organizers/applications` & `PATCH /api/organizers/applications/:id/review`: `SUPER_ADMIN` dan `ADMIN`.
- **Tujuan Bisnis**: Menyediakan jalur pendaftaran terbuka bagi promotor/EO eksternal untuk mengajukan kemitraan di platform tanpa harus menunggu diundang lebih dulu secara manual.
- **Input (`POST /api/organizers/apply`)**:
  - `picName`: string (wajib, min 2, max 100 — nama penanggung jawab).
  - `email`: string (wajib, format email valid, belum terdaftar di `users`).
  - `phone`: string (wajib, min 8, max 20 — nomor telepon/WhatsApp aktif).
  - `orgName`: string (wajib, min 3, max 100 — nama organisasi/promotor).
  - `socialMedia`: string (opsional, tautan media sosial / portofolio website).
  - `description`: string (wajib, min 10 — deskripsi organisasi dan rencana event).
- **Aturan Bisnis & Alur Persetujuan**:
  - **Tolak (`409 Conflict`)** jika `email` sudah terdaftar sebagai user aktif atau masih memiliki pengajuan berstatus `PENDING`.
  - Pengajuan baru disimpan di tabel `organizer_applications` dengan status awal `PENDING`.
  - Admin meninjau daftar pengajuan melalui `GET /api/organizers/applications` (mendukung filter `status`, `search`, dan pagination).
  - Saat Admin mengeksekusi `PATCH /api/organizers/applications/:id/review` dengan `status: "APPROVED"`:
    1. Sistem otomatis memicu pengiriman email undangan (`invites`) dengan role `ORGANIZER` ke alamat `email` pengaju (TTL 24 jam).
    2. Saat pengaju mengklik tautan undangan dan menyelesaikan `POST /api/users/accept-invite`, sistem otomatis membuat akun `USER` ber-role `ORGANIZER` sekaligus menginisialisasi data `organizer_profiles` (`org_name`, `phone`, `social_media`, `description`) dari data pengajuan sehingga organizer tidak perlu mengetik ulang profilnya.
  - Jika Admin memilih `status: "REJECTED"` (opsional disertakan `rejectionReason`), status pengajuan berubah menjadi `REJECTED` dan tidak ada undangan yang dikirim.

---

### FD-4.0: Event, Lineup, & Team Management

*(Catatan: Pengelolaan jenis tiket, gelombang harga, dan alokasi kuota `ONLINE`/`OFFLINE` dipisahkan ke modul **FD-5.0: Ticket, Pricing & Inventory Management**).*

#### F-4.1 Public Event Catalog & Detail
- **Akses**: Publik (Semua pengunjung / `USER`).
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
      - Jika `roleInEvent == "STAFF"`: Pengguna target wajib memiliki role `STAFF`, berstatus `ACTIVE`, dan terdaftar sebagai staf milik Lead Organizer tersebut pada tabel `organizer_staffs` (`organizer_staffs.organizer_id = lead_organizer_id`).
      - Jika `roleInEvent == "CO_ORGANIZER"`: Pengguna target wajib memiliki role `ORGANIZER`, berstatus `ACTIVE`, dan bukan merupakan Lead Organizer itu sendiri.
      - Tolak dengan `409 Conflict` jika pengguna sudah terdaftar di tim event tersebut.
  - `DELETE /api/events/:eventId/team/:userId`: Mencopot (unassign) `CO_ORGANIZER` atau `STAFF` dari event tersebut.

---

### FD-5.0: Ticket, Pricing Phase & Inventory Management

#### F-5.1 Ticket Type & Master Inventory Setup
- **Akses**:
  - `GET /api/events/:eventId/ticket-types`: Publik (melihat ketersediaan & fase harga aktif) / `ORGANIZER` (melihat detail kuota penuh `ONLINE` & `OFFLINE`).
  - Mutasi (`POST`, `PUT`, `DELETE`): `ORGANIZER` (Hanya **Lead Organizer** pemilik `events.organizer_id`).
- **Prinsip Authority Kuota (`inventories`)**:
  - Tabel `inventories` (`id`, `ticket_type_id`, `channel`, `quota`, `sold`, `reserved`) adalah **master kuota fisik**.
  - Jika `events.inventory_mode == "ONLINE_ONLY"`, sistem membuat 1 baris inventory (`channel = "ONLINE"`).
  - Jika `events.inventory_mode == "DUAL_CHANNEL"`, sistem membuat 2 baris inventory (`channel = "ONLINE"` dan `channel = "OFFLINE"`).
- **Endpoint**:
  - `GET /api/events/:eventId/ticket-types`: Daftar seluruh kategori tiket pada suatu event beserta rincian `inventories`, `pricing_phases`, dan `seatingType`.
  - `POST /api/events/:eventId/ticket-types`: Membuat kategori tiket baru sekaligus menginisialisasi kuota fisik di `inventories`.
    - **Input**:
      ```json
      {
        "name": "Festival / GA",
        "description": "Free standing area dekat panggung utama",
        "seatingType": "AREA", // Nilai valid: "AREA" (default) atau "SEATED"
        "onlineQuota": 500,
        "offlineQuota": 100
      }
      ```
  - `PUT /api/events/:eventId/ticket-types/:ticketTypeId`: Memperbarui nama, deskripsi, dan kuota fisik `inventories`.
    - **Aturan Bisnis**:
      - Kuota baru (`onlineQuota` / `offlineQuota`) **dilarang lebih kecil** dari jumlah tiket yang sudah terjual dan sedang ditahan (`quota >= sold + reserved`). Jika dilanggar, tolak dengan `422 Unprocessable Entity`.
      - `seatingType` (`AREA` / `SEATED`) **tidak dapat diubah** jika sudah ada tiket yang terjual/ditahan atau sudah memiliki daftar kursi di tabel `seats`.
  - `DELETE /api/events/:eventId/ticket-types/:ticketTypeId`: Menghapus kategori tiket (Hanya diizinkan jika `sold == 0` dan `reserved == 0` di seluruh kanal).

#### F-5.2 Pricing Phase (Gelombang Harga) Management
- **Akses**: `ORGANIZER` (Hanya **Lead Organizer** pemilik `events.organizer_id`).
- **Prinsip**:
  - `pricing_phases` (`id`, `ticket_type_id`, `name`, `price`, `quota`, `sold`, `start_at`, `end_at`) mengatur harga aktif dan batas alokasi penjualan per gelombang (contoh: `Early Bird`, `Presale 1`, `Normal`), namun pemotongan stok fisik tetap merujuk ke tabel `inventories` (`channel = "ONLINE"`).
  - Kolom `price` wajib menggunakan tipe `NUMERIC(15, 2)`.
- **Endpoint**:
  - `POST /api/ticket-types/:ticketTypeId/pricing-phases`: Menambahkan gelombang harga baru.
    - **Input**:
      ```json
      {
        "name": "Early Bird",
        "price": 350000.00,
        "quota": 150,
        "startAt": "2026-12-01T10:00:00Z",
        "endAt": "2026-12-10T23:59:59Z"
      }
      ```
    - **Aturan Bisnis**:
      - `price >= 0` dan `endAt > startAt`.
      - `quota` gelombang harga tidak boleh melebihi `inventories.quota` kanal `ONLINE`.
  - `PUT /api/ticket-types/:ticketTypeId/pricing-phases/:phaseId`: Memperbarui gelombang harga (`quota` baru tidak boleh `< sold`).
  - `DELETE /api/ticket-types/:ticketTypeId/pricing-phases/:phaseId`: Menghapus gelombang harga (Hanya diizinkan jika `sold == 0`).

#### F-5.3 Offline / On-The-Spot Sales Recording
- **Akses**: `ORGANIZER` (Lead Organizer atau Co-Organizer pada event berstatus `DUAL_CHANNEL`).
- **Endpoint**:
  - `POST /api/ticket-types/:ticketTypeId/inventory/offline-sale`: Mencatat penjualan tiket fisik/on-the-spot pada kanal `OFFLINE`.
    - **Input**:
      ```json
      {
        "quantity": 5,
        "notes": "Penjualan booth OTS Gate A"
      }
      ```
    - **Aturan Bisnis**:
      - Hanya berlaku untuk event dengan `inventory_mode == "DUAL_CHANNEL"`.
      - Menggunakan `SELECT ... FOR UPDATE` pada baris `inventories` (`channel = "OFFLINE"`). Jika `sold + quantity > quota`, tolak dengan `422 Unprocessable Entity`.

#### F-5.4 Seat Map / Numbered Seat Setup (Kategori `SEATED`)
- **Akses**:
  - `GET /api/ticket-types/:ticketTypeId/seats`: Publik (melihat denah daftar kursi dan ketersediaannya).
  - `POST /api/ticket-types/:ticketTypeId/seats/bulk` & `DELETE /api/ticket-types/:ticketTypeId/seats/:seatId`: `ORGANIZER` (Hanya **Lead Organizer** pemilik `events.organizer_id`).
- **Endpoint**:
  - `GET /api/ticket-types/:ticketTypeId/seats`: Daftar seluruh kursi milik kategori tiket tersebut, diurutkan berdasarkan `row_label` dan `seat_number`, beserta status (`AVAILABLE`, `RESERVED`, `SOLD`).
  - `POST /api/ticket-types/:ticketTypeId/seats/bulk`: Generate daftar kursi bernomor secara massal per baris.
    - **Input**:
      ```json
      {
        "rows": [
          { "rowLabel": "A", "count": 20 },
          { "rowLabel": "B", "count": 20 }
        ]
      }
      ```
    - **Aturan Bisnis**:
      - Hanya diizinkan jika `ticket_types.seating_type == "SEATED"` (jika `AREA`, tolak dengan `422 Unprocessable Entity`).
      - Total jumlah kursi yang dibuat pada kategori tersebut **tidak boleh melebihi** `inventories.quota` kanal `ONLINE` (jika melebihi, tolak dengan `422 Unprocessable Entity`).
      - Kombinasi `(ticket_type_id, row_label, seat_number)` bersifat unik; jika terdapat duplikat baris/nomor kursi, tolak dengan `409 Conflict`.
      - Setiap kursi menyimpan `label` (contoh: `"A-1"`, `"A-2"`) dengan status awal `AVAILABLE`.
  - `DELETE /api/ticket-types/:ticketTypeId/seats/:seatId`: Menghapus kursi tertentu (Hanya diizinkan jika `seats.status == "AVAILABLE"`).

---

### FD-6.0: Order, Checkout, Payment Webhook & E-Ticket (QR) Lifecycle

*(Catatan: Seluruh tabel finansial dan transaksional pada modul ini — `orders`, `order_items`, `payments`, `tickets` — bersifat **Append-Only / No Delete**).*

#### F-6.1 Ticket Checkout & Inventory Reservation
- **Akses**: `USER` (Pengguna terautentikasi).
- **Endpoint**:
  - `POST /api/orders/checkout`: Membuat pesanan tiket baru dan menahan (*reserve*) kuota selama **15 menit**.
- **Headers Wajib**:
  - `Idempotency-Key`: string unik (UUID) yang disimpan di Redis untuk mencegah *duplicate order* akibat klik ganda atau *network retry*.
- **Input**:
  ```json
  {
    "items": [
      {
        "ticketTypeId": "uuid-ticket-type",
        "pricingPhaseId": "uuid-pricing-phase",
        "quantity": 2,
        "seatIds": ["uuid-seat-1", "uuid-seat-2"] // Wajib untuk kategori SEATED (jumlah == quantity); kosongkan untuk kategori AREA
      }
    ],
    "paymentMethod": "QRIS"
  }
  ```
- **Aturan Bisnis & Concurrency Control**:
  1. Event wajib berstatus `PUBLISHED`, dan `pricing_phase` yang dipilih wajib sedang berada dalam rentang waktu aktif (`start_at <= now() <= end_at`).
  2. Validasi tipe kursi kategori tiket (`ticket_types.seating_type`):
     - Jika `SEATED`: `seatIds` wajib diisi dengan jumlah tepat sama dengan `quantity` dan tidak boleh ada duplikat ID.
     - Jika `AREA`: `seatIds` wajib kosong/tidak dikirim (`400 Bad Request` jika disertakan).
  3. Membuka transaksi database (`tx`) dan mengunci baris `inventories` (`channel = "ONLINE"`) serta `pricing_phases` menggunakan **`SELECT ... FOR UPDATE`** untuk mencegah *race condition / overselling*. Khusus kategori `SEATED`, kunci juga baris `seats` yang dipilih dengan `SELECT ... FOR UPDATE` **terurut berdasarkan `seats.id`** (mencegah *deadlock*), pastikan seluruh kursi milik `ticketTypeId` tersebut dan berstatus `AVAILABLE` (jika ada yang tidak tersedia, tolak dengan `409 Conflict`).
  4. Validasi ketersediaan kuota:
     - `inventories.sold + inventories.reserved + quantity <= inventories.quota`
     - `pricing_phases.sold + quantity <= pricing_phases.quota`
     - Jika kuota tidak mencukupi, batalkan transaksi dan kembalikan `422 Unprocessable Entity`.
  5. Tambahkan `inventories.reserved += quantity`. Untuk kategori `SEATED`, ubah status `seats` menjadi `RESERVED` dengan `reserved_by_order_id = order.id` dan `reserved_until = order.expires_at`.
  6. Simpan snapshot transaksi ke tabel `orders` (`status = "PENDING"`, `total_amount` tipe `NUMERIC(15, 2)`, `expires_at = now() + 15 minutes`), `order_items` (`unit_price` dan `subtotal` immutable), serta `payments` (`status = "PENDING"`).

#### F-6.2 Order History, Detail & Reservation Expiry Release
- **Akses**: `USER` (Pemilik pesanan) dan `ADMIN`/`SUPER_ADMIN`.
- **Endpoint**:
  - `GET /api/orders`: Riwayat pesanan pengguna yang sedang login (mendukung `BaseListParams` + filter `status`).
  - `GET /api/orders/:id`: Detail pesanan beserta rincian `order_items`, status `payments`, dan daftar `tickets` (termasuk `qrToken` jika pesanan sudah `PAID`).
  - `POST /api/orders/:id/cancel`: Membatalkan pesanan yang masih berstatus `PENDING` secara mandiri.
- **Aturan Bisnis Expiration (TTL 15 Menit)**:
  - Jika pesanan `PENDING` telah melewati `expires_at` (atau dibatalkan pengguna), status `orders` dan `payments` diubah menjadi `EXPIRED` / `CANCELLED`, kuota yang ditahan dikembalikan ke `inventories` (`inventories.reserved -= quantity`), dan kursi terkait (`seats.reserved_by_order_id == order.id`) dikembalikan ke `status = "AVAILABLE"` (`reserved_by_order_id = NULL`, `reserved_until = NULL`) dalam satu transaksi database.

#### F-6.3 Payment Webhook & E-Ticket (QR) Generation
- **Akses**: Publik / Payment Gateway Callback (Dilindungi verifikasi signature token provider).
- **Endpoint**:
  - `POST /api/payments/webhook`: Menerima notifikasi status pembayaran dari Payment Gateway (Midtrans Sandbox / Mock Provider).
- **Aturan Bisnis**:
  - **Idempotent**: Jika `orders.status` sudah `PAID`, abaikan pemrosesan ulang dan langsung kembalikan `200 OK`.
  - Jika status pembayaran sukses (`PAID` / `SETTLEMENT`):
    1. Dalam 1 transaksi database (`SELECT ... FOR UPDATE`): ubah `payments.status = "PAID"`, `payments.paid_at = now()`, dan `orders.status = "PAID"`.
    2. Finalisasi kuota pada `inventories` (`channel = "ONLINE"`): `reserved -= quantity` dan `sold += quantity`, serta tambahkan `pricing_phases.sold += quantity`. Untuk item kategori `SEATED`, ubah status kursi terkait di `seats` menjadi `SOLD`.
    3. **Generate E-Ticket (`tickets`)**: Buat record tiket individu sebanyak total `quantity` pada `order_items` dengan `status = "ISSUED"`, `ticket_code` unik, `seat_id` (jika kategori `SEATED`), dan `qr_token` yang dienkripsi menggunakan **AES-256-GCM** (memuat payload JSON terenkripsi: `ticketId`, `orderId`, `eventId`).
    4. Kirim email bukti pembayaran & e-ticket ke `orders.buyer_email`.
  - Jika status pembayaran gagal/kedaluwarsa (`EXPIRED` / `FAILED`) dan pesanan masih `PENDING`: ubah status menjadi `EXPIRED`, lepaskan `inventories.reserved -= quantity`, dan kembalikan status kursi terkait di `seats` menjadi `AVAILABLE`.

#### F-6.4 Gate Check-In & QR Ticket Validation
- **Akses**: `ORGANIZER` (Lead Organizer & Co-Organizer) dan `STAFF` yang ditugaskan pada event terkait di tabel `event_organizers`.
- **Endpoint**:
  - `POST /api/events/:eventId/tickets/check-in`: Memindai dan memvalidasi QR e-ticket penonton di gerbang masuk (*gate*).
- **Input**:
  ```json
  {
    "qrToken": "aes-256-gcm-encrypted-base64-string"
  }
  ```
- **Aturan Bisnis**:
  1. Verifikasi bahwa pemindai (`actor`) adalah Lead Organizer (`events.organizer_id`) atau terdaftar di `event_organizers` untuk `eventId` tersebut. Jika bukan, tolak dengan `403 Forbidden`.
  2. Dekripsi `qrToken` menggunakan **AES-256-GCM**. Jika gagal didekripsi atau `eventId` di dalam payload tidak cocok dengan `:eventId` pada URL, tolak dengan `400 Bad Request` ("Tiket tidak valid atau palsu").
  3. Kunci baris `tickets` dengan `SELECT ... FOR UPDATE`:
     - Jika `tickets.status == "USED"`, tolak dengan `409 Conflict` beserta waktu `checked_in_at` sebelumnya ("Tiket sudah digunakan").
     - Jika `tickets.status == "ISSUED"`, ubah status menjadi `USED`, set `checked_in_at = now()`, dan kembalikan detail pemilik tiket, kategori tiket, serta label kursi jika kategori `SEATED` (`200 OK`).
