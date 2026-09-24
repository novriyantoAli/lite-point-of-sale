# Penjualan: slice sendiri, checkout atomik, Nomor Struk dari transaksi

#6 menambahkan alur jual Tunai: Kasir menemukan Produk lewat Kode atau nama, menyusun keranjang, dan checkout menjadi satu **Penjualan** final — Stok berkurang, nama & harga tiap Item disalin, Nomor Struk terbit, Pembayaran Tunai tercatat berikut Kembaliannya. Beberapa keputusan yang diambil di situ mudah dipertanyakan lagi nanti, jadi dicatat di sini.

## Keputusan

### 1. Penjualan adalah vertical slice tersendiri

| Lapisan | Nama | Contoh |
| --- | --- | --- |
| Glosarium, label UI | Indonesia | `CONTEXT.md`, "Penjualan", "Keranjang", "Kembalian", "Nomor Struk" |
| Route | Indonesia | `/kasir` (layar), `/api/penjualan` |
| Domain frontend | Indonesia, di dalam slice `penjualan` | `Kasir.svelte`, `PenjualanSchema`, `CheckoutInputSchema`, `keranjangState` |
| Identifier Go | Inggris, di package `penjualan` | `domainpenjualan.Sale`, `SaleRepository`, `usecasepenjualan.Checkout` |
| DTO JSON | Inggris | `{"data":{"sale":{"receipt_number":1,"items":[{"product_id":1,"name":"Kopi","price":18000,"quantity":2,"subtotal":36000}],"payment":{"method":"cash","amount":50000,"change":14000}}}}` |

Berbeda dari Stok (ADR-0014), Penjualan punya tabelnya sendiri (`penjualan`, `penjualan_item`) dan aggregate-nya sendiri — satu Penjualan dengan Item-item dan satu Pembayaran, bukan satu field di baris Produk. Karena backend punya package `penjualan`, frontend punya `lib/domains/penjualan` (ADR-0006: schema zod mirror DTO Go 1:1).

### 2. Checkout satu transaksi; pemeriksaan Stok dua lapis dengan alasan berbeda

`usecasepenjualan.Checkout` **membaca** tiap Produk untuk menyalin nama & harga, menghitung total dan Kembalian, dan **menolak lebih dulu** keranjang yang Stoknya tidak cukup — pesannya menyebut Produk dan sisa Stoknya, karena itu yang dibutuhkan Kasir untuk membetulkan keranjang.

`SaleRepository.Create` **mengulang** pemeriksaan itu di dalam transaksi, dalam satu statement:

```sql
UPDATE produk SET stock = stock - ?, sold = 1 WHERE id = ? AND active = 1 AND stock >= ?
```

`RowsAffected() == 0` berarti Produk itu hilang, Nonaktif, atau Stoknya sudah tidak cukup — dan seluruh transaksi di-rollback. Dua lapis itu bukan duplikasi yang bisa disatukan: yang pertama memberi pesan yang berguna, yang kedua yang **atomik**. Pemeriksaan di usecase membaca Stok di luar transaksi, jadi nilainya bisa basi saat penulisan; guard SQL itulah yang menjaga Stok tidak pernah negatif. Karena pemeriksaan usecase berjalan lebih dulu, jalur rollback ini tidak bisa dipicu lewat HTTP secara deterministik — karena itu ia diuji di seam adapter (`adapter/sqlite/penjualan_repository_test.go`), bukan hanya di e2e.

### 3. Nomor Struk: `MAX(receipt_number) + 1` di dalam transaksi, bukan tabel counter

Nomor Struk adalah urutan global yang unik, tidak pernah dipakai ulang, dan tidak reset harian (CONTEXT.md). Ia diterbitkan di dalam transaksi checkout yang sama: `SELECT COALESCE(MAX(receipt_number), 0) + 1 FROM penjualan`, dengan `UNIQUE` pada kolomnya sebagai jaring pengaman. Satu toko, satu terminal, satu koneksi tulis (ADR-0002) — jadi baca dan tulis itu tidak bisa berselang dengan checkout lain.

Konsekuensi yang dipakai sebagai sifat: checkout yang ditolak **tidak membakar** Nomor Struk, karena rollback juga membatalkan kenaikan MAX. Tabel counter tersendiri akan butuh transaksi kedua dan tidak memberi apa pun yang belum dijamin di sini.

### 4. Item menyalin nama & harga, dan `product_id` tetap disimpan

`penjualan_item` menyimpan `name` dan `price` hasil salinan saat checkout, supaya rename/reprice Produk tidak menulis ulang riwayat. `product_id` tetap ada supaya barisnya masih menyebut Produk asalnya; foreign key-nya aman karena Produk yang pernah terjual hanya bisa dinonaktifkan (CONTEXT.md, Nonaktif) — dan justru checkout-lah yang menyetel `sold = 1`, di transaksi yang sama. Itu menutup celah yang ditinggalkan #4: sebelum ini hanya tes yang bisa menyetel `sold` lewat SQL langsung.

`cashier_name` juga salinan, karena nama itulah yang tercetak di Struk (#8).

### 5. `created_at` adalah waktu lokal toko, disimpan sebagai teks

`datetime('now','localtime')`, bukan UTC. Satu toko, satu terminal (ADR-0002): laporan omzet harian (#9) dan Struk tercetak (#8) keduanya dibaca dalam waktu lokal toko. Instant UTC akan terbaca sebagai "kemarin" pada Struk yang dicetak pagi hari, dan harus dikonversi balik di dua tempat.

### 6. Nilai `method` berbahasa Inggris; hanya Tunai yang diterima

`cash`/`qris`/`debit`/`transfer` — identifier Go dan nilai DTO JSON mengikuti sisi Inggris (ADR-0012), sementara labelnya tetap "Tunai", "QRIS", "Debit", "Transfer" (CONTEXT.md menaruh "Cash" di daftar `_Avoid_`, dan itu memang aturan glosarium, bukan aturan penamaan kode). Keempatnya sudah dinamai di domain dan di `CHECK` kolomnya sejak migrasi ini, supaya #7 tidak perlu migrasi kedua — tetapi checkout #6 **hanya** menerima Tunai dan menolak tiga lainnya, karena menerimanya berarti mencatat Pembayaran yang belum bisa dipertanggungjawabkan slice ini.

### 7. Kedua Peran boleh menjual; yang dibuka untuk Kasir hanya bacaan katalog

`POST /api/penjualan` ada di belakang cek token, **tanpa** role guard: di toko satu terminal, pemilik yang berperan Admin berdiri di belakang kasir sesering Kasirnya. Yang ikut dibuka adalah `GET /api/produk` — bacaan yang dipakai lookup kasir (`?active=true`) — karena route itu satu-satunya jalan menemukan Produk tanpa endpoint kedua di atas tabel yang sama. Setiap **tulisan** katalog, plus laporan `kategori` dan `stok-menipis`, tetap Admin-only. "Nonaktif hilang dari lookup kasir" ditegakkan oleh filter `active=true` yang dikirim layar kasir, bukan oleh role.

### 8. Route baca `GET /api/penjualan/{nomorStruk}` ada; pembacanya di frontend menyusul

Route itu bagian dari kontrak Penjualan (PRD: "ambil berdasarkan Nomor Struk") dan merupakan satu-satunya cara tes seam REST mengamati bahwa sebuah Penjualan benar-benar **tersimpan**, bukan hanya di-echo oleh handler penulisnya (ADR-0007). Frontend belum punya pembacanya: layar kasir menampilkan Penjualan yang dijawab mutasi checkout, dan pembaca ditambahkan saat #8 (reprint) atau #9 (daftar Penjualan) membutuhkannya — bukan sekarang sebagai kode tanpa pemakai.

### 9. `writeError` kini memisahkan status dari pesan

Sebelum ini, sebuah usecase yang menulis pesan sendiri selalu dijawab 400. Penolakan Stok butuh keduanya: **409** (bertabrakan dengan keadaan Stok saat ini) **dan** pesan yang menyebut Produknya. Jadi `writeError` sekarang membaca pesannya lebih dulu, lalu tetap memilih status dan kode dari sentinel yang di-unwrap error itu (`InputError` → `ErrInvalidInput` → 400; `StockError` → `ErrInsufficientStock` → 409). Perilaku `InputError` yang lama tidak berubah — keduanya tetap dijawab 400 dengan pesannya.

## Considered Options

- **Domain `penjualan` menumpang tabel Produk** — ditolak: Penjualan punya tabel dan aggregate sendiri; menumpang berarti setiap operasinya menyentuh tabel Produk seperti yang ditolak ADR-0014 untuk Stok, tetapi tanpa alasan yang sama.
- **Nomor Struk dari `id` AUTOINCREMENT tabel** — ditolak: menyamakan identitas internal dengan istilah domain (Nomor Struk) membuat keduanya tidak bisa berubah sendiri, dan menyembunyikan aturan "tidak dipakai ulang" di balik perilaku rowid.
- **Pemeriksaan Stok hanya di usecase** (tanpa guard SQL) — ditolak: itu read-modify-write, dan Stok bisa jadi negatif saat dua checkout berdekatan.
- **Pemeriksaan Stok hanya di repository** (tanpa pre-check) — ditolak: pesan penolakannya jadi generik ("salah satu Item"), padahal Kasir butuh tahu Item mana.
- **Menghitung total/Kembalian di repository** — ditolak: itu aturan bisnis di adapter.
- **Menerima keempat metode sekarang** — ditolak: mencatat QRIS/Debit/Transfer berarti mengklaim sesuatu yang belum diuji maupun ditampilkan (#7).
- **`sold` disetel saat Produk pertama kali muncul di keranjang** — ditolak: "pernah terjual" berarti sudah dibayar, dan hanya checkout yang tahu itu.
- **Scan Kode + Enter langsung menambah Item** — tidak diambil. Enter akan menambah berdasarkan hasil pencarian, dan hasil itu bisa belum kembali saat scanner menekan Enter — jadi ia bisa menambah Produk yang salah atau tidak menambah apa pun, tepat pada alur tercepat yang seharusnya paling diandalkan. Yang ada: Kode menyaring daftar, Tambah yang menambahkan.
- **Keranjang disimpan sebagai server state (TanStack Query)** — ditolak: keranjang adalah draf pilihan Kasir, bukan data server (ADR-0006). Stok & harga di barisnya memang berasal dari katalog, tetapi API membacanya ulang saat penjualan ditulis.

## Consequences

- #7 (non-tunai) memperluas `paymentMethod` di usecase, bukan mengubah skema: kolom `method` sudah menerima keempatnya, dan `MetodePembayaranSchema` di frontend juga.
- #8 (Struk) dan #9 (laporan) membangun di atas `GET /api/penjualan/{nomorStruk}` dan `FindByReceiptNumber`. Keduanya sudah mendarat dan pembacanya ada: layar `/penjualan` (#29) untuk lookup dan reprint, layar `/laporan` (#9) untuk omzet harian dan daftar Penjualan — lihat ADR-0018.
- Ambang Stok menipis tetap konstanta domain (ADR-0014); Penjualan yang menurunkan Stok membuat daftar itu berubah, jadi mutasi checkout meng-invalidate seluruh subtree cache `produk`.
- Tes yang menjaga keputusan ini: `backend/internal/usecase/penjualan/checkout_test.go` (aturan: merge baris, blokir Stok, Kembalian, Tunai saja), `backend/internal/adapter/sqlite/penjualan_repository_test.go` (satu transaksi, rollback, Nomor Struk, `sold`), `backend/tests/e2e/penjualan_test.go` (seam REST: 401, atomicity, blokir Stok, kembalian, keunikan Nomor Struk, 409 hapus Produk terjual), `frontend/src/lib/domains/penjualan/{schemas,api,state,components}/*.test.ts`, dan `frontend/tests/e2e/penjualan.spec.ts` (browser sungguhan: jual, cari via Kode & nama, blokir, Kembalian, Nomor Struk, peran, 409).
