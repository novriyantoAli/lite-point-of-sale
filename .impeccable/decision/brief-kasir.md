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

## Unresolved

Apakah dunia ini dipakai untuk seluruh aplikasi atau hanya layar Kasir. Belum diputuskan, dan tidak perlu diputuskan di ronde ini.

## Direction contract

**THESIS.** Kasir adalah mosaik modul berbingkai rambut yang memadatkan katalog ke satu layar, dan ia menolak susunan default kategori ini: sidebar ikon, empat kartu KPI, dan tabel berbayang lembut yang memakan ruang tanpa membawa informasi.

**OWN-WORLD.** Dasar `#fafafa`, petak `#ffffff`, garis rambut `#e8e8e8`, isian hover `#f5f5f5`, tinta `#000000`, dan satu merah utilitas `#cc0d0d` yang hanya boleh muncul di tab dan harga, di bawah 3% permukaan. Tidak ada abu-abu sebagai tinta huruf: abu-abu hanya garis dan isian. Hierarki dibawa kontras skala dan kepadatan. Tombol adalah petak berbingkai, bukan kapsul. Semua angka `tabular-nums`.

**STORY.** Kasir memahami bahwa seluruh katalog ada di depan matanya dan tidak ada yang tersembunyi di balik menu; ia percaya angka Stok dan harga yang dibacanya; ia menekan satu petak, menambah ke keranjang, dan menyimpan.

**FIRST VIEWPORT.** Rel navigasi mendatar dua baris menempel di tepi atas (Beranda/Kasir/Penjualan, lalu Produk sampai Backup). Di bawahnya strip judul selebar papan: "Kasir" pada 24px/700 dengan satu kalimat penjelas 12px di sebelahnya — disetujui pemilik, dan satu-satunya tempat ukuran 24px muncul. Lalu mosaik memadat tepi ke tepi: kiri, pencarian Kode dan Nama sebagai dua petak; tengah, katalog sebagai mosaik petak 4–6 kolom, masing-masing berkepala tab kecil berisi Kode dan bertumit harga di kanan bawah; kanan, keranjang sebagai satu kolom petak bertumpuk dengan total berjalan di kepala kolomnya. Aksi utama adalah petak bertinta penuh di dasar kolom keranjang, satu-satunya bidang gelap di layar.

**FORM.** Dunia katalog `japanese-high-density-web`, menang dari tangan yang di-deal pada ronde keempat. Ia bukan kandidat dari daftar grounded saya: dadu menugaskan Cable Tie, dan tangan itu kalah di dua sumbu. Seed key `22ccabe4`.

**FINISH.** unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
