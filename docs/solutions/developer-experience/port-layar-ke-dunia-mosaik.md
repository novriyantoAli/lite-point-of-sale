---
title: "Port layar ke dunia mosaik — varian shadcn yang bertabrakan dengan DESIGN.md"
date: 2026-09-26
last_updated: 2026-09-29
category: developer-experience
module: frontend-design
problem_type: developer_experience
component: development_workflow
severity: medium
applies_when:
  - "Memport salah satu dari delapan layar yang belum pindah ke dunia mosaik"
  - "Memakai varian primitif ui/ yang belum dipakai layar Masuk atau rel navigasi"
  - "Ketika DESIGN.md menyebut angka (tinggi, warna, bayangan) dan hasilnya harus dibuktikan"
symptoms:
  - "Tombol mati tetap pudar walau DESIGN.md melarang opasitas sebagai penanda keadaan"
  - "`shadow-none` tidak menghapus bayangan milik `variant=\"outline\"`"
  - "Cincin fokus terukur `2px none`: lebarnya ada, gayanya tidak"
  - "Hasil pengukuran Playwright tidak berubah setelah kode diubah"
root_cause: framework_constraint
resolution_type: workflow_improvement
tags: [frontend, design-system, shadcn-svelte, tailwind-v4, tailwind-merge, accessibility, verification, playwright, focus-ring, css-layers]
---

# Port layar ke dunia mosaik — varian shadcn yang bertabrakan dengan DESIGN.md

## Context

Layar Masuk diport lebih dulu (`e19c085`), lalu rel navigasi dan nama toko. Polanya sudah
tetap: token shadcn di `frontend/src/app.css` dipetakan ke enam token dunia, primitif
`lib/components/ui/` tidak pernah diedit, dan tiap layar menimpa tampilannya lewat prop
`class`. Tiga jebakan muncul di dua port pertama, dan ketiganya tidak terlihat dari membaca
kode — hanya dari mengukur DOM yang berjalan.

Jebakannya seragam: **primitif shadcn membawa keputusan tampilan sendiri, dan sebagian
keputusan itu adalah hal yang DESIGN.md tolak** (opasitas, bayangan, radius). Menyalin kelas
Tailwind ke prop `class` tidak selalu menang, karena `cn()` (tailwind-merge) hanya menyatukan
kelas yang dikenali sebagai grup yang sama.

## Guidance

1. **Jangan menambah palet kedua.** Token shadcn sudah menunjuk ke nilai dunia: `bg-card`
   = petak, `bg-muted` = isian, `border-border` = garis rambut, `border-foreground` = tinta,
   `bg-destructive` = merah utilitas, `text-foreground` = tinta huruf. Ukuran huruf ditulis
   literal (`text-[13px]`, `text-xs`, `text-[15px]`), bukan token baru — itulah yang sudah
   dipilih port pertama, dan dua kosakata di tengah port lebih mahal daripada pengulangan.

2. **Periksa base primitif untuk opasitas, bayangan, dan radius sebelum menimpanya.**
   `Button` membawa `disabled:opacity-50` di base, dan itu tepat keadaan yang DESIGN.md
   tolak (terukur 3,23:1, di bawah AA). Obatnya satu kelas: `disabled:opacity-100`. Karena
   base juga memasang `disabled:pointer-events-none`, tombol yang ingin tetap memperlihatkan
   `cursor: not-allowed` perlu `disabled:pointer-events-auto`.

3. **Kalau sebuah varian membawa hiasan yang dilarang dunia ini, pilih varian yang bersih —
   jangan bertarung lewat kelas.** Tombol `variant="outline"` membawa `shadow-xs`, dan
   `shadow-none` **tidak** mengalahkannya. Yang bekerja: `variant="ghost"` lalu bangun
   tampilannya dari `class`. Sebaliknya, `shadow-none` memang obat yang benar untuk `Input`
   (terukur: seluruh lapisan `box-shadow` menjadi nol), jadi jangan generalisasi "`shadow-none`
   tidak bekerja" — yang gagal adalah menimpanya pada varian yang sama-sama mendeklarasikan
   bayangan.

   Koreksi, 2026-09-26: catatan sebelumnya di sini menyalahkan `shadow-none` secara umum. Yang
   diukur ulang adalah `shadow-[none]` (nilai arbitrer), dan itulah yang kalah; `shadow-none`
   biasa menang.

4. **Keadaan tab aktif dipasang lewat `aria-current`, bukan kelas kondisional.** `class:`
   tidak bisa mencocokkan apa pun dari `aria-current`; yang bisa adalah varian Tailwind
   `aria-[current=page]:`. Dua hal yang mudah terlewat: hover pada tab aktif perlu variannya
   sendiri (`aria-[current=page]:hover:bg-destructive`), dan variannya harus **diverifikasi
   ada di CSS hasil build** — `grep -o 'aria-current[^{]*{[^}]*}' .svelte-kit/output/client/_app/immutable/assets/*.css`.

5. **Satu kelas yang dipakai sembilan kali ditulis sekali.** Rel navigasi punya sembilan tab;
   String kelas yang sama disalin sembilan kali akan berbeda sendiri dalam dua port. Simpan
   sebagai satu `const TAB` — pemindai Tailwind membaca teks berkas, jadi kelasnya tetap
   ikut ter-generate.

6. **Verifikasi port dengan mengukur, bukan dengan melihat.** DESIGN.md sendiri berbunyi
   "Ukur, jangan kira-kira", dan pengukuran juga satu-satunya cara memverifikasi port tanpa
   mata. Pola yang dipakai: spec Playwright sekali pakai yang menyalin `getComputedStyle`
   dari DOM aplikasi **yang berjalan** pada 1440×900 (bukan prototipe `file://` seperti
   `.impeccable/tools/audit-kasir.mjs`), lalu rasternya disimpan ke
   `.impeccable/preview/shots/` seperti `rail-sveltekit.png`.

7. **Aturan tingkat dunia tidak boleh ditulis di dalam `@layer`.** Tailwind menaruh
   `utilities` sesudah `base`, dan urutan layer mengalahkan spesifisitas: apa pun yang
   ditulis di `@layer base` — termasuk `input:focus-visible` (0,1,1) — kalah dari `.outline-none`
   milik primitif shadcn (0,1,0). Inilah kenapa cincin fokus dunia ini sempat terukur
   `2px none`: lebarnya dipasang, gayanya tidak. Cincin fokus yang benar (garis tinta 2px ke
   dalam) ditulis **di luar layer mana pun** di `app.css`, sekali untuk seluruh aplikasi.

8. **Mematikan glow bawaan tanpa `!important`: jadikan `--ring` transparan.** Primitif shadcn
   membawa `focus-visible:ring-3`, dan cincin itu digambar dari `--tw-ring-*`. Menyetel
   `--ring: transparent` membuat seluruh `ring-*` bawaan tidak menggambar apa pun, sementara
   aturan fokus di poin 7 yang menggambar cincinnya. Konsekuensinya `focus-visible:border-ring`
   juga menjadi transparan, jadi tombol berbingkai perlu `focus-visible:border-foreground`
   sendiri supaya bingkainya tidak hilang saat difokus.

9. **Suite e2e mengunci bentuk DOM, bukan hanya perilaku.** Tes `/penjualan` mencari
   `record.locator('p').filter({ hasText: 'Bayar · Tunai' })`; label dan angkanya harus berada
   di `<p>` **yang sama**, dan `getByRole('status')` di keranjang harus memuat kata "Total"
   beserta angkanya. Port yang memecah pasangan itu menjadi dua elemen akan merah walau
   perilakunya utuh — jadi bacalah tes yang menyentuh layar sebelum menggambar ulang
   markahnya, dan biarkan kalimatnya apa adanya.

10. **Tombol berbingkai tanpa `variant` berakhir putih di atas putih.** `Button` tanpa prop
    `variant` memakai varian `default`, yang membawa `text-primary-foreground` (putih) —
    dan kelas `bg-card` di prop `class` menang atas `bg-primary`-nya, jadi latarnya putih
    sementara tintanya tetap putih. Terukur 1,00:1 pada layar Produk (`#ffffff on #ffffff`,
    "Ubah"), dan itu **tak terlihat oleh tes**: `getByRole` tetap menemukan tombolnya, dan
    `toBeVisible()` tetap hijau. Satu-satunya yang menangkapnya adalah pengukuran kontras DOM
    yang berjalan (§6). Karena itu setiap tombol berbingkai memakai `variant="ghost"` lebih
    dulu, baru dibangun dari `class`.

11. **Dialog shadcn: tiga hal yang tidak bisa diselesaikan lewat prop.** Port pertama yang
    memakai `Dialog` menemukan tiga batas yang akan ditemui setiap formulir berikutnya.
    - **Tabirnya tidak punya prop kelas.** `Dialog.Content` merender `<Dialog.Overlay />`-nya
      sendiri, jadi `bg-black/10` + `backdrop-blur-xs` bawaan tidak bisa diganti dari komponen
      domain. Satu-satunya jalan adalah aturan tingkat dunia di `app.css`, di luar `@layer`
      seperti poin 7 — `[data-slot='dialog-overlay'] { background-color: transparent;
      backdrop-filter: none }` — karena tint 10% dan blur bukan kosakata dunia ini. Yang
      memisahkan dialog dari papan tetap ada: selisih `#fafafa` ke `#ffffff` dan garis rambut 1px.
      Bingkai dialog sendiri juga dipindah dari `ring-1` ke `border` sungguhan, supaya
      `box-shadow` benar-benar nol dan bukan garis rambut yang menyamar jadi bayangan.
    - **bits-ui mengunci `document.body` dan mengembalikannya lewat timer, bukan saat unmount.**
      Scroll-lock-nya menulis `pointer-events: none` ke `body` dan baru membersihkannya
      ~24 ms setelah dialog dibongkar. Di jsdom, tes berikutnya mulai di dalam jendela itu dan
      `user-event` menolak setiap klik dengan "Unable to perform pointer interaction". Obatnya
      satu baris di `tests/vitest-setup-client.ts`: `afterEach(() => document.body.removeAttribute('style'))`.
    - **Focus-trap-nya mendarat satu tick setelah klik yang membukanya.** Keystroke yang
      dikirim di antaranya dibaca dialog, bukan field — tombol pertama terisi, yang kedua tidak.
      Hanya tes yang bisa secepat itu, jadi obatnya ada di tes: `openForm()` menunggu
      `document.activeElement` berada di dalam `[data-slot="dialog-content"]` sebelum mengetik.

## Evidence — rel navigasi, 2026-09-26

| Yang diukur | Hasil | DESIGN.md |
| --- | --- | --- |
| Nama toko di rel (setelah Pengaturan disimpan) | `Toko Elektronik Jaya` | nama toko memimpin (PRODUCT.md, ADR-0019) |
| Dasar rel | `rgb(255,255,255)`, tanpa bayangan, `position: sticky` | petak, nol bayangan |
| Garis bawah rel / pemisah baris | `1px rgb(0,0,0)` / `1px rgb(232,232,232)` | tinta menutup rel, garis rambut di dalamnya |
| Tab aktif | `rgb(204,13,13)` + `rgb(255,255,255)`, 600, 13px, radius 0 | merah utilitas, teks putih, radius nol |
| Tab diam | teks `rgb(0,0,0)`, latar transparan, garis rambut kanan | tinta hitam, bukan abu-abu |
| Tag Peran | tinggi 15px, 12px/600, isian `rgb(245,245,245)` | quiet tag |
| Tombol Keluar | tinggi 26px, 13px/600, `box-shadow: none`, `opacity: 1` | tombol petak, keadaan mati tidak dipudarkan |
| Papan | `padding: 8px`, lebar 1440, tanpa gulir mendatar | papan selebar viewport, padding luar 8px |

## Evidence — layar Produk, 2026-09-28

Diukur dari DOM aplikasi yang berjalan pada 1440×900 (`build/` yang baru dibangun, lihat
catatan di bawah). Rasternya: `.impeccable/preview/shots/produk-sveltekit.png`,
`produk-mobile-sveltekit.png`, dan `produk-form-sveltekit.png`.

| Yang diukur | Hasil | DESIGN.md |
| --- | --- | --- |
| Kontras seluruh teks | 13/13 pasangan unik lolos AA | Zero-Grey (tinta hitam, abu-abu hanya garis/isian) |
| `border-radius` selain 0 | tidak ada — termasuk dialog | Square-Corner |
| Bayangan yang terlihat | tidak ada (135/139 elemen `box-shadow: none`; sisanya cincin transparan berukuran nol) | No-Shadow |
| Merah pada permukaan | 0,2% | Three-Percent (≤ 3%) |
| Monospace | 0 elemen | tanpa monospace, termasuk angka |
| Strip judul | 24px/700, `-0.015em`, garis bawah `rgb(0,0,0)` | Title + strip judul |
| Kepala modul (2 modul) | latar `rgb(245,245,245)`, garis bawah `rgb(0,0,0)` | Wash Grey + garis tinta |
| Garis rambut antar baris tabel | tepat satu `rgb(232,232,232)` per batas baris; baris terakhir tanpa garis (sampel piksel) | Shared-Hairline |
| Field saringan | tinggi 26px, radius 0, `box-shadow` transparan berukuran nol | Fields & Inputs |
| Label field | 12px/600, jarak 3px di atas field | Fields & Inputs |
| Tag Nonaktif | tinggi 15px, `opacity: 1`, garis coret | quiet tag; State-Is-Not-Faded |
| Tombol strip (Tambah Produk) | tinggi 26px, 13px/600, latar `rgb(0,0,0)`, teks putih | `.btn--solid` |
| Tombol baris tabel (Ubah/Nonaktifkan/Aktifkan) | tinggi 22px, 12px/600, tinta di atas petak | `.tbl__acts .btn` |
| Tombol kepala modul & kepala dialog | tinggi 26px, 13px/600 | `.btn` |
| Kolom Kode tanpa Kode | `tanpa Kode` (kata, bukan sel kosong) | Do's |
| Kolom Stok | `Stok habis` (0) · `120` · `2 · menipis` | kata di sebelah tandanya |
| `tabular-nums` | kolom Harga, kolom Stok, dan cacah modul Katalog | "setiap angka uang, jumlah, dan Stok", tanpa kecuali |
| Cacah modul Katalog | `5 Produk`, dari daftar yang sudah dimuat — bukan permintaan kedua | Tabs: cacah *tab* yang belum diputuskan |
| 1440×900 | dokumen 900px, tanpa geser mendatar | One-Screen |
| 390×844 | dokumen tanpa geser mendatar; tabel menggeser di dalam kotaknya (491/372); grid saringan satu kolom | layar sempit menumpuk, bukan menyusut |
| Dialog: bingkai | `1px rgb(232,232,232)` sebagai `border` sungguhan, `box-shadow` transparan; radius 0; petak putih | garis 1px, radius 0, nol bayangan |
| Dialog: tabir | latar transparan, `backdrop-filter: none` | No-Tint |
| Dialog: Commit Button | tinggi 40px, latar `rgb(0,0,0)`, radius 0 | Commit Button |
| Dialog: field & label | field 26px radius 0 · label 12px/600 | Fields & Inputs |

## Evidence — layar Stok, 2026-09-28

Diukur dari DOM aplikasi yang berjalan pada 1440×900 (`build/` yang baru dibangun), dengan
skrip Playwright sekali pakai yang menyalin `getComputedStyle` (pola §6). Rasternya:
`.impeccable/preview/shots/stok-sveltekit.png` dan `stok-mobile-sveltekit.png`.

| Yang diukur | Hasil | DESIGN.md |
| --- | --- | --- |
| Kontras seluruh teks | 12/12 pasangan unik lolos AA | Zero-Grey |
| `border-radius` selain 0 | tidak ada | Square-Corner |
| `box-shadow` yang terlihat | nol di seluruh 161 elemen | No-Shadow |
| Merah pada permukaan | 0,1% (hanya tab rel yang aktif) | Three-Percent (≤ 3%) |
| Monospace | 0 elemen | tanpa monospace |
| Strip judul | 24px/700, `-0.015em`, garis bawah `rgb(0,0,0)` | Title |
| Kepala modul (2 modul) | latar `rgb(245,245,245)`, garis bawah `rgb(0,0,0)` | Wash Grey + garis tinta |
| Kolom tabel | tepat 4: Nama, Kode, Stok, Aksi | tabel `Nama \| Kode \| Stok \| Aksi` |
| Quiet tag (`menipis`, `habis`, `Menipis`, `Nonaktif`) | tinggi 15px, 12px/600, isian `rgb(245,245,245)`, `opacity: 1` | Tags; State-Is-Not-Faded |
| Keadaan habis | nama dicoret (`line-through`); kata `Stok habis` di daftar, `Habis` di tabel | "habis kehilangan tintanya", bukan dipudarkan |
| `tabular-nums` | Stok, Kode, ambang, cacah modul, paragraf "Stok di bawah 5", field jumlah, pemberitahuan hasil restock | "setiap angka uang, jumlah, dan Stok", tanpa kecuali |
| Pemberitahuan hasil restock | berbagi satu garis dengan strip lewat `-mb-px` + `z-[2]`; `role="status"` di luar kedua region | Shared-Hairline |
| 1440×900 | dokumen 900px, tanpa geser mendatar | One-Screen |
| 390×844 | tanpa geser mendatar; tabel menggeser di dalam kotaknya | layar sempit menumpuk, bukan menyusut |

## Evidence — layar Pengguna, 2026-09-28

Diukur dari DOM aplikasi yang berjalan pada 1440×900 (`build/` yang baru dibangun), dengan
skrip Playwright sekali pakai yang menyalin `getComputedStyle` (pola §6). Rasternya:
`.impeccable/preview/shots/pengguna-sveltekit.png` dan `pengguna-mobile-sveltekit.png`.

| Yang diukur | Hasil | DESIGN.md |
| --- | --- | --- |
| Kontras seluruh teks | 4/4 pasangan unik lolos AA (terendah 5,78:1 — teks putih di tab rel yang aktif) | Zero-Grey |
| `border-radius` selain 0 | tidak ada, di 83 elemen | Square-Corner |
| `box-shadow` yang terlihat | nol: 80 elemen `none`, 3 cincin transparan berukuran nol | No-Shadow |
| Merah pada permukaan | 0,18% (hanya tab rel yang aktif) | Three-Percent (≤ 3%) |
| Monospace | 0 elemen | tanpa monospace |
| Strip judul | 24px/700, `-0.36px` (=-0.015em), latar petak, garis bawah `rgb(0,0,0)` | Title + strip judul |
| Kepala modul (2 modul) | latar `rgb(245,245,245)`, garis bawah `rgb(0,0,0)`, judul 13px/700 | Wash Grey + garis tinta |
| Tag Peran (`RoleBadge`, 5 buah) | tinggi 15px, 12px/600, isian `rgb(245,245,245)`, `opacity: 1` | quiet tag; satu bentuk untuk Peran |
| Tag `Nonaktif` | tinggi 15px, 12px/600, `opacity: 1`, `line-through` | quiet tag; State-Is-Not-Faded |
| Commit Button hidup | tinggi 26px, latar `rgb(0,0,0)`, teks putih | `.btn--solid`; Commit Button |
| Commit Button mati | latar `rgb(255,255,255)`, `border-style: dashed`, teks tinta, `opacity: 1` | "kehilangan tintanya", bukan memudar |
| Field | tinggi 26px (terukur 26), radius 0, `box-shadow` transparan berukuran nol, label 12px/600 tepat 3px di atasnya | Fields & Inputs |
| Picker Peran | trigger menulis `Kasir`, bukan nilai mentah `kasir` | #33: registry label terisi saat item mount |
| Garis rambut daftar | satu garis per batas: 3×`1px` di baris + `1px` penutup daftar; pindai piksel satu kolom menemukan 1 garis tinta (kepala) + 3 garis rambut berjarak 39px, tanpa garis ganda | Shared-Hairline |
| `tabular-nums` | cacah modul `4 akun` — satu-satunya angka di layar ini | "setiap angka uang, jumlah, dan Stok" |
| 1440×900 | dokumen 900px, tanpa geser mendatar (1440 = 1440) | One-Screen |
| 390×844 | tanpa geser mendatar; kisi form tiga kolom menumpuk jadi satu kolom (356px), kedua modul 374px | layar sempit menumpuk, bukan menyusut |

Keadaan memuat, gagal, dan kosong ketiganya berdiri di modul Daftar; dua di antaranya sudah
lama dijaga tes komponen (`shows the normalized error and a retry`, `says so when the list
comes back empty`) dan tetap utuh setelah port.

Satu temuan lama yang ikut diperbaiki di sini: picker Peran menampilkan nilai mentah `kasir`
selama dropdown-nya belum pernah dibuka, karena `Select.Value` membaca registry label yang
baru terisi saat item mount — persis jebakan yang #33 sudah hindari di layar Produk. Trigger-nya
kini menulis katanya (`PERAN_LABEL`, diekspor dari `RoleBadge` supaya kedua kata Peran hanya
ada di satu tempat), dan kueri `getByText('Kasir')` di satu tes komponen di-scope ke modul
Daftar karena trigger itu kini menemukan kata yang sama.

## Evidence — layar Penjualan, 2026-09-28

Diukur dari DOM aplikasi yang berjalan pada 1440×900 (`build/` yang baru dibangun), dengan
skrip Playwright sekali pakai yang menyalin `getComputedStyle` (pola §6). Rasternya:
`.impeccable/preview/shots/penjualan-sveltekit.png` dan `penjualan-mobile-sveltekit.png`.

| Yang diukur | Hasil | DESIGN.md |
| --- | --- | --- |
| Kontras seluruh teks | 4/4 pasangan unik lolos AA (terendah 5,78:1 — teks putih di atas merah: tab rel aktif dan tag `tersegel`) | Zero-Grey |
| `border-radius` selain 0 | tidak ada | Square-Corner |
| `box-shadow` yang terlihat | nol | No-Shadow |
| Merah pada permukaan | 0,26% (tab rel aktif + tag `tersegel`) | Three-Percent (≤ 3%) |
| Monospace | 0 elemen | tanpa monospace, termasuk angka |
| Strip judul | 24px/700, `-0.015em`, garis bawah `rgb(0,0,0)` | Title + strip judul |
| Kepala modul (2 modul) | latar `rgb(245,245,245)`, garis bawah `rgb(0,0,0)`, judul 13px/700 | Wash Grey + garis tinta |
| Tag `tersegel` | tinggi 15px, 12px/600, latar `rgb(204,13,13)`, teks putih, `opacity: 1` — dengan katanya, bukan warna sendirian | Tags; State-Is-Not-Faded |
| Field Nomor Struk | tinggi 26px, radius 0, `box-shadow` nol, label 12px/600 tepat 3px di atasnya | Fields & Inputs |
| Tombol Cari | tinggi 26px, latar `rgb(0,0,0)`, teks putih | `.btn--solid`; Commit Button |
| Aksi Cetak ulang Struk | selalu terlihat sebagai kata; tidak ada aksi di balik hover | Named-Not-Hidden |
| `tabular-nums` | Nomor Struk, tiap nominal, dan field Nomor Struk — 5 simpul berdigit semuanya tabular | "setiap angka uang, jumlah, dan Stok", tanpa kecuali |
| Rincian Penjualan | memakai `RincianPenjualan.svelte` yang sudah diporting saat layar Kasir, bukan markah baru yang mirip | satu bentuk, dua layar |
| 1440×900 | dokumen 900px, tanpa geser mendatar (1440 = 1440) | One-Screen |
| 390×844 | tanpa geser mendatar; form pencarian dan modul tersimpan menumpuk | layar sempit menumpuk, bukan menyusut |

Keadaan memuat, gagal, dan kosong ketiganya berdiri: memuat sebagai kata (bukan layar kosong),
gagal sebagai pesan Go yang terbaca plus tombol "Coba lagi", dan kosong sebagai kalimat
"Ketik Nomor Struk lalu tekan Cari…". Nomor Struk yang menamai tak satu Penjualan pun
menuliskan pesan Go-nya ("Penjualan tidak ditemukan.") — bukan layar kosong — dan sudah lama
dijaga tes e2e (`a Nomor Struk that names nothing is a readable message`).

Markah yang dikunci e2e tetap utuh: `Bayar · Tunai` dan angkanya berada di `<p>` yang sama,
dan rincian yang dibuka memakai komponen yang sama dengan struk Kasir — bukan salinan kedua
(solution doc §9).

Satu keputusan dunia ikut diselesaikan di sini: **Commit Button** kini menunjuk dua ukuran
yang ditulis di `DESIGN.md` §Commit Button — kolom 40px/15px/600 dan baris 26px/13px/600,
dengan tinta, radius, dan keadaan mati yang sama — supaya layar berikutnya tidak menebaknya
sendiri dari prototipe masing-masing.

## Evidence — layar Laporan, 2026-09-29

Diukur dari DOM aplikasi yang berjalan pada 1440×900 (`build/` yang baru dibangun), dengan
skrip Playwright sekali pakai yang menyalin `getComputedStyle` (pola §6). Rasternya:
`.impeccable/preview/shots/laporan-sveltekit.png` dan `laporan-mobile-sveltekit.png`.

| Yang diukur | Hasil | DESIGN.md |
| --- | --- | --- |
| Kontras seluruh teks | 13/13 pasangan unik lolos AA | Zero-Grey |
| `border-radius` selain 0 | tidak ada | Square-Corner |
| `box-shadow` yang terlihat | nol | No-Shadow |
| Merah pada permukaan | 0,15% (hanya tab rel yang aktif) | Three-Percent (≤ 3%) |
| Monospace | 0 elemen | tanpa monospace, termasuk angka |
| Strip judul | 24px/700, `-0.015em`, garis bawah `rgb(0,0,0)` | Title + strip judul |
| Field Tanggal | tinggi 26px, radius 0, `box-shadow` nol, label 12px/600 tepat 3px di atasnya | Fields & Inputs |
| Kepala modul (Omzet harian, Daftar Penjualan) | latar `rgb(245,245,245)`, garis bawah `rgb(0,0,0)` | Wash Grey + garis tinta |
| Sub-judul Omzet per Kasir | garis rambut, bukan Wash Grey — rincian di dalam modul, bukan kepala modul kedua | Modules |
| Figures (Total omzet, Jumlah transaksi) | 20px/700 `tabular-nums`, didahului label 12px/600 | Figures |
| Metode tanpa transaksi (Debit, Transfer) | `—`, bukan `0` — dan tidak satu pun selnya berisi nol | Figures |
| `tabular-nums` | 16 simpul berdigit di UI (tanggal, cacah, nominal, Nomor Struk, waktu) semuanya tabular | "setiap angka uang, jumlah, dan Stok", tanpa kecuali |
| Aksi Buka / Cetak ulang Struk | selalu terlihat sebagai kata, tidak ada aksi di balik hover | Named-Not-Hidden |
| 1440×900 | dokumen 1440px, tanpa geser mendatar | One-Screen |
| 390×844 | tanpa geser mendatar; tabel menggeser di dalam kotaknya | layar sempit menumpuk, bukan menyusut |

Keadaan memuat, gagal, dan kosong ketiganya berdiri di tiap modul: memuat sebagai kata
("Memuat omzet…"), gagal sebagai pesan Go yang terbaca plus tombol "Coba lagi", dan kosong
sebagai kalimat yang menjelaskan. Hari kosong tidak dijawab dengan empat baris nol: tabelnya
menulis `—`, dan kalimat "Belum ada Penjualan pada tanggal ini." yang berdiri di atasnya yang
mengatakannya. Field Tanggal yang rusak (mis. `2026-02-30`) ditolak dengan pesan 12px/600 di
bawah fieldnya, bukan layar kosong — dan tes di `LaporanHarian.svelte.test.ts` mengunci
ketiganya (angka `—` menggantikan `0`, kalimat field, dan hari kosong).

Markah yang dikunci e2e tetap utuh: "Total omzet" tetap diikuti `Rp 0` pada hari kosong (yang
belum bisa dihitung adalah angka *per metode*, bukan total harinya), baris Kasir tetap di dalam
`region "Omzet harian"`, dan rincian yang dibuka tetap memakai `PenjualanTersimpan` yang sama
dengan layar Penjualan — bukan salinan kedua (solution doc §9).

## Kesalahan yang hampir dilakukan

`pnpm exec playwright test <spec>` **tidak** membangun ulang SvelteKit — ia menyajikan
`build/` yang sudah ada. Setelah mengubah kelas Tailwind, pengukuran akan membaca CSS lama
dan memberi angka yang tampak sah padahal bukan milik kodenya. Jalankan `pnpm build` dulu,
atau pakai `pnpm test:e2e` yang memang `pnpm build && playwright test`. Ini kelas cacat yang
sama dengan `definition-of-done-verification.md`: sinyal hijau yang bukan bukti.

`getComputedStyle` **juga** tidak boleh dibaca sesaat setelah keadaan diubah kalau elemennya
membawa `transition-*`: yang terbaca nilai antara, bukan nilai yang dilihat Admin. Tombol
`Button` membawa `transition-all`, jadi `commit.disabled = true` lalu langsung membaca
`backgroundColor` mengembalikan tinta lamanya — dan itu terbaca persis seperti "varian
`disabled:`-nya tidak bekerja", lengkap dengan selector yang cocok di CSSOM. Dua obat, dan
keduanya dipakai di sini: matikan `style.transition` pada elemen yang sedang diukur, dan beri
jeda sebelum mengukur kontras di halaman yang baru dinavigasi (tab rel bertransisi warnanya
150ms, dan pembacaan di tengah jalan mengukur merah yang belum selesai). Persis kelas cacat
yang sama dengan cincin fokus di `app.css`: transisi 150ms milik primitif yang menyamarkan
keadaan akhirnya.

## Yang belum diputuskan

Cacah pada tab (`Produk 24`, `Stok 4`) ada di prototipe sebagai angka contoh. Angka asli
butuh permintaan ke API dari kerangka, yang berarti setiap layar membayar dua permintaan demi
kerangka. Untuk sekarang tab membawa katanya saja.

Dua hal yang tiap port berikutnya akan temui lagi, dan yang belum punya satu jawaban:

- **Titik menumpuknya kisi field punya dua angka.** Prototipe layar Produk menumpuk kisi
  empat kolomnya di 1080px (`min-[1081px]:grid-cols-4`, ikut titik papan), sedangkan prototipe
  layar Pengguna menumpuk kisi tiga kolomnya di 900px (`min-[901px]:grid-cols-3`, dari
  `@media (max-width: 900px)` milik `style.css`). Keduanya setia pada prototipenya
  masing-masing — tapi "kapan kisi field menumpuk" adalah satu keputusan dunia yang sekarang
  dijawab dua angka, dan yang membacanya harus tahu itu disengaja, bukan kelalaian.
- **Kelas quiet tag disalin per komponen.** `TAG`/`TAG_NONAKTIF` sekarang hidup di
  `ProdukList`, `StokList`, dan `PenggunaList` (§1 menoleransi pengulangan di tengah port).
  Setelah layar terakhir mendarat, satu `lib/components/shared/` yang memuat strip judul,
  kepala modul, tag, dan Commit Button akan menghapus tiga salinan sekaligus — dan itu
  perubahan yang harus ditunggu sampai tidak ada port yang sedang berjalan.

## Tabular-nums — angka Inter Variable proporsional, 2026-09-28

Diukur pada Inter Variable 13px, sepuluh digit:

| `font-variant-numeric` | `1111111111` | `0000000000` | `8888888888` |
| --- | --- | --- | --- |
| `normal` | 50px | 80px | 80px |
| `tabular-nums` | 80px | 80px | 80px |

Jadi kelas itu **menahan lebar**, bukan hiasan: angka yang berubah tanpa kelasnya menggeser
tetangganya sampai 30px per sepuluh digit. Dua tempat paling terasa dampaknya — cacah yang
berubah sambil orang mengetik di kolom saringan, dan nominal yang sedang diketik di dalam
fieldnya sendiri. DESIGN.md menuntutnya "tanpa kecuali", dan prototipe menandai `tnum` bahkan
di field nominal (`<input id="bayar" class="input tnum">`).

**Cara memeriksa, dan hasilnya pada Masuk + Kasir + rel.** Telusuri setiap elemen yang simpul
teksnya sendiri memuat digit, lalu baca `font-variant-numeric`-nya — membaca kelasnya satu per
satu di layar akan melewatkan yang justru paling sering berubah. Audit pertama menemukan 16
elemen berangka; 5 bukan `tabular-nums`, dan **3 di antaranya benar-benar angka**:

| Elemen | Sebelum | Sesudah | Kenapa |
| --- | --- | --- | --- |
| Cacah Katalog (`{n} Produk Aktif · tekan satu petak…`) | `normal` | `tabular-nums` | berubah sambil saringan diketik; prototipe menandai cacah yang sama `tnum` |
| Ringkasan Keranjang (`{n} unit dalam {m} Item.`) | `normal` | `tabular-nums` | berubah setiap unit; prototipe: `mod__note tnum` |
| Field "Jumlah bayar" (`#kasir-bayar`) | `normal` | `tabular-nums` | nominal yang sedang diketik; prototipe: `class="input tnum"` |

Sisanya nama Produk yang sekadar memuat digit (`… E2E`) — bukan angka menurut aturan ini.

Sengaja **tidak** disentuh: kolom pencarian Kode dan Nama (prototipe tidak menandainya, dan
Kode adalah barcode — bukan uang, jumlah, atau Stok), dan nama Pengguna di rel (prototipe
menulis `class="tnum"` pada `admin`, string tanpa angka, jadi tanda itu tidak menggambar apa
pun — bukan pernyataan tentang digit).

Field angka memakai satu const terpisah dari pakaian field-nya — `FIELD_ANGKA = ${FIELD}
tabular-nums` di `Pembayaran.svelte` — supaya yang dibaca dari kelasnya bukan "field",
melainkan "field yang isinya angka". Kolom pencarian memakai `FIELD` biasa.

## Related

- `DESIGN.md` — sumber token, ukuran, dan aturan bernama (Zero-Grey, State-Is-Not-Faded, No-Shadow).
- `docs/solutions/developer-experience/definition-of-done-verification.md` — kenapa kelima
  perintah §12 harus dijalankan, dan kenapa exit code lebih dipercaya daripada teksnya.
- `docs/adr/0019-nama-toko-baris-pertama-header-endpoint-publik.md` — nama toko dan endpoint publiknya.

## Permukaan peramban & gerak — 2026-09-28

DESIGN.md punya bagiannya sendiri untuk ini ("Browser surfaces": teks terpilih, caret,
scrollbar, cincin fokus) dan Do's menutupnya lagi: *"theme the browser surfaces — selection,
caret, scrollbar, focus ring — from this palette"*. Prototipe menulis keempatnya di
`sistem/style.css` dengan alasan yang sama: *"bawaan itu bukan milik dunia mana pun"*.

Port ini sempat kehilangan hampir semuanya. Dari empat permukaan itu, **hanya cincin fokus**
yang mendarat (di poin 7–8 di atas). Yang hilang, terukur sebelum perubahan:

| Yang diukur di aplikasi yang berjalan | Sebelum | Sesudah |
| --- | --- | --- |
| `::selection` | latar `rgba(0,0,0,0)` — jatuh ke biru bawaan peramban | `var(--foreground)` di atas `var(--card)` |
| `caret-color` html | `rgb(0,0,0)` — kebetulan benar, sebagai bawaan peramban | `var(--foreground)` |
| `caret-color` `#kasir-bayar` | `rgb(0,0,0)` | `rgb(204,13,13)` (merah utilitas, lewat `data-caret="utility"`) |
| `scrollbar-width` / `-color` | `auto` / `auto` | `thin` / garis rambut di atas dasar |
| `prefers-reduced-motion: reduce` → `transition-duration` baris | `0.15s` (diabaikan) | `0s` |
| idem → `animation-duration` dialog | `0.1s` (diabaikan) | `0s` |

Prototipe hanya menolkan `transition-duration`; di sini `animation-duration` ikut, karena animasi
masuknya dialog dan dropdown datang dari `tw-animate-css` dan keduanya juga gerak.

**Kenapa tidak satu pun tes bisa melihatnya.** Keempatnya bukan teks, bukan tata letak, dan
bukan nama: `getByRole`, `toBeVisible`, kontras teks, radius, bayangan, `tabular-nums` — semuanya
diam. Yang menemukannya hanya membaca empat properti itu dari DOM yang berjalan, dan
trapesiumnya adalah `grep` di `app.css`: **tidak satu pun dari keempat kata itu ada di sana.**

### Cincin fokus yang masuk secara halus — dan cincin yang tak terlihat

Cincinnya sendiri sudah ada, tetapi ia **beranimasi masuk**, dan itu akar yang sama: primitif
shadcn membawa `transition-all`, sementara prototipe hanya mentransisikan latar dan garis. Jadi
`outline-width` dan `outline-offset` milik cincin ikut bertransisi:

| Kontrol | Saat fokus mendarat | Sesudah tenang |
| --- | --- | --- |
| Button shadcn (petak putih) | `3px solid rgb(0,0,0)` offset `0px` | `2px solid rgb(0,0,0)` offset `-2px` |
| Button shadcn (bidang bertinta) | `3px solid rgb(255,255,255)` offset `0px` | `2px solid rgb(0,0,0)` offset `-2px` |
| Input shadcn | `2px solid` — benar sejak awal | sama |
| Select trigger bits-ui | `2px solid` — benar sejak awal | sama |

Jadi pada tombol, ~150ms pertama cincinnya salah lebar, salah offset, dan salah warna — di atas
bidang bertinta ia bahkan sempat putih. Obatnya satu aturan tingkat dunia, di luar `@layer`:
`:focus-visible { transition-property: background-color, border-color, color }` — cincinnya
mendarat seketika, sementara latar dan garisnya tetap halus.

**Diputuskan: petak.** Setelah tenang, cincinnya `2px solid var(--foreground)` di atas bidang
bertinta pada tombol aksi utama — strip Produk, Commit Button, field Masuk — terukur **1,00:1**,
tak terlihat, dan itu salah satu pengendali pertama yang dituju Tab. Prototipe melakukan hal yang
sama (`:focus-visible` global dengan `outline: 2px solid var(--ink)`, dan `.method:focus-within`
pun memakai tinta padahal sel metode yang terpilih justru bidang bertinta), jadi port ini setia;
yang cacat sistemnya, dan DESIGN.md hanya menyebut "outline 2px ke dalam" tanpa warna.

Tiga kandidat diukur pada dua permukaan sekaligus, dan raster perbandingannya — diperbesar 3×
supaya cincin 2px bisa dinilai mata — disimpan sebagai rekaman keputusan:

| Kandidat | Di atas tinta `rgb(0,0,0)` | Di atas tab merah `rgb(204,13,13)` |
| --- | --- | --- |
| tinta (sebelum) | 1,00:1 ✗ | 4,33:1 ✓ |
| **petak (dipilih)** | **21:1** ✓ | **4,85:1** ✓ |
| merah utilitas | 5,88:1 ✓ | 1,00:1 ✗ |

Merah ditolak bukan karena seleranya: ia **memindahkan** cacatnya ke tab rel yang aktif, dan
DESIGN.md sudah membatasi merah pada harga dan tab. Petak satu-satunya warna palet yang lolos di
kedua permukaan, dan ia membaca sebagai kebalikan cincin dunia ini — tombol yang hidup kehilangan
tintanya di tepinya, seperti tombol yang mati kehilangan seluruhnya.

Aturannya berdiri sekali di `app.css`, dan menemukan bidang bertintanya lewat `bg-primary`:
`--primary` memetakan ke tinta, dan varian `default` milik `Button` — satu-satunya bidang bertinta
penuh yang sistem ini kenal — membawa kelas itu. Jadi tidak ada satu pun tombol yang perlu
ditambahi apa pun, dan port berikutnya tidak bisa lupa. Sel metode Pembayaran adalah pengecualian
yang membuktikan aturannya: tintanya datang dari `has-[:checked]`, bukan dari `bg-primary`, jadi
cincinnya ditanggung di komponennya.
