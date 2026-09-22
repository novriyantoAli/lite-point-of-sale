---
title: "Definition of Done §12 — check hijau bukan bukti"
date: 2026-09-22
category: developer-experience
module: quality-gates
problem_type: developer_experience
component: development_workflow
severity: medium
applies_when:
  - "Sebelum menandai sebuah issue selesai atau membuka PR"
  - "Ketika sebuah perintah melaporkan sukses tetapi perilakunya belum pernah dijalankan"
  - "Ketika output perintah melewati tool yang memotong atau meringkasnya"
symptoms:
  - "`pnpm check` melaporkan 0 error sementara 6 tes merah"
  - "Kegagalan `prettier --check` tidak terlihat karena outputnya terpotong"
root_cause: missing_workflow_step
resolution_type: workflow_improvement
tags: [definition-of-done, verification, exit-code, svelte-check, prettier, green-signal, truncated-output]
---

# Definition of Done §12 — check hijau bukan bukti

## Context

Saat mengerjakan issue #4 (Produk) ada tiga kejadian terpisah di mana sebuah perintah melaporkan sukses padahal kodenya salah. Ketiganya tertangkap — tetapi hanya karena perintah **lain** dijalankan. Yang membuatnya berbahaya: tidak ada bedanya secara visual antara `pnpm check` yang hijau di kode benar dan `pnpm check` yang hijau di kode salah.

Akar kejadian pertama adalah pekerjaan setengah jadi yang **tampak** selesai: pohon kerja berisi domain `produk` dengan skema dan API lengkap plus tesnya, `pnpm check` bersih, backend hijau. Yang belum pernah dijalankan adalah `pnpm test` — dan tes itu merah enam.

## Guidance

Jalankan **seluruh** perintah §12, dan baca **exit code**-nya, bukan teks outputnya.

Setiap perintah menangkap kelas cacat yang tidak bisa dilihat perintah lain. Menjalankan sebagian berarti membiarkan kelas cacat tertentu lolos tanpa penjaga:

| Perintah | Yang ditangkap | Yang tidak ditangkap |
| --- | --- | --- |
| `pnpm check` (svelte-check) | tipe, prop yang salah, `undefined` yang tidak dijaga | logika runtime, perilaku, format |
| `pnpm test` (vitest) | perilaku schema, `api/`, komponen | format, integrasi browser sungguhan |
| `pnpm lint` (prettier + eslint) | format + aturan lint | tipe, perilaku |
| `pnpm build` | bundling, impor yang putus, kompilasi SSR | perilaku |
| `pnpm test:e2e` (Playwright) | alur nyata di browser nyata + BFF + Go | apa pun yang tidak tersentuh alur itu |

`pnpm check` adalah yang **terlemah** dari kelimanya, tetapi juga yang paling murah dan paling cepat — dan justru itu yang membuatnya berbahaya sebagai satu-satunya penjaga. Typecheck menjawab "apakah bentuknya konsisten?", bukan "apakah hasilnya benar?".

### Baca exit code, bukan teks output

Kejadian ketiga bukan cacat kode melainkan cacat verifikasi: `pnpm lint` gagal dan saya membacanya sebagai lulus, karena output perintah terpotong oleh tool dan saya mencocokkan teks, bukan status keluar.

Pola yang menutup celah ini:

```bash
for c in check lint build; do
  pnpm "$c" > "/tmp/dod-$c.txt" 2>&1
  echo "pnpm $c EXIT=$?"
done

pnpm test > /tmp/dod-test.txt 2>&1; echo "pnpm test EXIT=$?"
pnpm test:e2e > /tmp/dod-e2e.txt 2>&1; echo "pnpm test:e2e EXIT=$?"
```

Simpan output ke file, cetak exit code, baru baca file bila ada yang bukan nol. Ini juga menutup masalah kedua: output panjang yang dipotong di tengah membuat baris kegagalan pertama tidak pernah terlihat.

Untuk backend, padanannya: `gofmt -l .` (harus kosong), `go vet ./...`, `go test ./...`.

## Why This Matters

Tanpa keduanya, dua hal terjadi dan keduanya sudah pernah terjadi di repo ini:

1. **Cacat perilaku bisa lolos sampai PR.** Skema `produk` punya dua bug nyata — `optionalText` hanya `.nullable()` bukan `.optional()` (sehingga Kode yang absen gagal validasi dan pesannya **menutupi** error Harga/Stok), dan `z.coerce.number()` mengeluarkan pesan bawaan zod untuk NaN sebelum `.int()` sempat jalan. Keduanya tidak terlihat oleh `pnpm check`, karena keduanya bug *perilaku*, bukan bug tipe.
2. **Kegagalan format sampai ke CI.** `prettier --check` gagal pada 4 file dan baru ketahuan di job `Frontend (SvelteKit)`. CI menangkapnya, tetapi seharusnya tidak perlu sampai ke sana.

Ada alasan struktural mengapa ini penting khusus untuk agen: agen cenderung menjalankan perintah termurah lebih dulu dan berhenti saat hijau. "Hijau" terasa seperti bukti, padahal ia hanya bukti untuk satu kelas cacat.

## When to Apply

- Sebelum menandai issue selesai, sebelum commit, dan sebelum membuka PR.
- Setelah mengambil alih pohon kerja yang belum di-commit — terutama bila tidak jelas perintah apa yang sudah dijalankan. Jalankan seluruh §12 dari awal; jangan percaya keadaan "sepertinya sudah jalan".
- Kapan pun sebuah perintah hijau tetapi perilakunya belum pernah benar-benar dijalankan.
- Kapan pun output perintah melewati tool yang memotong, meringkas, atau membungkusnya.

## Examples

### `pnpm check` hijau sementara enam tes merah

```
$ pnpm check
svelte-check found 0 errors and 0 warnings

$ pnpm test
 Test Files  1 failed | 14 passed (15)
      Tests  6 failed | 106 passed (112)
```

Penyebabnya satu baris:

```ts
// Salah — `.nullable()` menerima null, bukan field yang absen
const optionalText = z.string().trim().nullable().transform(...)

// Benar
const optionalText = z.string().trim().nullable().optional().transform(...)
```

Fixture tes menghilangkan `code` sepenuhnya. Karena `code` wajib ada, zod mengeluh soal `code` lebih dulu dan pesan asli tentang `price`/`stock` tidak pernah muncul. Typecheck tidak punya cara melihat ini: tipe outputnya tetap `string | null`, hanya jalur runtime-nya yang salah.

### `pnpm lint` gagal tetapi terbaca lulus

```
$ pnpm lint
Checking formatting...
[warn] src/lib/domains/produk/api/produk.api.test.ts
[warn] src/lib/domains/produk/components/ProdukList.svelte
[warn] Code style issues found in 4 files.
```

Terbaca sebagai "ESLint: No issues found" karena bagian kegagalan terpotong dari output yang saya lihat. Yang membuktikan sebaliknya:

```bash
$ npx prettier --check . > /tmp/pc.txt 2>&1; echo "CHECK EXIT=$?"
CHECK EXIT=1
```

Perhatikan bahwa `pnpm lint` adalah `prettier --check . && eslint .` — ketika prettier gagal, **eslint tidak pernah dijalankan**. Jadi output "ESLint: No issues found" yang terlihat itu bahkan bukan hasil dari perintah yang sama.

## Related

- Issue #4 (`Produk: kelola katalog`) dan PR #16.
- Skill `frontend-ddd-sveltekit` §12 — daftar perintah DoD yang jadi rujukan tabel di atas.
- `docs/adr/0007-testing-strategy.md` — pembagian tanggung jawab antara tes unit dan e2e.
- `docs/adr/0009-e2e-playwright-ci.md` — job `e2e` di CI dan mengapa ia menjalankan proses sungguhan.
