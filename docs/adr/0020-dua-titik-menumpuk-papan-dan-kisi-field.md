# Dua titik menumpuk: papan 1080px, kisi field 900px

`DESIGN.md` hanya menyebut satu titik menumpuk ("di bawah 1080px papan menumpuk"), sementara
prototipe dunia ini punya dua: `.board` — papan Kasir yang tiga kolom — menumpuk di 1080px,
dan `.formgrid` — 2, 3, atau 4 kolom field — menumpuk di 900px, lewat satu
`@media (max-width: 900px)` di `.impeccable/build/sistem/style.css` yang menyentuh
`.formgrid`, `.formgrid--3`, dan `.formgrid--4` sekaligus.

Port #33 membaca prototipe layar Produk sebagai "menumpuk di 1080px" dan menulis kisi
saringannya `min-[1081px]:grid-cols-4`; port #35 menulis kisi Pengguna
`min-[901px]:grid-cols-3`. Playbook port mencatat itu sebagai dua angka yang sama-sama setia
pada prototipenya masing-masing.

Bacaan itu keliru. Prototipe layar Produk memakai `.formgrid--4` dari `sistem/style.css`
(`.impeccable/build/produk/index.html:60`), tanpa aturan per layar dan tanpa `<style>` di
berkasnya, jadi kisi empat kolomnya menumpuk di **900px** seperti `.formgrid--3` milik
Pengguna. Angka 1080 di kisi saringan Produk bukan bacaan yang setia, melainkan titik papan
yang dipakai untuk hal yang bukan papan. Setelah dipisahkan, tidak ada dua angka untuk satu
hal: **kisi field menumpuk di 900, papan di 1080.**

## Keputusan

1. **Kisi field menumpuk di 900px.** Berlaku untuk kisi `.formgrid` mana pun — 2, 3, atau 4
   kolom — baik di dalam modul maupun di dalam dialog. Dialog pun sebuah kisi field: lebarnya
   448px pada layar lebar (`sm:max-w-md`), jadi dua kolomnya memang sudah sempit sebelum 900.
2. **Papan Kasir menumpuk di 1080px.** Ia satu-satunya permukaan tiga kolom tetap
   (`232px | 1fr | 300px`), dan ia menumpuk saat kolomnya sendiri tidak lagi muat.
3. **Rel dan mosaik menumpuk di 560px**, tidak berubah.

Konsekuensinya: `ProdukList` pindah dari `min-[1081px]:grid-cols-4` ke `min-[901px]:grid-cols-4`,
dan `ProdukForm` dari `sm:grid-cols-2` (640px) ke `min-[901px]:grid-cols-2`. `PenggunaList`
sudah 900px; `PengaturanForm` menyusul saat layar Pengaturan diport (#38). Setelah ini,
`grid-cols` berprefiks breakpoint hanya boleh berbunyi `min-[901px]` untuk kisi field dan
`min-[1081px]` untuk papan.

## Considered Options

- **Satu titik untuk semuanya, 1080px.** Paling mudah diingat, dan itu yang #33 sudah pakai.
  Ditolak: kisi field tidak punya alasan menunggu papannya menumpuk, dan aturan itu
  menyalahkan `sistem/style.css` yang justru memisahkan keduanya.
- **Menyimpan `sm:` (640px) untuk dialog.** Dialog punya lebar sendiri, jadi ambang viewport
  terasa tidak nyambung. Ditolak: itu meninggalkan angka ketiga untuk konsep yang sama, dan
  "kapan kisi field menumpuk" kembali jadi pertanyaan yang harus dijawab tiap port.
- **Membiarkan dua angka sebagai keputusan yang disengaja.** Ditolak: satu di antaranya adalah
  salah baca, bukan pilihan.
