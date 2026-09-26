---
version: 1
slug: "frontend-src-routes-app-kasir-page-svelte"
primary_target: "frontend/src/routes/(app)/kasir/+page.svelte"
related_targets: []
---

## Scope

Layar **Kasir** (`/(app)/kasir`) — permukaan pertama dunia baru, dan satu-satunya permukaan yang dibangun di ronde ini. Mode **Operate**: pemilik toko menyelesaikan satu tugas, melayani pembeli di depan meja.

## Audience, job, action

Pemilik toko elektronik, sendiri, di laptop meja kasir, keyboard dan mouse. Dia belum pernah memakai sistem kasir, atau sistem apa pun, sebelumnya (PRODUCT.md). Tugasnya: temukan Produk yang benar di antara banyak kotak yang mirip, susun keranjang, terima bayar, simpan Penjualan, cetak Struk. Aksi utamanya satu: **Bayar & Simpan Penjualan**.

## Constraints

Empat yang dikonfirmasi pemilik, semuanya mengikat: kecepatan di depan pembeli tidak boleh bertambah lambat; keterbacaan untuk pemula; wibawa alat kerja; istilah domain CONTEXT.md tetap apa adanya.

## Keputusan yang mengikat permukaan ini

- **Strip judul 24px disetujui pemilik.** Hierarki huruf dibawa kontras skala, dan strip itulah lompatan skalanya.
- **Aksi tidak pernah disembunyikan di balik hover.** Dunia ini memunculkan aksi sekunder saat hover; permukaan ini menolaknya, karena pemilik belum pernah memakai sistem dan aksi tersembunyi adalah kosakata yang harus dipelajari lebih dulu. Penolakan ini disengaja, bukan kekurangan.
- **Tombol `−` dan `+` tetap teks operator**, mengikuti aplikasi yang ada, bukan ikon yang digambar. Keputusan pemilik.
- **Keadaan memuat belum diuji di sini**: prototipe tidak punya jaringan. Keadaan memuat aplikasi harus diport dan diverifikasi saat dunia ini masuk ke `frontend/src`.

## Keputusan yang menutup pertanyaan ronde ini

**Dijawab pemilik, 2026-09-26: dunia ini dipakai untuk SELURUH aplikasi, bukan hanya layar Kasir.**
Konsekuensinya berurutan: rel navigasi dan nama toko dikerjakan lebih dulu, karena kedelapan
layar lain duduk di dalam kerangka itu (`frontend/src/routes/(app)/+layout.svelte`), lalu tiap
layar diport satu per satu. Keputusan ini menutup pertanyaan "Unresolved" ronde sebelumnya;
tidak ada yang tersisa di sana.

Catatan yang belum ditutup: **cacah pada tab** (`Produk 24`, `Stok 4`) ada di prototipe sebagai
angka contoh. Angka itu butuh permintaan ke API dari kerangka, yang berarti setiap layar
membayar dua permintaan demi kerangka — belum diputuskan, jadi tab untuk sekarang membawa
katanya saja, tanpa cacah.

## Direction contract

**THESIS.** Kasir adalah mosaik modul berbingkai rambut yang memadatkan katalog ke satu layar, dan ia menolak susunan default kategori ini: sidebar ikon, empat kartu KPI, dan tabel berbayang lembut yang memakan ruang tanpa membawa informasi.

**OWN-WORLD.** Dasar `#fafafa`, petak `#ffffff`, garis rambut `#e8e8e8`, isian hover `#f5f5f5`, tinta `#000000`, dan satu merah utilitas `#cc0d0d` yang hanya boleh muncul di tab dan harga, di bawah 3% permukaan. Tidak ada abu-abu sebagai tinta huruf: abu-abu hanya garis dan isian. Hierarki dibawa kontras skala dan kepadatan. Tombol adalah petak berbingkai, bukan kapsul. Semua angka `tabular-nums`.

**STORY.** Kasir memahami bahwa seluruh katalog ada di depan matanya dan tidak ada yang tersembunyi di balik menu; ia percaya angka Stok dan harga yang dibacanya; ia menekan satu petak, menambah ke keranjang, dan menyimpan.

**FIRST VIEWPORT.** Rel navigasi mendatar dua baris menempel di tepi atas (Beranda/Kasir/Penjualan, lalu Produk sampai Backup). Di bawahnya strip judul selebar papan: "Kasir" pada 24px/700 dengan satu kalimat penjelas 12px di sebelahnya — disetujui pemilik, dan satu-satunya tempat ukuran 24px muncul. Lalu mosaik memadat tepi ke tepi: kiri, pencarian Kode dan Nama sebagai dua petak; tengah, katalog sebagai mosaik petak 4–6 kolom, masing-masing berkepala tab kecil berisi Kode dan bertumit harga di kanan bawah; kanan, keranjang sebagai satu kolom petak bertumpuk dengan total berjalan di kepala kolomnya. Aksi utama adalah petak bertinta penuh di dasar kolom keranjang, satu-satunya bidang gelap di layar.

**FORM.** Dunia katalog `japanese-high-density-web`, menang dari tangan yang di-deal pada ronde keempat. Ia bukan kandidat dari daftar grounded saya: dadu menugaskan Cable Tie, dan tangan itu kalah di dua sumbu. Seed key `22ccabe4`.

**FINISH.** unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance

## Ronde port ke `frontend/src`, 2026-09-26

Prototipe sekarang hidup sebagai kode di `frontend/src/lib/domains/penjualan/components/`:
`Kasir.svelte` (papan tiga kolom + wajah struk), `PencarianProduk.svelte` (kolom kiri),
`KatalogProduk.svelte` (kolom tengah, pemilik query katalog), `Keranjang.svelte` +
`Pembayaran.svelte` (kolom kanan), `StrukPenjualan.svelte`, `RincianPenjualan.svelte`,
`CetakStruk.svelte`. Saringan katalog pindah ke `state/pencarian.state.svelte.ts` karena dua
kolom bersebelahan membacanya (skill §6.4).

### Keputusan yang muncul saat port

- **Keadaan memuat diuji dan dijawab** — satu-satunya butir yang prototipe tinggalkan kosong.
  Muat pertama menggambar 24 petak hantu berukuran sama dengan petak asli (73px) supaya papan
  tidak melompat saat data tiba; ada `role="status"` yang membacakan "Memuat Produk…" untuk
  pembaca layar. Saat penyaring berubah, mosaik juga menampilkan kerangka lagi — perilaku yang
  sama dengan sebelumnya, dan alasannya: menahan daftar lama akan menampilkan Produk yang tidak
  cocok dengan Kode yang baru diketik.
- **Nama aksesibilitas petak tetap satu kalimat untuk semua keadaan.** Prototipe memakai
  `"<nama> Stok habis"` pada petak yang mati; di sini petak yang habis tetap bernama
  `"Tambah <nama> ke keranjang"` dan yang menyatakan keadaannya adalah kata "Stok habis" +
  garis coret di dalamnya. Aksi adalah nama tombolnya; keadaan adalah teks yang terlihat.
- **Wajah struk menggantikan kasir, bukan menemani mosaiknya.** Prototipe menaruh struk di
  kolom keranjang sementara mosaik tetap hidup — dan itu lubang yang sudah ada di skrip
  prototipenya sendiri: petak yang ditekan setelah Penjualan tersimpan menambah keranjang yang
  tidak terlihat. Aksi tersembunyi itu yang DESIGN.md tolak, jadi layar struk berdiri sendiri
  dengan papan dua kolom: catatan Penjualan di kiri, cetak + "Penjualan Baru" di kanan.
- **Pesan galat dan peringatan memakai tinta, bukan merah.** `.note--warn` prototipe menggambar
  garis merah di atas pesan; DESIGN.md (Shapes, Zero-Grey, Three-Percent) hanya mengizinkan
  merah pada tab, harga, dan tanda field yang tidak valid. Jadi yang membawa merah di sini
  adalah field qty yang `aria-invalid` (garis + outline merah penuh, terukur `2px solid
  rgb(204,13,13)`), dan kalimatnya tetap tinta.
- **Baris keranjang menyebut harga dan Stok, tanpa tab Kode** — anatomi `.row` prototipe.
  Namanya sudah mengidentifikasi Produk, dan kolomnya hanya 300px.
- **"Kosongkan" pindah ke pita total**, bukan di kepala kolom: kepalanya tinggal nama modul +
  ringkasan, dan aksi keranjang berdiri di pita keranjangnya sendiri. Form Pembayaran tidak
  lagi mengulang "Total yang harus dibayar" — pitanya tepat di atasnya, di kolom yang sama.
- **Cincin fokus dan field tidak valid dipasang sekali untuk seluruh aplikasi** di `app.css`,
  di luar layer Tailwind (di dalam layer, `utilities` selalu menang — lihat
  `docs/solutions/developer-experience/port-layar-ke-dunia-mosaik.md`).
- **Saringan dikosongkan tiap Penjualan tersimpan.** Dulu saringan itu state komponen, jadi ia
  ikut terhapus saat komponennya dilepas; sekarang ia state modul, jadi ia harus dikosongkan
  sendiri — kalau tidak, pembeli berikutnya mewarisi Kode pembeli sebelumnya.

### Angka yang terukur (DOM aplikasi yang berjalan, 1440×900)

| Yang diukur | Hasil | Kontrak |
| --- | --- | --- |
| Papan | `232px 892px 300px` | 232 / minmax(0,1fr) / 300 |
| Petak | 73px, `minmax(154px,1fr)` | 154px ke atas |
| Satu layar | `scrollHeight 900` vs viewport 900, 14 Produk + keranjang + pembayaran | tanpa menggulir |
| Kontras | 16 pasangan unik, 0 gagal | AA |
| Radius selain nol | tidak ada | nol |
| Bayangan terlihat | tidak ada | nol |
| Merah | 0,86% dari viewport (putih 91%, tinta 1,41%) | ≤ 3% |
| Fokus | `2px solid rgb(0,0,0)`, offset −2px, pada field dan pada sel metode | garis tinta ke dalam |
| Field tidak valid | `2px solid rgb(204,13,13)` + garis merah | merah penuh, bukan tint |
| 390px | `scrollWidth 390` = `clientWidth 390`, satu kolom 374px | tanpa geser mendatar |

### Raster dan asalnya

Semuanya dari aplikasi SvelteKit yang berjalan (bukan prototipe `file://`), Chromium, device
scale 1, `font.ready` ditunggu, 2026-09-26:

- `kasir-sveltekit.png` — 1440×900, satu Produk di keranjang, jumlah bayar terisi.
- `kasir-board-sveltekit.png` — papan saja, potongan dari keadaan yang sama.
- `kasir-blokir-sveltekit.png` — keadaan terblokir: satu baris melebihi Stok.
- `kasir-struk-sveltekit.png` — wajah struk setelah Penjualan tersimpan.
- `kasir-memuat-sveltekit.png` — kerangka 24 petak, dengan `GET /api/produk` ditahan 2,5 detik.
- `kasir-mobile-sveltekit.png` — 390×844, satu kolom, halaman penuh.
