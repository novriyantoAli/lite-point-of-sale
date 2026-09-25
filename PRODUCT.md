# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

**Pengguna utama: pemilik toko itu sendiri.** Dia duduk di meja kasir dengan laptop
(keyboard + mouse) dan mengerjakan seluruh pekerjaan toko dari satu layar: melayani
pembeli di Kasir, mengelola Produk dan Stok, memeriksa Laporan, mengatur template
Struk, dan mengambil Backup. Dia bukan operator terlatih — tidak ada pelatihan, tidak
ada manual, dan tidak ada orang lain yang bisa ditanya saat layar membingungkan.

**Dia belum pernah memakai sistem kasir — atau sistem apa pun — sebelumnya**
(dikonfirmasi pemilik, 2026). Konsekuensinya mengikat: tidak ada istilah teknis, tidak
ada kontrol yang hanya berupa ikon, dan tidak ada pola antarmuka yang boleh dianggap
sudah dikenal. Kejelasan dan kesederhanaan bukan preferensi estetika di produk ini,
melainkan syarat supaya pemakainya bisa bekerja sama sekali.

Peran **Kasir** tetap ada di sistem sebagai pembatas hak akses, bukan sebagai orang
kedua: pada pemakaian nyata sekarang satu orang memegang dua peran itu sekaligus.
Pemisahan Admin/Kasir harus terasa sebagai hak akses, bukan sebagai dua aplikasi.

**Audiens kedua: pembeli di seberang meja.** Dia tidak pernah menyentuh aplikasi, tapi
membaca keluarannya — **Struk** cetak (58/80 mm) adalah satu-satunya artefak yang
sampai ke tangannya, dan itu bagian dari wajah toko.

Situasi pemakaian: berdiri/duduk di meja, pembeli menunggu, satu transaksi harus
selesai dalam hitungan detik.

## Product Purpose

Kasir ringan untuk **satu toko, satu terminal** (ADR-0002): mencatat Penjualan,
mencetak Struk, dan mengurangi Stok — tanpa akuntansi, tanpa gateway pembayaran,
tanpa perangkat tambahan di luar laptop dan printer thermal.

Produk ini dipakai di toko nyata milik pemiliknya, bukan demo (dikonfirmasi
pemilik), jadi "berhasil" berarti: transaksi di depan pembeli selesai lebih cepat
daripada mencatat manual, Omzet harian bisa dipercaya tanpa menghitung ulang, dan
tidak ada Stok yang hilang tanpa jejak Penjualan.

## Positioning

Ringan dan **mandiri di satu mesin** — dan itu bukan slogan, itu mekanisme:

- Struk dicetak **langsung dari Go sebagai ESC/POS mentah** ke printer thermal,
  tanpa JVM dan tanpa renderer laporan pihak ketiga (ADR-0003).
- **Non-tunai hanya dicatat** sebesar total, tanpa gateway (ADR-0016).
- Sesi dipegang SvelteKit sebagai cookie httpOnly berisi token internal Go — browser
  tidak pernah melihat tokennya (ADR-0001, ADR-0010).
- Seluruh Pengaturan toko adalah **satu baris** (template Struk, lebar kertas, ambang
  Stok menipis), dan schema sengaja tidak punya `store_id` maupun konsep terminal
  (ADR-0002).

Produk kasir lain bisa meniru daftar fiturnya, tapi tidak bisa mengklaim hal yang
sama: satu proses Go + satu file SQLite + satu printer, dijalankan pemilik toko
sendiri.

## Operating Context

- **Satu toko, satu terminal.** Tidak ada multi-outlet, tidak ada sinkronisasi antar
  mesin, tidak ada dua kasir bersamaan (ADR-0002). Menambahnya berarti migrasi schema
  + desain sinkronisasi Stok.
- **Perangkat meja kasir:** laptop, keyboard + mouse, printer thermal USB
  (`POS_PRINTER_DEVICE`, tanpa default — tebakan yang salah lebih buruk daripada
  "belum diatur", ADR-0017).
- **Alur utama Kasir:** cari Produk lewat **Kode** (barcode discan atau kode internal
  diketik) atau nama → susun keranjang → checkout → bayar (Tunai / QRIS / Debit /
  Transfer) → Struk tercetak otomatis. Tunai menghasilkan Kembalian; non-tunai
  dicatat sebesar total.
- **Cetak yang gagal tidak menggagalkan Penjualan.** Printer kosong/belum diatur →
  pesan di layar + tombol **Cetak ulang**; Penjualan tetap tersimpan dan bisa dibuka
  lagi lewat **Nomor Struk** di layar Penjualan, dicetak memakai template Pengaturan
  saat itu (ADR-0017).
- **Pengaturan** diisi Admin: blok teks **header** dan **footer** Struk (baris
  dipisah newline, dicetak verbatim), lebar kertas 58/80 mm, ambang Stok menipis.
- **Stok menipis** adalah satu daftar berurutan (termasuk Stok 0 = Habis) yang
  dipakai Admin untuk tahu apa yang harus ditambah.
- **Backup** memakai snapshot file SQLite: otomatis harian + manual, retensi N hari.
- **Semua permukaan berbahasa Indonesia**, dan istilah domainnya mengikuti
  `CONTEXT.md` (Produk, Kode, Stok, Struk, Penjualan, Item, Pembayaran, Kembalian,
  Pengguna, Kasir, Admin, Pengaturan, Backup, Kategori, Nonaktif, Nomor Struk,
  Laporan, Omzet harian) — bukan padanan Inggrisnya.

## Capabilities and Constraints

Tersedia (satu slice domain per konsep, ADR-0006): auth/sesi + Pengguna & Peran,
Produk (Kode opsional, Kategori, Nonaktif), Stok (penambahan manual, daftar menipis),
Penjualan (keranjang, checkout atomik, Nomor Struk), Pembayaran, Struk (cetak &
reprint), Pengaturan, Laporan (Omzet harian + daftar Penjualan), Backup, health.

Batas yang mengikat:

- **Penjualan bersifat final** — tidak ada void, refund, atau pembatalan.
- **Tidak ada gateway pembayaran**; tidak ada pajak, diskon, atau promo di schema.
- **Laporan = rekap**, bukan grafik/analitik, dan bukan laba; yang ada hanya Omzet
  harian (per metode Pembayaran dan per Kasir) plus daftar Penjualan.
- **Nama toko tidak punya field terstruktur.** Ia hidup sebagai teks di dalam blok
  `header` Pengaturan. Keputusan terbuka: bagaimana layar membaca nama toko dari blok
  bebas itu (baris pertama? field baru di schema?) — dan mengubah schema berarti
  perubahan kontrak Go ↔ frontend.
- **Keputusan terbuka:** apakah toko akan punya pegawai Kasir yang login sendiri;
  aset visual toko (logo/warna) belum ada; tidak ada kebutuhan aksesibilitas spesifik
  yang ditetapkan pemilik.
- **Cacat yang diketahui:** `src/app.html` masih `lang="en"` padahal seluruh UI
  berbahasa Indonesia.

## Brand Commitments

- **Nama tokolah yang harus tampil di layar** (dikonfirmasi pemilik): aplikasi menjadi
  latar, toko yang tampil. Teks "Lite Point of Sale" tidak lagi otomatis menjadi
  identitas utama di header.
- **Voice:** bahasa Indonesia, tenang dan faktual. Permukaan aplikasi memakai istilah
  domain Indonesia, bukan istilah Inggris.
- **Belum ada** logo, warna brand, atau aset visual toko yang ditetapkan. Nama toko
  literalnya tidak dicatat di sini karena berasal dari Pengaturan saat runtime, bukan
  dari repo.

## Evidence on Hand

- Pemilik repo ini memakai sistemnya di toko nyata — pemakaian nyata, bukan studi kasus.
- Kontrak domain lengkap: `CONTEXT.md` (glosarium, ADR-0013) dan `docs/adr/0001`–`0018`.
- `README.md` memuat alur nyata, env, dan batas operasional yang sudah dijalankan.
- Test suite nyata: unit Vitest per domain (`*.test.ts` di sebelah komponen/schema) dan
  e2e Playwright (`frontend/tests/e2e/`), dijalankan CI.
- **Yang tidak boleh dikarang:** tidak ada testimoni, jumlah toko/pelanggan, metrik
  performa, harga, jaminan, screenshot, data Penjualan nyata, atau aset brand di repo.

## Product Principles

1. **Melayani pembeli dulu — dan tanpa diajari.** Layar yang dibuka saat transaksi
   berjalan harus bisa diselesaikan cepat dengan keyboard, dan harus bisa dipahami
   orang yang belum pernah memakai sistem kasir: setiap kontrol membawa katanya,
   setiap keadaan ditulis, tidak ada aksi tersembunyi di balik ikon.
2. **Satu orang, bukan dua mode.** Jangan paksa pengguna berpindah "sisi" Admin/Kasir;
   hak akses berbeda, aplikasinya satu.
3. **Uang dan Stok tidak boleh berbohong.** Setiap angka harus bisa ditelusuri ke satu
   Penjualan/Nomor Struk, dan tidak ada operasi yang diam-diam mengubah data.
4. **Kegagalan perangkat bukan kegagalan transaksi.** Printer mati tidak menghapus
   Penjualan, dan Struk selalu bisa dicetak ulang.
5. **Yang tampil adalah tokonya, bukan aplikasinya.** Identitas toko memimpin;
   identitas produk mengalah.
