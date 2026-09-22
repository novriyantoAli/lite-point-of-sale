# Glosarium tunggal di `CONTEXT.md`; `CONCEPTS.md` tidak dibuat

Dua lapis skill di repo ini mencari kosakata domain di file yang berbeda, dan salah satu file itu tidak ada: alur `domain-modeling`/`grill-with-docs`/`ce-brainstorm`/`ce-plan` membaca **`CONTEXT.md`** lewat `docs/agents/domain.md`, sementara keluarga `ce-compound` mencari **`CONCEPTS.md`** di root. `AGENTS.md` sudah menetapkan `CONTEXT.md` sebagai satu-satunya glosarium, dan `CONCEPTS.md` tidak pernah ada. Keputusannya: **`CONTEXT.md` tetap satu-satunya glosarium; `CONCEPTS.md` tidak dibuat, dan setiap vocabulary capture ditulis ke `CONTEXT.md`.**

## Di mana batasnya

| Yang mencari kosakata | File | Catatan |
| --- | --- | --- |
| `docs/agents/domain.md`, alur Matt Pocock | `CONTEXT.md` | satu-satunya yang dirujuk `AGENTS.md` |
| `ce-compound`, `ce-compound-refresh` | `CONCEPTS.md` | satu-satunya yang **membuat** file (Phase 2.4 / 4.5) |
| `ce-plan`, `ce-brainstorm`, `ce-pov`, `ce-explain`, `ce-debug`, `ce-ideate`, `ce-optimize`, `ce-code-review` | `CONCEPTS.md` | hanya membaca; eksplisit "creation is owned by ce-compound and ce-compound-refresh" |

## Kenapa tidak mengikuti `CONCEPTS.md`

**Seed `CONCEPTS.md` akan melanggar ADR-0012.** Aturan seed-nya adalah "core domain nouns the area's declared domain model exposes — schema, core types, primary models, top-level domain docs". Di repo ini declared domain model terbelah dua oleh ADR-0012: istilah Indonesia di glosarium/route/tipe frontend, istilah Inggris di identifier Go dan DTO JSON. Seed dari `produk.schema.ts` menghasilkan glosarium Indonesia yang menduplikasi `CONTEXT.md`; seed dari `domain/produk/produk.go` menghasilkan glosarium Inggris (`Product`, `Sale`) yang mengajarkan persis drift yang ADR-0011 dan ADR-0012 larang. Keduanya buruk, jadi tidak ada arah seed yang benar.

**Dua glosarium akan divergen.** `CONTEXT.md` memuat daftar `_Avoid_` yang sengaja dibangun (`_Avoid_: SKU, barcode`, `_Avoid_: Transaksi, order, sale`). Glosarium kedua tidak akan memuatnya, jadi sinonim yang sudah diputuskan dibuang akan hidup lagi — dan entri `docs/solutions/` akan di-ground pada kosakata yang berbeda dari PRD #1 dan ADR.

**Split-brain-nya sudah hidup sekarang, bukan risiko nanti.** Dalam satu run `ce-code-review`, `repo-profiler` mengisi slot `vocabulary` dari `CONCEPTS.md` (tidak ada, jadi kosong) sementara langkah domain-docs di skill yang sama membaca `CONTEXT.md`. `learnings-researcher` malah melewati grounding kosakata sepenuhnya ("skip entirely if absent").

## Considered Options

- **Adopsi `CONCEPTS.md`, pensiunkan `CONTEXT.md`** — ditolak: `CONTEXT.md` dirujuk `AGENTS.md`, `docs/agents/domain.md`, ADR-0011, ADR-0012, PRD #1, dan seluruh alur skill; commit `760f019` mengonfigurasi alur itu dengan sengaja. Membalik keputusan repo demi menyesuaikan asumsi satu skill adalah biaya besar untuk keuntungan nol di sisi domain.
- **`CONCEPTS.md` berisi pointer ke `CONTEXT.md`** — ditolak: `ce-compound` Phase 2.4 berbunyi "if `CONCEPTS.md` exists at repo root, **add missing qualifying terms**", jadi file pointer itu akan di-append dan berubah menjadi glosarium kedua — persis yang ingin dicegah.
- **Tunda sampai `/ce-compound` pertama dijalankan** — ditolak: begitu file itu dibuat, keputusannya berubah dari "pilih satu nama file" menjadi "migrasi dan rekonsiliasi dua glosarium yang sudah divergen".

## Consequences

- Entri baru ke `CONTEXT.md` mengikuti format entri yang sudah ada (`**Istilah**`, definisi satu paragraf, baris `_Avoid_`), sebagaimana `ce-plan` sudah menerapkan "follow the format set by existing entries".
- Ini penjaga **lunak**, dan itu diakui: `AGENTS.md` tidak bisa mematikan asumsi skill. Agen yang menghormati `AGENTS.md` menulis ke `CONTEXT.md`; agen yang mengikuti Phase 2.4 secara literal tetap bisa membuat `CONCEPTS.md`. Bedanya, kejadian itu menjadi pelanggaran keputusan yang tercatat — bisa di-review dan dibalik — bukan kejutan yang baru ketahuan setelah dua glosarium divergen.
- Grounding kosakata `learnings-researcher` tetap kosong di repo ini. Itu diterima: lebih baik kosong daripada salah.
- Skill tidak diubah oleh ADR ini. Asumsi `CONCEPTS.md` hidup di `~/.pi/agent/skills/` — di luar repo, hilang saat skill di-update, dan tidak bisa dijamin dari sini.
- Sinyal tinjau ulang: kalau `/ce-compound` atau `/ce-compound-refresh` pertama (kemungkinan saat learning kedua ditulis, setelah #5 atau #6) benar-benar membuat `CONCEPTS.md`, penjaga lunak ini gagal dan perlu penjaga keras — patch idempoten atas salinan vendored `repo-profiler.md` dan `repo-profile-cache.md`/`.py` (masing-masing sembilan skill), serta `learnings-researcher.md` (empat skill), disimpan di repo supaya bisa dijalankan ulang setelah update skill.
