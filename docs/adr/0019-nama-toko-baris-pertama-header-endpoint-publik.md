# Nama toko: baris pertama header Struk, endpoint publik

PRODUCT.md berjanji **nama toko harus tampil di layar** (dikonfirmasi pemilik), tetapi schema sengaja tidak punya `store_name` — nama toko hidup sebagai teks bebas di blok `header` Pengaturan. Layar **Masuk** ditampilkan *sebelum* ada sesi, sedangkan `GET /api/pengaturan` ada di belakang role guard Admin. ADR ini memutuskan dari mana layar Masuk membaca nama toko, dan lewat jalur mana.

## Keputusan

### 1. Nama toko adalah baris pertama yang tidak kosong dari blok header

| Lapisan | Nama |
| --- | --- |
| Glosarium, label UI | Indonesia | `CONTEXT.md`, "Nama toko" |
| Aturan derivasi | domain Go | `domainpengaturan.Settings.StoreName()` |
| Route | Inggris (boundary ADR-0012) | `GET /api/store-name` |
| DTO JSON | Inggris | `{"data":{"store_name":"Toko Kopi"}}` |
| Domain frontend | Indonesia, di slice `pengaturan` | `createStoreNameQuery()`, `StoreNameEnvelopeSchema` |

Nama toko **tetap tidak punya field sendiri**: ia adalah baris pertama yang tidak kosong dari `Settings.Header` (baris kosong di depan dilewati, spasi dirapikan). Blok header kosong berarti toko belum punya nama, dan endpoint menjawab string kosong — layar Masuk lalu jatuh ke nama produk. Aturan derivasi ditaruh sebagai method di domain, bukan di adapter HTTP, supaya bisa diuji tanpa database (ADR-0007).

Ditolak: kolom `store_name` baru di schema. Itu berarti migrasi + perubahan kontrak Go ↔ frontend, dan memaksa pemilik menjaga nama di dua tempat (kolom dan blok header) atau mengubah arah ketergantungannya — padahal header Struk sudah memuat nama itu dan dicetak apa adanya.

### 2. Satu endpoint publik yang hanya menjawab nama

`GET /api/store-name` didaftarkan di bagian *public* router (sejajar `GET /api/health` dan `POST /api/auth/login`), tanpa guard token maupun role. Jawabannya **hanya** `store_name`: ia membaca `SettingsService.Get()` yang sama, tetapi tidak pernah mengirim `paper_width` atau `low_stock_threshold` — Pengaturan selebihnya tetap Admin-only di `GET /api/pengaturan`.

Ditolak: membuka `GET /api/pengaturan` untuk publik. Nama memang cuma satu baris, tapi membuka seluruh row berarti membocorkan lebar kertas dan ambang ke layar pra-login tanpa alasan.

### 3. Frontend membaca lewat slice `pengaturan`

`createStoreNameQuery()` tinggal di slice `pengaturan` (namanya berasal dari data Pengaturan), diimpor layar Masuk lewat barrel `$lib/domains/pengaturan`. BFF proxy `api/store-name/+server.ts` meneruskan tanpa token, dan `enabled: browser` sama seperti query domain lain. Layar Masuk menampilkan nama saat tiba, `""` (toko belum bernama) jatuh ke nama produk, dan keadaan memuat tidak menampilkan nama produk.

## Considered Options

- **Kolom `store_name` baru** — eksplisit dan tahan banting, tapi migrasi + kontrak + nama rangkap. Ditolak (keputusan 1).
- **Membuka `GET /api/pengaturan`** — paling sedikit kode, tapi membocorkan setelan Admin ke layar pra-login. Ditolak (keputusan 2).
- **Frontend membaca header penuh** — tidak mungkin: header ada di balik guard yang sama, dan menurunkan guard demi satu baris sama saja dengan opsi di atas.
