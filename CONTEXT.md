# Point of Sale

Sistem kasir (point of sale) ringan untuk satu toko dan satu terminal — mencatat penjualan, mencetak struk, dan mengurangi stok.

## Language

**Produk**:
Barang yang dijual di kasir. Punya nama, harga, stok, dan `Kode` (opsional).
_Avoid_: Barang, item (item dipakai untuk baris di dalam Penjualan)

**Kode**:
String unik opsional yang dipakai untuk menemukan Produk dengan cepat — bisa berupa barcode (discan) atau kode internal (diketik). Produk tanpa Kode dicari lewat nama.
_Avoid_: SKU, barcode (sebagai konsep terpisah)

**Stok**:
Jumlah unit Produk yang tersedia. Di-set saat Produk dibuat, ditambah lewat penambahan manual, berkurang otomatis saat Produk terjual, dan tidak boleh negatif. Tidak ikut berubah saat Produk diubah.
_Avoid_: Inventory, persediaan

**Stok menipis**:
Keadaan Produk Aktif yang Stok-nya di bawah ambang — masuk daftar yang harus ditambah Admin, urut dari yang paling sedikit. Ambang adalah Stok pertama yang masih cukup, jadi Produk tepat di angka ambang belum menipis. Stok 0 disebut **Habis** dan tetap bagian dari daftar yang sama. Ambangnya satu konstanta domain (saat ini 5), bukan pengaturan per toko — layar Pengaturan belum ada.
_Avoid_: Low stock, stok kritis, out of stock

**Struk**:
Bukti cetak Penjualan yang diberikan ke pembeli — berisi info toko, daftar Item, total, metode bayar, dan (bila Tunai) jumlah bayar & Kembalian. Template-nya berupa blok sederhana yang isinya bisa diatur.
_Avoid_: Nota, kuitansi, resi

**Penjualan**:
Transaksi yang sudah selesai (checkout) dan tercatat permanen. Berisi Item-item dan satu Pembayaran. Bersifat final — tidak bisa dibatalkan/di-void.
_Avoid_: Transaksi, order, sale

**Item**:
Satu baris di dalam Penjualan — satu Produk, jumlah (qty) bulat per unit, dan harga saat transaksi (nama & harga disalin, bukan referensi ke Produk).
_Avoid_: Baris produk, line item

**Pembayaran**:
Catatan metode (Tunai/QRIS/Debit/Transfer) dan nominal untuk satu Penjualan. Non-tunai hanya dicatat, tidak diproses lewat gateway.
_Avoid_: Payment

**Tunai**:
Pembayaran dengan uang fisik; bila nominalnya melebihi total, menghasilkan Kembalian.
_Avoid_: Cash

**Kembalian**:
Selisih uang yang dikembalikan ke pembeli saat Pembayaran Tunai melebihi total Penjualan.
_Avoid_: Change

**Pengguna**:
Akun login (username + password) yang memakai aplikasi; punya satu Peran.
_Avoid_: staff, operator

**Kasir**:
Peran Pengguna yang boleh menjual — buka keranjang, checkout, cetak Struk. Namanya tercetak di Struk.
_Avoid_: Staff, operator

**Admin**:
Peran Pengguna yang boleh mengelola Produk, Stok, Pengaturan (termasuk template Struk), dan melihat laporan.
_Avoid_: Manager, owner

**Kategori**:
Label opsional satu level untuk mengelompokkan Produk; hanya untuk filter/pengelompokan di daftar produk, tidak memengaruhi harga.
_Avoid_: Group, jenis

**Nonaktif**:
Status Produk yang tidak lagi muncul di lookup kasir. Produk yang pernah terjual hanya bisa dinonaktifkan, bukan dihapus permanen.
_Avoid_: Archived, inactive, hidden

**Nomor Struk**:
Nomor urut global yang unik pada setiap Penjualan — tercetak di Struk dan dipakai untuk reprint/buka ulang transaksi. Tidak pernah dipakai ulang, tidak reset harian.
_Avoid_: Invoice number, nomor nota
