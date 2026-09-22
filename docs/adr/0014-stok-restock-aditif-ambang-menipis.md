# Stok: field domain Produk, restock aditif, ambang menipis sebagai konstanta

**Stok** sudah ada sejak #4 sebagai field Produk (di-set saat Produk dibuat), tetapi belum ada jalan menambahnya setelah itu. #5 menambahkan **penambahan Stok manual** dan **daftar Stok menipis**. Tiga keputusan yang diambil di situ dicatat di sini, karena ketiganya mudah dipertanyakan lagi nanti.

## Keputusan

### 1. Stok tetap di domain Produk, bukan domain sendiri

| Lapisan | Nama | Contoh |
| --- | --- | --- |
| Glosarium, label UI | Indonesia | `CONTEXT.md`, "Stok", "Tambah Stok", "Stok menipis" |
| Route | Indonesia | `/stok` (layar Admin), `/api/produk/{id}/stok`, `/api/produk/stok-menipis` |
| Domain frontend | Indonesia, di dalam slice `produk` | `StokList.svelte`, `TambahStokInputSchema`, `StokMenipisSchema`, `createAddStokMutation` |
| Identifier Go | Inggris, di dalam package `produk` | `usecaseproduk.Service.AddStock`, `ProductRepository.AddStock`, `domainproduk.LowStockThreshold` |
| DTO JSON | Inggris | `{"data":{"product":{"stock":13}}}`, `{"quantity":3}`, `{"threshold":5,"products":[…]}` |

Stok memang istilah tersendiri di `CONTEXT.md`, tetapi **bukan** vertical slice tersendiri: setiap operasinya menyebut satu Produk, dan Stok disimpan sebagai kolom di tabel `produk`. Domain `stok` terpisah hanya bisa bekerja dengan membaca tabel Produk — persis "reach into another domain's repository" yang dilarang ADR-0004/ADR-0006 — atau dengan port duplikat di atas tabel yang sama. Sisi frontend mengikuti sisi backend, karena schema zod memang mirror DTO Go 1:1 (ADR-0006): backend tidak punya package `stok`, jadi frontend tidak punya `lib/domains/stok`. Layar `/stok` tetap ada sebagai **route**, dan route memang hanya adapter (ADR-0006).

### 2. Restock itu aditif; ia tidak pernah mengirim total baru

`POST /api/produk/{id}/stok` dengan body `{"quantity": 3}` **menambahkan** 3 ke Stok yang tersimpan, dalam satu statement SQL (`stock = stock + ?`, dengan `RETURNING` untuk menjawab Stok yang baru). Body-nya membawa *yang datang*, bukan *jumlah akhirnya*.

- Kirim-total adalah read-modify-write: dua delivery yang berdekatan bisa saling menimpa, dan yang hilang tidak terlihat di mana pun.
- Hanya `quantity > 0` diterima. Stok keluar dari katalog lewat **Penjualan**, yaitu transaksi yang wajib menjaga Stok tidak negatif (#6) — bukan lewat form Admin. Karena itu tidak ada "kurangi Stok manual".
- Endpoint ini Admin-only seperti sisa katalog.

**Tegangan yang dibiarkan terbuka:** `PUT /api/produk/{id}` masih mengganti Stok sebagai bagian dari record yang bisa diedit (warisan #4, dan form katalog memang menampilkan field Stok). Jadi dua jalur menyentuh angka yang sama dengan semantik berbeda: `PUT` menetapkan, `POST …/stok` menambah. Untuk satu toko satu terminal (ADR-0002) dengan satu Admin yang mengerjakan keduanya, ini diterima: `PUT` adalah jalur **koreksi** salah ketik, `POST …/stok` adalah jalur **barang masuk**. Yang tidak boleh terjadi adalah menambah field Stok ke form restock — itu yang akan mengubah jalur aditif menjadi jalur penetapan. Kalau nanti form katalog perlu berhenti menyentuh Stok, itu perubahan yang disengaja pada #4, bukan efek samping #5.

### 3. "Menipis" adalah konstanta domain, dan aturannya ikut dijawab

`domainproduk.LowStockThreshold` = 5. Produk dianggap **menipis** bila Stoknya **di bawah** ambang itu; Stok 0 berarti **habis** dan tetap bagian dari daftar yang sama. Ambang adalah Stok pertama yang *masih cukup*, jadi Produk yang duduk tepat di angka 5 belum menipis — persis kalimat issue #5, "di bawah ambang atau nol". Daftarnya hanya memuat Produk **Aktif** — Produk Nonaktif tidak sedang dijual, jadi Stoknya tidak bisa "habis" dalam arti yang penting (CONTEXT.md, Nonaktif). Urutannya dari yang paling sedikit, supaya yang harus segera ditambah terbaca lebih dulu.

Ambangnya konstanta, bukan pengaturan per toko, karena belum ada tempat menyimpannya: layar **Pengaturan** mendarat bersama template Struk (#8), dan ambang ini pindah ke sana saat itu terjadi. Sementara itu `GET /api/produk/stok-menipis` menjawab ambangnya bersama daftarnya (`{"threshold":5,"products":[…]}`), supaya layar bisa menuliskan aturannya ("Stok di bawah 5") tanpa menyimpan salinan kedua yang bisa menyimpang dari daftar di bawahnya.

## Considered Options

- **Domain `stok` tersendiri (backend + frontend)** — ditolak: seluruh isinya membaca tabel Produk; yang didapat hanya satu lapisan tambahan dan satu akses lintas-domain.
- **Restock sebagai field di `PUT /api/produk/{id}`** — ditolak: form yang mengirim total akan menimpa delivery yang belum sempat dibaca, dan tidak ada cara membedakan "menambah 3" dari "sekarang 13".
- **`?stokMenipis=true` sebagai filter `GET /api/produk`** — ditolak: ambangnya bukan filter yang dipilih Admin melainkan aturan domain, dan filter tidak punya tempat mengembalikan aturan itu ke UI.
- **Ambang menipis disimpan di tabel pengaturan sekarang** — ditolak: satu baris konfigurasi tanpa satu pun layar yang mengubahnya adalah tempat sampah yang menunggu #8.
- **Mengurangi Stok lewat form Admin** — ditolak: Stok berkurang karena Produk terjual, dan pengurangan itu milik transaksi checkout yang harus menjaga Stok tidak negatif (#6).

## Consequences

- Daftar Stok menipis memakai ambang yang sama untuk seluruh toko; kalau nanti perlu per Produk atau per toko, konstanta ini yang berubah lebih dulu.
- `LowStockThreshold` yang berubah **tidak** memerlukan perubahan frontend: UI menampilkan angka yang dijawab API, bukan angka sendiri.
- Tes yang menjaga keputusan ini: `backend/internal/usecase/produk/stok_test.go` (aturan aditif + isi daftar), `backend/internal/adapter/sqlite/produk_repository_test.go` (satu statement `stock = stock + ?`), `backend/tests/e2e/stok_test.go` (seam REST: 401/403, penambahan, penolakan, ambang), `frontend/src/lib/domains/produk/components/StokList.svelte.test.ts`, dan `frontend/tests/e2e/stok.spec.ts`.
