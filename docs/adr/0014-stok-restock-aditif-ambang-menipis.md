# Stok: field domain Produk, restock aditif, ambang menipis sebagai konstanta

**Stok** sudah ada sejak #4 sebagai field Produk (di-set saat Produk dibuat), tetapi belum ada jalan menambahnya setelah itu. #5 menambahkan **penambahan Stok manual** dan **daftar Stok menipis**. Tiga keputusan yang diambil di situ dicatat di sini, karena ketiganya mudah dipertanyakan lagi nanti.

## Keputusan

### 1. Stok tetap di domain Produk, bukan domain sendiri

| Lapisan             | Nama                               | Contoh                                                                                           |
| ------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------ |
| Glosarium, label UI | Indonesia                          | `CONTEXT.md`, "Stok", "Tambah Stok", "Stok menipis"                                              |
| Route               | Indonesia                          | `/stok` (layar Admin), `/api/produk/{id}/stok`, `/api/produk/stok-menipis`                       |
| Domain frontend     | Indonesia, di dalam slice `produk` | `StokList.svelte`, `TambahStokInputSchema`, `StokMenipisSchema`, `createAddStokMutation`         |
| Identifier Go       | Inggris, di dalam package `produk` | `usecaseproduk.Service.AddStock`, `ProductRepository.AddStock`, `domainproduk.LowStockThreshold` |
| DTO JSON            | Inggris                            | `{"data":{"product":{"stock":13}}}`, `{"quantity":3}`, `{"threshold":5,"products":[…]}`          |

Stok memang istilah tersendiri di `CONTEXT.md`, tetapi **bukan** vertical slice tersendiri: setiap operasinya menyebut satu Produk, dan Stok disimpan sebagai kolom di tabel `produk`. Domain `stok` terpisah hanya bisa bekerja dengan membaca tabel Produk — persis "reach into another domain's repository" yang dilarang ADR-0004/ADR-0006 — atau dengan port duplikat di atas tabel yang sama. Sisi frontend mengikuti sisi backend, karena schema zod memang mirror DTO Go 1:1 (ADR-0006): backend tidak punya package `stok`, jadi frontend tidak punya `lib/domains/stok`. Layar `/stok` tetap ada sebagai **route**, dan route memang hanya adapter (ADR-0006).

### 2. Restock itu aditif; ia tidak pernah mengirim total baru

`POST /api/produk/{id}/stok` dengan body `{"quantity": 3}` **menambahkan** 3 ke Stok yang tersimpan, dalam satu statement SQL (`stock = stock + ?`, dengan `RETURNING` untuk menjawab Stok yang baru). Body-nya membawa _yang datang_, bukan _jumlah akhirnya_.

- Kirim-total adalah read-modify-write: dua delivery yang berdekatan bisa saling menimpa, dan yang hilang tidak terlihat di mana pun.
- Hanya `quantity > 0` diterima. Stok keluar dari katalog lewat **Penjualan**, yaitu transaksi yang wajib menjaga Stok tidak negatif (#6) — bukan lewat form Admin. Karena itu tidak ada "kurangi Stok manual".
- Endpoint ini Admin-only seperti sisa katalog.

**Tegangan ini ditutup oleh #22.** Sampai #5 ada dua jalur menyentuh angka yang sama dengan semantik berbeda: `POST …/stok` **menambah**, sementara `PUT /api/produk/{id}` **menetapkan** Stok sebagai bagian dari record yang bisa diedit (warisan #4, dan form katalog memang menampilkan field Stok). Yang tidak boleh terjadi adalah form yang mengirim _total_ — dan `PUT` memang begitu: Admin yang membuka form edit, lalu menekan Simpan, menulis balik Stok yang sudah usang, menimpa delivery yang belum sempat dibacanya.

#22 mencabut Stok dari jalur edit, dan membuat pencabutan itu **struktural**, bukan aturan yang harus diingat:

| Lapisan         | Bentuknya                                                                                                                          |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Domain Go       | `domainproduk.ProductEdit` hanya memuat `Name`, `Code`, `Price`, `Category` — tidak ada `Stock` untuk dibawa                       |
| Port repository | `Update(ctx, id, edit)`; statement-nya `SET name = ?, code = ?, price = ?, category = ?` — kolom `stock` tidak disebut sama sekali |
| Usecase         | `UpdateInput` tanpa `Stock`, di samping `CreateInput` yang punya (Stok awal)                                                       |
| DTO JSON        | `updateProductRequest` tanpa `stock`, terpisah dari `createProductRequest`                                                         |
| Frontend        | `UpdateProdukInputSchema = CreateProdukInputSchema.omit({ stock, active })`; form edit tidak merender field Stok                   |

Stok karena itu bergerak lewat tepat tiga jalur: **di-set saat create** (Stok awal), **ditambah** `POST /api/produk/{id}/stok`, **dikurangi** Penjualan (#6).

Dua alternatif ditolak, dan alasannya berbeda:

- **Biarkan `PUT` menerima `stock` lalu mengabaikannya di usecase.** Field yang diabaikan API adalah field yang terus dikirim form; yang dibutuhkan bukan penjaga di satu usecase, melainkan bentuk data yang tidak punya tempat untuk Stok. Penjaga di usecase juga bisa dilewati pemanggil berikutnya dari port yang sama.
- **Pertahankan `PUT` sebagai jalur koreksi salah ketik.** Itu memang peran yang diberikan #5 kepadanya, tetapi PRD tidak pernah memberi Admin field Stok di form edit (story 7 menyebut nama, harga, Kode, Kategori), dan koreksi ke bawah tidak punya arti selama Stok hanya naik lewat delivery dan turun lewat Penjualan.

### 3. "Menipis" adalah konstanta domain, dan aturannya ikut dijawab

`domainproduk.LowStockThreshold` = 5. Produk dianggap **menipis** bila Stoknya **di bawah** ambang itu; Stok 0 berarti **habis** dan tetap bagian dari daftar yang sama. Ambang adalah Stok pertama yang _masih cukup_, jadi Produk yang duduk tepat di angka 5 belum menipis — persis kalimat issue #5, "di bawah ambang atau nol". Daftarnya hanya memuat Produk **Aktif** — Produk Nonaktif tidak sedang dijual, jadi Stoknya tidak bisa "habis" dalam arti yang penting (CONTEXT.md, Nonaktif). Urutannya dari yang paling sedikit, supaya yang harus segera ditambah terbaca lebih dulu.

Ambangnya konstanta, bukan pengaturan per toko, karena belum ada tempat menyimpannya: layar **Pengaturan** mendarat bersama template Struk (#8), dan ambang ini pindah ke sana saat itu terjadi. Sementara itu `GET /api/produk/stok-menipis` menjawab ambangnya bersama daftarnya (`{"threshold":5,"products":[…]}`), supaya layar bisa menuliskan aturannya ("Stok di bawah 5") tanpa menyimpan salinan kedua yang bisa menyimpang dari daftar di bawahnya.

## Considered Options

- **Domain `stok` tersendiri (backend + frontend)** — ditolak: seluruh isinya membaca tabel Produk; yang didapat hanya satu lapisan tambahan dan satu akses lintas-domain.
- **Restock sebagai field di `PUT /api/produk/{id}`** — ditolak: form yang mengirim total akan menimpa delivery yang belum sempat dibaca, dan tidak ada cara membedakan "menambah 3" dari "sekarang 13". #22 menutup jalan ini sepenuhnya: `PUT` tidak punya field Stok lagi.
- **`?stokMenipis=true` sebagai filter `GET /api/produk`** — ditolak: ambangnya bukan filter yang dipilih Admin melainkan aturan domain, dan filter tidak punya tempat mengembalikan aturan itu ke UI.
- **Ambang menipis disimpan di tabel pengaturan sekarang** — ditolak: satu baris konfigurasi tanpa satu pun layar yang mengubahnya adalah tempat sampah yang menunggu #8.
- **Mengurangi Stok lewat form Admin** — ditolak: Stok berkurang karena Produk terjual, dan pengurangan itu milik transaksi checkout yang harus menjaga Stok tidak negatif (#6).

## Consequences

- Daftar Stok menipis memakai ambang yang sama untuk seluruh toko; kalau nanti perlu per Produk atau per toko, konstanta ini yang berubah lebih dulu.
- `LowStockThreshold` yang berubah **tidak** memerlukan perubahan frontend: UI menampilkan angka yang dijawab API, bukan angka sendiri.
- **Tidak ada lagi jalur koreksi salah ketik Stok.** Admin yang salah memasukkan Stok awal hanya bisa menaikkannya lewat penambahan manual; menurunkannya tidak punya jalur, karena Stok turun hanya lewat Penjualan (#6). Ini konsekuensi yang diterima, bukan yang terlewat: PRD memang tidak pernah memberi Admin field Stok di form edit. Kalau koreksi ke bawah dibutuhkan nanti, itu keputusan baru — stok opname — bukan efek samping dari ADR ini.
- Tes yang menjaga keputusan ini: `backend/internal/usecase/produk/stok_test.go` (aturan aditif + isi daftar), `backend/internal/usecase/produk/catalog_test.go` (`Update` mempertahankan Stok tersimpan), `backend/internal/adapter/sqlite/produk_repository_test.go` (satu statement `stock = stock + ?` untuk restock; `stock` tidak disebut statement `Update`), `backend/tests/e2e/stok_test.go` (seam REST: 401/403, penambahan, penolakan, ambang), `backend/tests/e2e/produk_test.go` (`PUT` dengan Stok usang tidak mengubah Stok), `frontend/src/lib/domains/produk/components/StokList.svelte.test.ts`, `frontend/src/lib/domains/produk/schemas/produk.schema.test.ts`, dan `frontend/tests/e2e/{stok,produk}.spec.ts`.
