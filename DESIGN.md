---
name: Lite Point of Sale
description: Kasir mosaik berkepadatan tinggi — satu tinta di atas putih, garis rambut, satu merah utilitas.
colors:
  ground: '#fafafa'
  tile: '#ffffff'
  hairline: '#e8e8e8'
  wash: '#f5f5f5'
  ink: '#000000'
  utility: '#cc0d0d'
typography:
  title:
    fontFamily: 'Inter Variable, system-ui, sans-serif'
    fontSize: '1.5rem'
    fontWeight: 700
    lineHeight: '1'
    letterSpacing: '-0.015em'
  figure:
    fontFamily: 'Inter Variable, system-ui, sans-serif'
    fontSize: '1.25rem'
    fontWeight: 700
    lineHeight: '1.1'
  strong:
    fontFamily: 'Inter Variable, system-ui, sans-serif'
    fontSize: '0.9375rem'
    fontWeight: 600
    lineHeight: '1.2'
  body:
    fontFamily: 'Inter Variable, system-ui, sans-serif'
    fontSize: '0.8125rem'
    fontWeight: 400
    lineHeight: '1.25'
  micro:
    fontFamily: 'Inter Variable, system-ui, sans-serif'
    fontSize: '0.75rem'
    fontWeight: 600
    lineHeight: '1.2'
rounded:
  none: '0px'
spacing:
  pad-y: '6px'
  pad-x: '8px'
  gap-tight: '4px'
  gap: '6px'
  gap-loose: '8px'
  board-pad: '16px'
components:
  module:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
    rounded: '{rounded.none}'
    padding: '6px 8px'
  module-head:
    backgroundColor: '{colors.wash}'
    textColor: '{colors.ink}'
    typography: '{typography.body}'
    padding: '6px 8px'
  tile:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
    typography: '{typography.body}'
    rounded: '{rounded.none}'
    padding: '6px 8px 8px'
  tile-hover:
    backgroundColor: '{colors.wash}'
  tile-disabled:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
  tag:
    backgroundColor: '{colors.utility}'
    textColor: '{colors.tile}'
    typography: '{typography.micro}'
    height: '15px'
    padding: '0 5px'
  tag-quiet:
    backgroundColor: '{colors.wash}'
    textColor: '{colors.ink}'
    typography: '{typography.micro}'
    height: '15px'
  tab:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
    typography: '{typography.body}'
    padding: '6px 10px'
  tab-active:
    backgroundColor: '{colors.utility}'
    textColor: '{colors.tile}'
    typography: '{typography.body}'
  input:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
    typography: '{typography.body}'
    rounded: '{rounded.none}'
    height: '26px'
    padding: '0 6px'
  method:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
    typography: '{typography.body}'
    height: '26px'
  method-checked:
    backgroundColor: '{colors.ink}'
    textColor: '{colors.tile}'
    typography: '{typography.body}'
    height: '26px'
  button-commit:
    backgroundColor: '{colors.ink}'
    textColor: '{colors.tile}'
    typography: '{typography.strong}'
    rounded: '{rounded.none}'
    height: '40px'
    width: '100%'
  button-commit-disabled:
    backgroundColor: '{colors.tile}'
    textColor: '{colors.ink}'
    height: '40px'
  figure:
    textColor: '{colors.ink}'
    typography: '{typography.figure}'
  price:
    textColor: '{colors.utility}'
    typography: '{typography.body}'
---

# Design System: Lite Point of Sale

## Overview

**Creative North Star: "The Dense Mosaic"**

Seluruh katalog, keranjang, dan pembayaran berada di depan mata sekaligus. Layar ini tidak
punya menu yang menyembunyikan apa pun, tidak punya kartu KPI, tidak punya sidebar, dan
tidak punya satu pun bayangan: yang ada hanya petak-petak berbingkai rambut yang dipadatkan
tepi ke tepi sampai satu layar penuh. Yang memisahkan satu petak dari yang lain adalah satu
garis 1px, dan garis itu dipakai bersama oleh kedua petaknya, bukan digambar dua kali.

Sistem ini dibangun dari pengukuran, bukan dari selera. Dunianya adalah tradisi web
berkepadatan tinggi Jepang, dan tokennya diambil dari QUALITY BAR yang diukur piksel per
piksel: dasar `#fafafa`, petak `#ffffff`, garis `#e8e8e8`, tinta `#000000`, dan satu merah
utilitas `#cc0d0d` yang pada bar hanya menempati 3,03% permukaan. Warnanya sengaja dingin
netral. Dasar krem hangat adalah penampilan yang dikirim hampir semua model untuk subjek
apa pun, dan sistem ini tidak mendarat di sana.

Aturan yang paling menentukan bukan warnanya, melainkan yang tidak ada di dalamnya: **tidak
ada abu-abu sebagai tinta huruf.** Di bar-nya, semua abu-abu adalah garis dan isian, dan
teksnya selalu hitam. Itu yang membuat seluruh teks di layar ini lolos WCAG AA, dan itu yang
memperbaiki cacat terukur di sistem sebelumnya (tint merah 3,97:1 dan keadaan mati 3,23:1
karena `opacity: 0.5`). Hierarki di sini dibawa kontras skala dan kepadatan, bukan warna.

**Key Characteristics:**

- **Enam token, satu di antaranya merah.** Dasar, petak, garis, isian, tinta, dan satu merah
  utilitas. Tidak ada gradien, tidak ada tint, tidak ada keluarga warna kedua.
- **Merah hanya di tab dan harga, dan tetap di bawah 3% permukaan.** Terukur 1,39% pada
  layar Kasir.
- **Radius nol dan bayangan nol di mana-mana.** Kedalaman dibawa kisi berbingkai rambut.
- **Lima ukuran huruf dengan lompatan skala nyata:** 24 / 20 / 15 / 13 / 12px.
- **Lantai huruf 12px.** Ukuran kecil milik bar ditolak; kepadatan dibayar dengan kisi.
- **Satu layar tanpa menggulir.** 24 Produk, keranjang, dan pembayaran muat dalam 1440×900.
- **Keadaan adalah tanda, bukan warna pudar.** Tombol mati kehilangan tintanya; Produk habis
  dicoret garis, bukan dipudarkan.
- **Tidak ada yang bersembunyi di balik hover.** Setiap kontrol membawa katanya sendiri.

## Colors

Palet ini enam nilai, dan lima di antaranya tidak berwarna. Warna di sini bekerja sebagai
terang-gelap dan sebagai struktur, bukan sebagai identitas. Seluruh nilainya diukur dari
QUALITY BAR board dan hero (2048×1152) sebelum satu baris kode ditulis.

### Primary

- **Ink Black** (`#000000`): hitam murni, bukan near-black. Seluruh teks, seluruh garis
  kepala, isi metode pembayaran yang terpilih, dan satu-satunya bidang bertinta penuh di
  layar: tombol Bayar & Simpan Penjualan.
- **Tile White** (`#ffffff`): permukaan setiap petak, modul, dan input. Petak sengaja lebih
  putih daripada halaman; itu yang membuat mosaiknya terbaca sebagai petak, bukan sebagai
  halaman.

### Secondary

- **Utility Red** (`#cc0d0d`): rata-rata piksel merah pada bar. Hanya dua pekerjaan: tab
  (Kode Produk di setiap petak, dan tab navigasi yang aktif) dan harga. Tidak pernah untuk
  hiasan, tidak pernah untuk menandai keadaan yang tenang, dan tidak pernah dipakai sebagai
  tint.

### Neutral

- **Ground Grey** (`#fafafa`): dasar halaman. Dingin netral, tanpa kroma, dan sengaja bukan
  krem hangat.
- **Hairline Grey** (`#e8e8e8`): setiap garis 1px di sistem ini, dan tidak pernah jadi tinta
  huruf.
- **Wash Grey** (`#f5f5f5`): isian hover, kepala modul, dan latar tag yang tenang.

### Named Rules

**The Zero-Grey Rule.** Abu-abu hanya boleh jadi garis dan isian; ia tidak pernah jadi tinta
huruf. Setiap teks memakai Ink Black. Inilah sebabnya seluruh teks di layar ini lolos AA, dan
ini pula yang membuat Quiet Ink di sistem sebelumnya tidak diwariskan.

**The Three-Percent Rule.** Merah utilitas menempati kurang dari 3% permukaan pada setiap
layar, dan hanya muncul sebagai tab atau harga. Ukur, jangan kira-kira; pada layar Kasir
angkanya 1,39%.

**The No-Tint Rule.** Tidak ada nilai warna yang dipakai sebagai tint 10% atau 20%. Keadaan
yang butuh pembeda memakai bidang penuh, garis rambut, atau tanda; bukan warna yang
dilembutkan. Cacat kontras sistem sebelumnya datang tepat dari tint, bukan dari warnanya.

## Typography

**Display Font:** tidak ada. Ukuran terbesar di sistem ini adalah 24px.
**Body Font:** Inter Variable, disematkan sendiri (`@fontsource-variable/inter`).
**Label/Mono Font:** sama dengan body. **Tidak ada monospace** di sistem ini, termasuk pada
angka.

**Character:** satu huruf kerja untuk permukaan Operate, dengan karakter padat yang diambil
dari leading dan kisi, bukan dari jenis hurufnya. Dunia asalnya memakai gothic padat berukuran
kecil; sistem ini mengambil kepadatannya dan **menolak ukuran kecilnya**, karena pemakainya
belum pernah memakai sistem kasir dan harus bisa membacanya.

### Hierarchy

- **Title** (700, 1,5rem/1, `-0.015em`): satu per layar, dan satu-satunya tempat ukuran ini
  muncul. Di layar Kasir ia adalah strip judul di atas papan.
- **Figure** (700, 1,25rem/1,1): angka yang jadi jawaban layar — Kembalian, jumlah dibayar.
  Selalu `tabular-nums`.
- **Strong** (600, 0,9375rem/1,2): total berjalan, nilai di kepala kolom, teks tombol utama.
- **Body** (400, 0,8125rem/1,25): badan seluruh antarmuka — nama Produk, isi tabel, input,
  teks tombol. Ukuran kerja sistem ini.
- **Micro** (600, 0,75rem/1,2): tab, badge, kepala tabel, label field. **Ini lantainya**;
  tidak ada teks di bawah 12px.

Semua angka uang, jumlah, dan Stok memakai `font-variant-numeric: tabular-nums`, tanpa
kecuali.

### Named Rules

**The Scale-Contrast Rule.** Hierarki dibawa kontras skala dan kepadatan, bukan warna dan bukan
bobot sendirian. Lompatan terbesarnya 24 → 13px, dan itu satu-satunya lompatan besar yang
dibutuhkan sebuah layar.

**The Legibility Floor Rule.** Tidak ada teks di bawah 12px, dan tidak ada teks abu-abu.
Dunia asalnya memakai huruf kecil untuk memadatkan layar; sistem ini menolaknya dan membayar
kepadatan dengan kisi serta hilangnya ruang kosong, bukan dengan huruf yang mengecil.

## Layout

Sistem ini bukan kolom tengah yang mengambang. Papannya selebar viewport dengan padding luar
8px, dan seluruh isinya adalah satu mosaik yang memadat tepi ke tepi.

- **Papan:** tiga kolom, `232px` untuk pencarian, `minmax(0, 1fr)` untuk katalog, `300px`
  untuk keranjang dan pembayaran. Semuanya rata atas, dan tiap kolom adalah tumpukan modul
  yang berbagi garis.
- **Rel navigasi:** menempel di tepi atas, dua baris nav di bawah satu baris identitas
  (Beranda/Kasir/Penjualan, lalu Produk sampai Backup). Baris pertama ditutup garis tinta,
  baris berikutnya garis rambut.
- **Strip judul:** selebar papan, judul 24px/700 dengan satu kalimat 12px di sebelahnya.
  Ditarik `-1px` supaya garis bawahnya menutup garis atas kolom.
- **Mosaik katalog:** `repeat(auto-fill, minmax(154px, 1fr))`, celah 0, dan setiap petak
  memakai `margin: -1px 0 0 -1px` supaya garis 1px dipakai bersama dua petak alih-alih
  digambar dua kali.
- **Ritme:** 6px vertikal dan 8px horizontal di dalam modul, 4–8px antar elemen di dalamnya,
  dan nol antar modul. Jarak adalah hal yang pertama dibuang di sini.
- **Layar sempit:** di bawah 1080px papan menumpuk jadi satu kolom, dan di bawah 560px rel
  membungkus serta mosaik turun ke minimum 132px per petak. Laptop di meja kasir adalah
  perangkatnya (PRODUCT.md), jadi ini hanya supaya tidak rusak.

### Named Rules

**The Shared-Hairline Rule.** Dua modul yang bersebelahan berbagi satu garis 1px, bukan dua
garis yang menempel. Kalau sebuah garis bisa dihitung dua kali, ia salah dipasang.

**The One-Screen Rule.** Seluruh katalog, keranjang, dan pembayaran harus muat dalam satu
viewport 1440×900 tanpa menggulir. Pada 24 Produk, tingginya tepat 900px. Kalau sebuah
perubahan mendorongnya melewati itu, yang dikurangi adalah ruang kosong, bukan jumlah Produk
yang terlihat.

## Elevation & Depth

**Tidak ada bayangan sama sekali.** Nol elemen di sistem ini punya `box-shadow`. Yang
memisahkan permukaan adalah garis 1px, dan yang membedakan halaman dari petak adalah selisih
`#fafafa` ke `#ffffff` sebesar satu langkah kecil.

Kedalaman di sini bukan tumpukan, melainkan **satu-satunya bidang bertinta penuh**: tombol
Bayar & Simpan Penjualan. Ia tidak diangkat dengan bayangan, ia diangkat dengan tinta. Itu
membuat aksi utama mustahil terlewat di layar yang seluruhnya putih, tanpa satu pun trik
kedalaman.

### Named Rules

**The No-Shadow Rule.** Nol bayangan, di mana pun. Kalau sesuatu perlu terlihat lebih tinggi,
ia memakai tinta penuh atau garis yang lebih tebal, bukan bayangan. Bayangan di sistem ini
berarti sistemnya sudah berubah.

## Shapes

**Radius nol di mana-mana.** Tidak ada sudut bulat di sistem ini, dan tidak ada `border-radius`
selain 0. Petak, modul, tab, input, tombol, dan tag semuanya persegi.

Garis selalu 1px. Tidak ada garis 2px, tidak ada garis ganda, tidak ada garis berwarna selain
garis tinta di bawah kepala modul dan di atas tombol utama. Tidak ada `clip-path`, tidak ada
tekstur, tidak ada bentuk geometris hiasan.

Tag adalah satu-satunya bentuk yang tingginya ditentukan: 15px, supaya ia terbaca sebagai tab
kecil di kepala petak tanpa menambah tinggi baris.

### Named Rules

**The Square-Corner Rule.** Radius nol. Sudut bulat adalah kosakata sistem lain; di sini ia
akan membuat mosaiknya terbaca sebagai kumpulan kartu, dan kartu adalah hal yang sistem ini
tolak.

## Components

### Modules

Blok berisi terkecil: latar Tile White, satu garis rambut, padding 6px × 8px. **Module head**
memakai latar Wash Grey dan ditutup garis tinta di bawahnya; itu satu-satunya penanda kepala
di sistem ini, dan tidak ada bayangan yang menemaninya.

### Tiles

Petak katalog, dan unit terkecil yang bisa ditekan.

- **Shape:** persegi, radius 0, garis rambut 1px yang dipakai bersama tetangganya.
- **Anatomi:** tab Kode di kiri atas, nama Produk 13px/500, lalu kaki baris berisi Stok di
  kiri dan harga di kanan. Harga selalu merah utilitas dan selalu `tabular-nums`.
- **Hover:** latar jadi Wash Grey dan **garisnya naik jadi tinta**. Tidak ada bayangan, tidak
  ada pergerakan, dan tidak ada aksi sekunder yang muncul.
- **Produk tanpa Kode:** tab-nya jadi tag tenang bertuliskan "tanpa Kode", supaya kolom yang
  kosong tidak pernah terlihat seperti data yang hilang.
- **Stok menipis:** Stok menuliskan "menipis" di sebelah angkanya, dan Produk yang habis
  kehilangan tintanya: nama dan harganya dicoret garis, dan petaknya mati.

### Tags

Tinggi 15px, huruf 12px/600, radius 0. **Red tag** untuk Kode Produk, tab navigasi yang aktif,
dan `tersegel` — Penjualan yang sudah tersimpan dan tidak bisa dibatalkan. **Quiet tag** (Wash
Grey, garis rambut) untuk keadaan yang tenang: tanpa Kode, menipis, dan label Peran. Tidak ada
tag yang menyampaikan keadaan lewat warnanya saja.

### Tabs

Navigasi adalah tab, bukan daftar: setiap tab adalah sel persegi dengan garis rambut di
kanannya. Tab yang aktif memakai isian merah penuh dengan teks putih. Prototipe menaruh cacah
(`Produk 24`) di sebelah nama tab sebagai angka contoh; **belum ada cacah yang dikirim** — angka
aslinya berarti satu permintaan ke API dari kerangka di setiap layar, dan itu belum diputuskan.

### Fields & Inputs

Tinggi 26px, radius 0, latar Tile White, garis rambut 1px, teks 13px. Label selalu di atas
field dengan jarak 3px, dan selalu 12px/600. Fokus memakai garis tinta plus outline 2px ke
dalam, bukan glow. Field yang tidak valid memakai garis dan outline merah utilitas, dengan
pesannya sebagai baris 12px/600 di bawah field.

### Methods

Empat metode Pembayaran sebagai radio asli yang disembunyikan secara visual, dengan label
sebagai sel persegi yang terlihat. Terpilih berarti **bidang tinta penuh dengan teks putih**,
bukan tint dan bukan centang. Cincin fokusnya muncul di sel, bukan di input yang tak terlihat.

### Commit Button

Satu-satunya bidang bertinta penuh di layar: 40px, radius 0, latar Ink Black, teks putih
15px/600, lebar penuh kolom keranjang. Saat mati ia **kehilangan tintanya** dan berubah jadi
putih bergaris putus-putus, bukan jadi pudar.

### Figures

Angka yang jadi jawaban layar: 20px/700, `tabular-nums`, selalu didahului label 12px/600 yang
menyebut apa angka itu. Angka yang belum bisa dihitung ditulis `—`, bukan `0`.

### Browser surfaces

Teks terpilih memakai tinta penuh dengan teks putih. Caret memakai tinta, kecuali di field
jumlah bayar yang memakai merah utilitas. Scrollbar memakai `scrollbar-width: thin` dengan
garis rambut di atas dasar. Cincin fokus selalu outline 2px ke dalam, tidak pernah glow.

### Named Rules

**The State-Is-Not-Faded Rule.** Keadaan mati tidak pernah disampaikan dengan menurunkan
opasitas. Tombol kehilangan tintanya, Produk habis dicoret, field tidak valid diberi garis
merah penuh. Sistem sebelumnya memakai `opacity: 0.5` dan itu terukur 3,23:1, di bawah AA.

**The Named-Not-Hidden Rule.** Tidak ada aksi yang bersembunyi di balik hover. Dunia asalnya
memunculkan aksi sekunder saat hover; sistem ini menolaknya, karena pemakainya belum pernah
memakai sistem kasir dan kosakata tersembunyi adalah kosakata yang harus dipelajari lebih
dulu. Hover hanya mengubah latar dan garis.

## Do's and Don'ts

### Do:

- **Do** keep the ground at `#fafafa`, cool and neutral. The craft bar's ground is measured, not chosen, and warm cream is the rendition this system deliberately did not land on.
- **Do** keep every piece of text Ink Black on a light ground; greys are lines and fills only.
- **Do** keep the utility red under 3% of any screen, spent only on tabs and prices.
- **Do** share one 1px hairline between adjacent modules instead of drawing two.
- **Do** keep radius at 0 and `box-shadow` at none, everywhere.
- **Do** keep `tabular-nums` on every amount, count, and Stok figure.
- **Do** keep the 12px floor and the 24 → 13px scale jump; hierarchy comes from scale contrast and density.
- **Do** keep the whole catalogue, cart, and payment inside one 1440×900 viewport.
- **Do** let a disabled control lose its ink or take a strike, never its opacity.
- **Do** theme the browser surfaces — selection, caret, scrollbar, focus ring — from this palette.
- **Do** write every state's word next to its mark: `menipis`, `Stok habis`, `tanpa Kode`, `tersegel`.

### Don't:

- **Don't** introduce a tint. A 10% red tint measured 3,97:1 in the previous system; full-strength red on white measures 5,88:1. Tints are how this palette loses its contrast.
- **Don't** put grey text anywhere, at any size.
- **Don't** add a shadow, a radius, a gradient, or a second hue. This system's depth is ink, and its shape is the square.
- **Don't** shrink type below 12px to fit more in. The density is the grid's job; the craft bar's tiny type was refused on the owner's legibility constraint.
- **Don't** use monospace, for numbers or anything else. It is not in this system.
- **Don't** hide an action behind hover, a tooltip, or a layer above the page.
- **Don't** mark a quiet state in red. Red means a price or an active tab; `Nonaktif` and `menipis` are quiet tags, and `Stok habis` is a strike.
- **Don't** wrap a screen in a card, and never nest one: a module is a ruled region of the grid, and a card is what this system refuses to look like.
- **Don't** let the mosaic scroll sideways at any width; below 1080px the board stacks.
- **Don't** re-settle the header: the store's own name leads the rail now, and the product's name is only the fallback for a store that has not been named yet (ADR-0019).
