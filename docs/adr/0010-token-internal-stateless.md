# Token internal stateless: logout menghapus cookie, bukan mencabut token

SvelteKit memegang sesi dan Go memvalidasi token pada tiap panggilan (ADR-0001). Token itu **stateless**: HMAC-SHA256 atas `{username, role, exp}` yang ditandatangani `POS_TOKEN_SECRET`, tanpa tabel sesi di SQLite. Logout karena itu tidak memanggil Go sama sekali — BFF menghapus cookie httpOnly-nya dan token berhenti dipakai. Yang tetap berlaku seketika bukan tokennya, melainkan status Pengguna: Go membaca ulang Pengguna dari database pada tiap panggilan.

## Keputusan

- Token ditandatangani HMAC-SHA256 (`internal/adapter/token`), isinya username + peran + kedaluwarsa; tidak ada state di sisi Go.
- Masa berlaku dari `POS_SESSION_TTL` (default `12h`) — satu hari kerja kasir.
- `POST /api/auth/logout` **tidak ada** di API Go. Logout adalah `POST /api/auth/logout` milik SvelteKit yang menghapus cookie.
- Nonaktif Pengguna (`PATCH /api/pengguna/{id}`) langsung memutus akses: `Authenticate` membaca ulang Pengguna dan menolak yang `active = 0`, jadi tokennya tidak berguna walau belum kedaluwarsa. Perubahan Peran juga berlaku pada panggilan berikutnya, bukan saat token diperbarui.

## Considered Options

- **Tabel sesi di SQLite** (token acak + lookup per panggilan) — bisa mencabut satu token seketika, tetapi menambah satu tabel, satu tulisan tiap login, dan satu pembacaan tiap panggilan, hanya untuk melayani kasus yang belum ada: mencabut token yang sudah bocor.
- **JWT pihak ketiga** — ditolak: satu dependensi untuk sesuatu yang butuh 30 baris di `adapter/token`, dan algoritma/klaimnya justru lebih mudah salah dipakai.
- **Token disimpan di cookie berisi seluruh Pengguna** — ditolak: cookie jadi salinan data yang bisa basi (peran berubah, akun dinonaktifkan) dan ikut dikirim di tiap permintaan.

## Consequences

- Token yang sudah bocor tetap sah sampai `exp`. Mengganti `POS_TOKEN_SECRET` membatalkan **semua** sesi sekaligus — itulah jalur daruratnya, dan karena itu secret-nya wajib diganti di produksi (server mencatat peringatan selama memakai default pengembangan).
- "Nonaktifkan Pengguna" aman dipakai sebagai cara memutus sesi Kasir yang berhenti; tidak perlu fitur "keluar dari semua perangkat".
- Kalau nanti butuh pencabutan per-sesi (daftar sesi aktif, logout semua perangkat), tabel sesi harus ditambahkan — ADR ini yang direvisi lebih dulu, bukan diam-diam menambah state di Go.
- Cookie hanya berisi token, tidak ada salinan Pengguna; siapa yang login dijawab server pada tiap navigasi (`hooks.server.ts` → Go), sehingga tidak ada data sesi yang bisa basi di klien.
