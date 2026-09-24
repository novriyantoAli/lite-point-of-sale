# Laporan: omzet harian & daftar Penjualan

#9 menambahkan layar **Laporan** untuk Admin: **Omzet harian** (total, jumlah transaksi, pecahan per metode **Pembayaran** dan per **Kasir**) dan **daftar Penjualan** (dengan Nomor Struk) yang bisa dibuka dan di-reprint. Ia berdiri di atas `GET /api/penjualan/{nomorStruk}` dan `FindByReceiptNumber` yang ADR-0015 sudah janjikan, dan menutup konsekuensi "belum punya pembaca di frontend" di ADR-0015 dan ADR-0017.

## Keputusan

### 1. Laporan tinggal di slice `penjualan`, bukan slice `laporan` baru

| Lapisan | Nama | Contoh |
| --- | --- | --- |
| Glosarium, label UI | Indonesia | `CONTEXT.md`, "Laporan", "Omzet harian", "Daftar Penjualan" |
| Route | Indonesia | `/laporan` (layar), `/api/penjualan` dan `/api/penjualan/omzet` |
| Domain frontend | Indonesia, di dalam slice `penjualan` | `LaporanHarian.svelte`, `PenjualanRingkasSchema`, `OmzetHarianSchema`, `laporanState` |
| Identifier Go | Inggris, di package `penjualan` | `domainpenjualan.SaleSummary`, `DailyRevenue`, `MethodTotal`, `CashierTotal` |
| DTO JSON | Inggris | `{"data":{"date":"2026-09-23","total":72000,"transactions":3,"by_method":[{"method":"cash","total":36000,"transactions":2}],"by_cashier":[…]}}` |

Laporan bukan aggregate baru: ia **membaca** tabel `penjualan` dan tidak menulis apa pun. Presedennya sudah ada di repo ini — laporan **Stok menipis** tinggal di slice `produk` (`usecase/produk.LowStock`, `GET /api/produk/stok-menipis`), bukan slice `laporan` tersendiri (ADR-0017, keputusan 3). Route API-nya pun tetap di bawah `/api/penjualan` karena data yang dilaporkan adalah Penjualan; hanya **layar**-nya yang bernama `/laporan`, dan nama itu yang dipakai Admin.

Ditolak: slice `laporan` baru dengan repository kedua di atas tabel `penjualan`. Itu berarti dua tempat yang harus dijaga selaras saat bentuk Penjualan berubah, tanpa satu pun aturan yang hanya dimiliki laporan.

### 2. Daftar Penjualan adalah read model ringkas, detailnya dibaca ulang

`GET /api/penjualan` menjawab **`SaleSummary`** — Nomor Struk, waktu, Kasir, total, metode — bukan `Sale` penuh. Daftar tidak pernah menampilkan baris Item, dan membaca Item tiap Penjualan berarti satu query per baris. Membuka satu baris membaca **`GET /api/penjualan/{nomorStruk}`** yang sudah ada, dan reprint memakai **`POST /api/penjualan/{nomorStruk}/struk`** yang sudah ada: layar baru tidak menambah jalur tulis atau baca baru di atas tabel yang sama.

Urutannya **Nomor Struk menurun** (terbaru dulu), karena itu yang dicari Admin saat membuka laporan hari ini.

Ditolak: `Sale` penuh per baris (N+1 query untuk kolom yang tidak ditampilkan), dan endpoint daftar tersendiri yang mengembalikan Item-nya.

### 3. Omzet selalu menjawab keempat metode, urut dari till

`usecase/penjualan.DailyRevenue` menjawab `by_method` dengan **satu baris per metode**, termasuk metode yang tidak dipakai hari itu (nol). Repository sengaja hanya mengembalikan metode yang muncul; **pengisian baris nol terjadi di use case**, tempat aturan domain itu bisa diuji tanpa database. `METODE_URUT` di `penjualan.schema.ts` tetap **satu sumber urutan tampil** (ADR-0016, keputusan 3): layar mengiterasi `METODE_URUT` dan mencari angkanya di jawaban API, bukan mengikuti urutan array yang kebetulan dikirim.

`by_cashier` dikelompokkan per **id Pengguna**, bukan per nama salinan, supaya satu Kasir tetap satu baris; namanya adalah nama yang tersalin di Penjualannya. Di aplikasi ini Pengguna tidak bisa di-rename (hanya dibuat dan diaktifkan/nonaktifkan), jadi tidak ada dua nama yang bisa bersaing.

Ditolak: mengembalikan hanya metode yang terpakai (layar lalu mengarang baris nol — aturan kedua yang bisa menyimpang), dan menghitung agregasi di repository (aturan bisnis di adapter, ADR-0015 keputusan 2).

### 4. Tiga agregat dalam satu transaksi baca

`DailyRevenue` menjalankan tiga agregat — total, pecahan per metode, pecahan per Kasir — di dalam **satu transaksi baca**. Checkout yang mendarat di antara ketiganya akan membuat total tidak lagi sama dengan jumlah pecahannya, dan Admin yang membaca laporan justru sedang berdampingan dengan Kasir yang menjual. SQLite berjalan di mode WAL, jadi transaksinya memegang snapshot yang konsisten tanpa memblokir tulisnya till.

### 5. Hari adalah tanggal lokal toko, dan tanggalnya bisa dipilih

Filter laporan adalah **`?date=YYYY-MM-DD`** dalam waktu lokal toko — clock yang sama dengan `created_at` (ADR-0015, keputusan 5). Tanggal kosong berarti **hari ini**, dibaca di use case. Tanggal yang tidak berbentuk `YYYY-MM-DD` ditolak 400 dengan pesan, bukan diam-diam tidak mencocokkan apa pun: perbandingannya tekstual terhadap `date(created_at)`.

Layar `/laporan` memberi satu field tanggal yang default-nya hari ini. Tanggal yang bisa dipilih adalah kebutuhan laporan harian — tanpa itu, laporan hanya bisa menampilkan hari ini dan hari kosong tidak bisa dilihat. Field kosong dibaca sebagai hari ini, sama seperti API.

### 6. Laporan adalah layar Admin; lookup `/penjualan` tetap milik kedua Peran

`GET /api/penjualan` dan `GET /api/penjualan/omzet` ada di belakang **role guard Admin**, dan `/laporan` masuk daftar `ADMIN_ONLY` di `hooks.server.ts`. Omzet adalah pemasukan toko, bukan layar till.

Pintu masuk reprint yang ADR-0017 buka untuk Kasir **tidak berubah**: layar `/penjualan` (ketik Nomor Struk → tampilkan → cetak) tetap bisa diakses kedua Peran. Daftar Penjualan yang bisa di-scroll adalah cara Admin menemukan Penjualan tanpa tahu Nomor Struk-nya lebih dulu.

Ditolak: membuka laporan omzet untuk Kasir, dan memindahkan reprint dari `/penjualan` ke `/laporan` (Kasir kehilangan jalan cetak ulang yang sudah ada).

### 7. Kedua route laporan adalah literal di samping wildcard Nomor Struk

`GET /api/penjualan/omzet` didaftarkan sebagai pola literal di samping `GET /api/penjualan/{receiptNumber}`. Pola literal lebih spesifik, jadi ia menang — sama seperti `/api/produk/kategori` dan `/api/produk/stok-menipis` di ADR-0015, keputusan 7. "omzet" bukan Nomor Struk. Hal yang sama berlaku di SvelteKit: `api/penjualan/omzet/+server.ts` adalah segmen statis di samping `api/penjualan/[receiptNumber]/+server.ts`.

## Considered Options

- **Slice `laporan` baru** — ditolak (keputusan 1): repository kedua di atas tabel yang sama, tanpa aturan yang hanya dimiliki laporan. Laporan Stok menipis sudah memilih pola sebaliknya.
- **Daftar berisi `Sale` penuh** — ditolak (keputusan 2): satu query per baris untuk Item yang tidak pernah ditampilkan.
- **Endpoint daftar dan omzet di bawah `/api/laporan`** — ditolak: datanya Penjualan, dan route lain di repo ini tinggal di slice yang memiliki datanya (`stok-menipis` di bawah `produk`).
- **Agregasi di repository** — ditolak (keputusan 3): itu aturan bisnis, dan ADR-0015 keputusan 2 sudah menolak menghitung total di adapter.
- **Hanya metode yang terpakai** — ditolak (keputusan 3): layar lalu harus tahu aturan baris nolnya sendiri.
- **Paginasi daftar Penjualan** — ditolak: di luar scope MVP; batasnya adalah satu hari, bukan halaman.
- **Laporan terbuka untuk kedua Peran** — ditolak (keputusan 6): omzet adalah pemasukan toko; Kasir sudah punya pintu reprint-nya sendiri.
- **Tanggal selalu hari ini (tanpa field tanggal)** — ditolak (keputusan 5): laporan harian harus bisa membaca hari lain, dan hari kosong tidak bisa diperiksa.

## Consequences

- **ADR-0015 "belum punya pembaca di frontend" terpenuhi.** `FindByReceiptNumber` dan `GET /api/penjualan/{nomorStruk}` kini punya pembaca: layar `/penjualan` (#29) dan daftar Laporan (#9). ADR-0017 keputusan 5 "menumbuhkan layar ini atau menambah `/laporan` Admin-only" dijawab dengan **keduanya**: `/penjualan` tetap untuk lookup Kasir, `/laporan` untuk omzet dan daftar Admin.
- **`CONTEXT.md` diperbarui saat #9 mendarat**: entri **Laporan** dan **Omzet harian** ditambahkan. Glosarium tidak mendahului kode.
- **Laporan membaca snapshot, bukan angka yang disimpan.** Tidak ada tabel laporan dan tidak ada kolom omzet: setiap pembacaan menjumlahkan Penjualan hari itu, jadi laporan tidak bisa menyimpang dari Penjualannya. Konsekuensinya, hari yang Penjualannya dihapus manual dari database ikut berubah — konsekuensi yang diterima di MVP.
- **Nomor Struk adalah identitas satu-satunya** yang dibutuhkan layar untuk membuka detail dan mencetak ulang; daftar tidak menyimpan apa pun di cache selain halaman yang sedang dibaca.
- Tes yang menjaga keputusan ini: `backend/internal/usecase/penjualan/laporan_test.go` (aturan: tanggal default/ditolak, keempat metode terisi, daftar terbaru dulu), `backend/internal/adapter/sqlite/penjualan_repository_test.go` (agregat per metode dan per Kasir, hari kosong, urutan), `backend/tests/e2e/laporan_test.go` (seam REST: agregasi, daftar, hari kosong, 400 tanggal, 401/403), `frontend/src/lib/domains/penjualan/{schemas,api,state,components}/*.test.ts`, `frontend/src/routes/api/penjualan/{proxy.test.ts,omzet/proxy.test.ts}`, dan `frontend/tests/e2e/laporan.spec.ts` (browser sungguhan: omzet, daftar, buka detail, cetak ulang dari daftar, guard Admin).
