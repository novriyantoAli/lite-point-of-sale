# Istilah domain, route, dan tipe frontend memakai istilah Indonesia; identifier Go dan DTO JSON memakai istilah Inggris

ADR-0011 menetapkan batas ini untuk **Pengguna** saja dan menutup dengan sengaja: "ADR ini **tidak** menetapkan aturan umum untuk domain berikutnya. Apakah Produk nanti bernama `Produk` atau `Product` di Go masih terbuka, dan dijawab saat domain itu mendarat (#4)." Domain Produk sudah mendarat dan memilih sisi yang sama, jadi aturannya diangkat menjadi aturan umum di sini: **glosarium `CONTEXT.md`, route, dan tipe domain frontend memakai istilah Indonesia; identifier Go dan DTO JSON memakai istilah Inggris.** Yang terlihat campur aduk itu memang batas, bukan kelalaian.

## Di mana batasnya

| Lapisan | Nama | Contoh |
| --- | --- | --- |
| Glosarium, label UI | Indonesia | `CONTEXT.md`, "Produk", "Kode", "Kategori", "Nonaktif" |
| Route | Indonesia | `/produk`, `/api/produk`, `/api/produk/kategori` |
| Domain frontend (zod + komponen) | Indonesia | `ProdukSchema`, `ProdukEnvelopeSchema`, `ProdukList.svelte`, `produk.schema.ts` |
| Identifier Go | Inggris | `domainproduk.Product`, `ProductRepository`, `ProductService`, `usecaseproduk.CreateInput`, `usecaseproduk.UpdateInput` |
| DTO JSON | Inggris | `{"data":{"product":{…}}}`, `name`, `code`, `price`, `category`, `stock`, `active`, `sold` |

DTO JSON ikut sisi Go karena schema zod memang mirror DTO Go 1:1 (ADR-0006) — `ProdukEnvelopeSchema` yang menerjemahkannya menjadi `Produk` di sisi domain frontend. Jadi route `/api/produk` yang menjawab `{"data":{"product":…}}` itu disengaja, bukan sisa rename yang tertinggal.

Satu konsekuensi yang mudah salah dibaca: karena identifier Go bukan sinonim domain yang buruk melainkan nama untuk konsep yang sama, ia **tidak** masuk daftar `_Avoid_` di `CONTEXT.md`. `CONTEXT.md` tetap glosarium murni — pemetaannya dicatat di sini, bukan di sana (alasan yang sama seperti ADR-0011).

## Considered Options

- **Semua lapisan memakai istilah Indonesia** — ditolak: `Produk` bukan idiom penamaan Go, dan DTO JSON berbahasa Indonesia akan memaksa frontend menerjemahkan setiap field di `api/`, memindahkan pekerjaan yang sekarang tidak ada ke tempat yang paling sering disentuh.
- **Semua lapisan memakai istilah Inggris** — ditolak: glosarium dan seluruh label UI berbahasa Indonesia, dan `CONTEXT.md` adalah sumber kebenaran istilah yang dibaca manusia, bukan kode.
- **Dua nama tanpa catatan** — ditolak: inilah kondisi yang membuat review menandai `Pengguna`/`User` sebagai drift dan *Mysterious Name* (ADR-0011).
- **Aturan umum ditunda lagi sampai domain ketiga** — ditolak: dua domain yang sepakat sudah cukup untuk menetapkan aturan, dan menundanya berarti domain ketiga menebak-nebak.

## Consequences

- Domain berikutnya (Penjualan, Pembayaran, Struk) mengikuti batas ini tanpa perlu ADR baru: istilah Indonesia di glosarium/route/tipe frontend, Inggris di identifier Go dan DTO JSON.
- Yang tetap pelanggaran: istilah Inggris di label UI, di route, atau di nama tipe domain frontend; dan istilah Indonesia di identifier Go atau field DTO JSON.
- Perubahan nama di satu sisi **bukan** alasan untuk me-rename sisi lain. Kalau nanti diputuskan satu sisi menyerah dan mengikuti sisi lain, ADR inilah yang direvisi lebih dulu, bukan diam-diam di-rename.
- ADR-0011 tetap berlaku untuk detail spesifik Pengguna (termasuk daftar contoh identifier-nya) dan tidak perlu direvisi; ADR ini yang mengangkatnya menjadi aturan umum.
