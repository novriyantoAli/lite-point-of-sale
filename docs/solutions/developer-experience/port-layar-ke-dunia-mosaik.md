---
title: "Port layar ke dunia mosaik — varian shadcn yang bertabrakan dengan DESIGN.md"
date: 2026-09-26
last_updated: 2026-09-26
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
  - "Hasil pengukuran Playwright tidak berubah setelah kode diubah"
root_cause: framework_constraint
resolution_type: workflow_improvement
tags: [frontend, design-system, shadcn-svelte, tailwind-v4, tailwind-merge, accessibility, verification, playwright]
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

3. **`shadow-none` tidak mengalahkan `shadow-xs`.** Pada tombol `variant="outline"`,
   `shadow-none` terukur tetap menyisakan `0 1px 2px rgba(0,0,0,0.05)`. Yang bekerja: pilih
   varian tanpa bayangan (`variant="ghost"`) lalu bangun tampilannya dari `class`. Aturan
   umumnya — kalau sebuah varian membawa hiasan yang dilarang dunia ini, jangan bertarung
   dengan kelas; pilih varian yang bersih.

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

## Kesalahan yang hampir dilakukan

`pnpm exec playwright test <spec>` **tidak** membangun ulang SvelteKit — ia menyajikan
`build/` yang sudah ada. Setelah mengubah kelas Tailwind, pengukuran akan membaca CSS lama
dan memberi angka yang tampak sah padahal bukan milik kodenya. Jalankan `pnpm build` dulu,
atau pakai `pnpm test:e2e` yang memang `pnpm build && playwright test`. Ini kelas cacat yang
sama dengan `definition-of-done-verification.md`: sinyal hijau yang bukan bukti.

## Yang belum diputuskan

Cacah pada tab (`Produk 24`, `Stok 4`) ada di prototipe sebagai angka contoh. Angka asli
butuh permintaan ke API dari kerangka, yang berarti setiap layar membayar dua permintaan demi
kerangka. Untuk sekarang tab membawa katanya saja.

## Related

- `DESIGN.md` — sumber token, ukuran, dan aturan bernama (Zero-Grey, State-Is-Not-Faded, No-Shadow).
- `docs/solutions/developer-experience/definition-of-done-verification.md` — kenapa kelima
  perintah §12 harus dijalankan, dan kenapa exit code lebih dipercaya daripada teksnya.
- `docs/adr/0019-nama-toko-baris-pertama-header-endpoint-publik.md` — nama toko dan endpoint publiknya.
