# Identifier Go memakai `User`, istilah domain dan UI memakai `Pengguna`

`CONTEXT.md` menetapkan **Pengguna** sebagai istilah domain, tetapi kode memakai dua nama untuk satu konsep: identifier Go (`User`, `PublicUser`, `UserRepository`) berdampingan dengan nama domain (`pengguna.go`, `PenggunaList.svelte`, route `/pengguna`). Review menandainya sebagai drift. Keputusannya: **identifier Go memakai `User`; glosarium, label UI, route, dan tipe domain frontend memakai `Pengguna`.** Yang terlihat campur aduk itu memang batas, bukan kelalaian.

## Di mana batasnya

| Lapisan | Nama | Contoh |
| --- | --- | --- |
| Glosarium, label UI | `Pengguna` | `CONTEXT.md`, "Pengguna", "Nonaktifkan" |
| Route | `Pengguna` | `/pengguna`, `/api/pengguna` |
| Domain frontend (zod + komponen) | `Pengguna` | `PenggunaSchema`, `PenggunaEnvelopeSchema`, `PenggunaList.svelte` |
| Identifier Go | `User` | `domainauth.User`, `PublicUser`, `UserRepository`, `CreateUser` |
| DTO JSON | `User` | `{"data":{"user":{…}}}` |

DTO JSON ikut sisi Go karena schema zod memang mirror DTO Go 1:1 (ADR-0006) — `PenggunaEnvelopeSchema` yang menerjemahkannya menjadi `Pengguna` di sisi domain frontend. Jadi route `/api/pengguna` yang menjawab `{"data":{"user":…}}` itu disengaja, bukan sisa rename yang tertinggal.

## Considered Options

- **Rename total ke `Pengguna`** — ditolak: sekitar 123 perubahan identifier di Go tanpa mengubah perilaku apa pun, dan `Pengguna` bukan idiom penamaan Go.
- **Rename total ke `User`** — ditolak: glosarium dan seluruh label UI berbahasa Indonesia (Produk, Kode, Stok, Struk, Penjualan, Pengguna), sementara `Pengguna` muncul 271 kali di frontend.
- **Dua nama tanpa catatan** (kondisi sebelum ADR ini) — ditolak: itulah yang membuat review menandainya sebagai drift dan *Mysterious Name*.

## Consequences

- Entri **Pengguna** di `CONTEXT.md` tidak lagi memuat `User` di daftar `_Avoid_`: `User` bukan sinonim domain yang buruk, melainkan nama identifier Go untuk konsep yang sama. `CONTEXT.md` tetap glosarium murni — pemetaannya dicatat di sini, bukan di sana.
- Yang tetap pelanggaran: `User` di label UI, di route, atau di nama tipe domain frontend.
- ADR ini **tidak** menetapkan aturan umum untuk domain berikutnya. Apakah Produk nanti bernama `Produk` atau `Product` di Go masih terbuka, dan dijawab saat domain itu mendarat (#4) — bukan disimpulkan dari sini.
- Kalau nanti diputuskan semua domain memakai istilah Indonesia di Go, ADR inilah yang direvisi lebih dulu, bukan diam-diam di-rename.
