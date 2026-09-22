---
title: "Sinyal hijau, subjek salah: squash merge PR bertumpuk dan file yang tak pernah diperiksa"
date: 2026-09-23
category: developer-experience
module: quality-gates
problem_type: developer_experience
component: development_workflow
severity: medium
applies_when:
  - "Ketika PR ditumpuk (stacked) di atas branch PR lain dan akan di-squash merge"
  - "Ketika sebuah check hijau dilaporkan untuk commit atau file yang berbeda dari yang akan digabung"
  - "Ketika file yang diubah tidak disentuh oleh satu pun perintah di Definition of Done §12"
  - "Sebelum memilih base branch, menulis body PR, atau menandai issue selesai"
  - "Setelah menggabungkan beberapa PR yang menyentuh file yang sama"
symptoms:
  - "`grep -c \"^### Documented solutions\" AGENTS.md` mengembalikan 2 di main — heading dan paragraf yang byte-identik muncul dua kali"
  - "CI melaporkan 3/3 hijau, tetapi untuk commit db150c8, bukan head branch 59f04ab yang akan digabung"
  - "Body PR mengklaim perbaikan yang tidak ada di diff PR; angka yang dikutip meleset: go test 175 vs 177, pnpm test 131 vs 144, e2e 15 vs 17"
  - "Seluruh §12 (gofmt, go vet, go build, go test, pnpm build, check, lint, test, test:e2e) hijau, tetapi tidak ada satu perintah pun yang membaca AGENTS.md"
  - "Diff `db150c8..59f04ab` berisi 17 file, +413/-52, yang belum pernah dilihat CI"
root_cause: missing_workflow_step
resolution_type: workflow_improvement
related_components:
  - tooling
  - documentation
tags: [green-signal, squash-merge, stacked-pr, stale-pr-head, verification-gap, definition-of-done, merge-base, ci-verification]
---

# Sinyal hijau, subjek salah: squash merge PR bertumpuk dan file yang tak pernah diperiksa

## Context

Tiga PR terbuka dan semuanya hijau: #17 (`98f2019`, menambah section `### Documented solutions` di `AGENTS.md` + learning pertama di `docs/solutions/`), #18 (`2fa3409`, menambah `docs/adr/0013-single-glossary-context-md.md` dan menyunting paragraf `### Domain docs` di `AGENTS.md`), dan #16 (`5305ffa`, fitur katalog Produk). #18 sengaja **ditumpuk (stacked)** di atas branch #17 supaya `AGENTS.md` tidak konflik.

Motif stacking-nya perlu dicatat dengan jujur, karena motif itu salah: `.github/workflows/ci.yml` hanya terpicu oleh `pull_request: branches: [main]`, jadi PR yang di-base ke branch non-main dianggap tidak punya CI sama sekali — bukan berbagi CI milik PR induknya. `gh pr checks 18` benar-benar menjawab `no checks reported on the 'docs/adr-0013-glosarium-tunggal' branch`. Jadi stacking menghindari konflik dua baris teks dengan biaya kehilangan seluruh cakupan CI. Mengembalikannya pun tidak cukup dengan retarget base ke `main`, karena perubahan base tidak memicu event `synchronize`; perlu satu `git commit --amend` + `push --force-with-lease` untuk memicu CI di commit yang sudah ada.

Urutan merge: #17, lalu #18, lalu #16, semuanya **squash merge** (konvensi repo: tiap PR jadi satu commit berjudul `"<judul PR> (#N)"`).

Setelah #18 masuk main, section `### Documented solutions` beserta paragrafnya muncul **dua kali** di `AGENTS.md` — byte-identik, di baris 15-17 dan 19-21. Hanya `AGENTS.md` yang terdampak: `docs/solutions/developer-experience/definition-of-done-verification.md` tetap 130 baris dengan satu heading, dan 13 ADR utuh.

**Prediksi yang salah.** Sebelum merge, saya memperkirakan hal yang sebaliknya. Saya menyatakan bahwa merge tiga-arah akan mengenali "sisipan identik di kedua sisi" sebagai satu perubahan dan membuangnya. Itu tidak berlaku untuk sisipan. Penyebab sebenarnya: squash merge menghitung ulang diff terhadap merge base **lama**. Branch #18 memuat commit #17 sebagai ancestor, sehingga `diff(16708d2..cc2650d)` ikut membawa sisipan `AGENTS.md` milik #17. Meng-squash diff itu di atas main yang sudah berisi sisipan yang sama menerapkannya untuk kedua kalinya. Prediksi itu terasa benar karena terdengar seperti perilaku three-way merge yang benar-benar ada (untuk *perubahan* identik di kedua sisi), tetapi squash merge bukan three-way merge pada level konten — ia menyusun ulang diff di atas base yang berbeda.

**Prakondisinya sudah terbentuk sejak pagi** (session history). Pada 2026-09-22 ada dua suntingan `AGENTS.md` independen di branch yang berbeda: satu commit langsung ke `main` pada ~09:36 yang menambah pointer "Code organization" ke ADR-0004–0007, dan satu lagi di branch `docs/dod-verification` pada ~15:17 yang menambah pointer `docs/solutions/`. Itu persis bentuk yang membuat squash merge menggandakan sebuah section. **Catatan kehati-hatian:** kelima sesi yang disisir tidak memuat insiden penggandaan itu sendiri — yang terekam di sana hanya prakondisinya, bukan kejadiannya.

Friksinya: **tidak ada satu pun penjaga di repo ini yang bisa melihat cacat ini.** Duplikat itu lolos seluruh Definition of Done (skill `frontend-ddd-sveltekit` §12): `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` (177 tes), `pnpm build`, `pnpm check` (0 error 0 warning), `pnpm lint`, `pnpm test` (144 tes), `pnpm test:e2e` (17 tes) — dan lolos CI di main. Sebabnya bukan penjaga yang salah baca, tetapi tidak adanya penjaga: `pnpm lint` adalah `prettier --check . && eslint .` yang dijalankan dari `frontend/`, jadi prettier tidak pernah melihat markdown di root (`AGENTS.md`, `docs/**`), dan tidak ada tes yang membaca `AGENTS.md`. Duplikat itu akhirnya ditemukan hanya dengan **membaca hasil merge**.

Bahwa wilayah itu memang tidak pernah terjaga juga terlihat dari kebiasaan yang terbentuk sebelum insiden ini (session history): verifikasi suntingan `AGENTS.md` di PR #17 hanya berupa membaca ulang filenya, dan `README.md` di root bahkan harus diformat lewat pemanggilan prettier manual dari `frontend/` terhadap `../README.md` — tanpa satu pun perintah §12 yang menyentuhnya.

## Guidance

### 1. Perlakuan terhadap stacking

Tiga pilihan yang jujur, dengan konsekuensinya masing-masing:

- **(a) Jangan menumpuk.** Batas PR = batas branch = `main`. Tidak ada diff yang dihitung ulang di atas base yang bergerak. Biaya: konflik teks pada file yang sama harus diselesaikan manual di PR kedua, dan PR kedua tidak punya CI selama ia belum di-rebase ke main (karena `ci.yml` hanya `pull_request: branches: [main]`).
- **(b) Menumpuk, tetapi rebase ke base yang sudah diperbarui sebelum merge.** Setelah PR induk (#17) masuk main, rebase branch #18 ke `origin/main` dan push ulang sebelum merge. Setelah rebase, branch #18 hanya berisi commit-nya sendiri, sehingga diff terhadap base yang baru tidak lagi memuat sisipan #17.
- **(c) Tetap menumpuk, tetapi selalu baca pohon pasca-merge.** Merge, lalu periksa hasilnya secara manual di area yang tidak diperiksa penjaga mana pun.

**Rekomendasi: (a) untuk PR dokumentasi/teks, dan (b) + (c) bila stacking memang dibutuhkan.** Alasannya:

- Konflik yang dihindari oleh stacking di sini adalah konflik dua baris teks di `AGENTS.md` — jauh lebih murah diselesaikan manual daripada biaya satu cacat tak terdeteksi di file instruksi yang dibaca setiap agen. Tradeoff melawan alasan stacking dipakai harus diakui: tanpa stacking, #18 memang kehilangan CI; tetapi CI-nya sendiri tidak memeriksa `AGENTS.md`, jadi CI yang "hilang" itu tidak akan menangkap cacat kelas ini juga. **Stacking dibayar dengan risiko, bukan dengan cakupan verifikasi.**
- (b) menutup penyebab akar (diff yang dihitung ulang di atas base yang bergerak), tetapi bergantung pada disiplin: rebase mudah terlupa justru karena semuanya tampak hijau. Karena itu (c) bukan alternatif (b) melainkan **backstop wajib**. Yang menangkap duplikat ini di sesi ini adalah (c).
- (c) saja tidak cukup sebagai kebijakan: ia menemukan cacat setelah cacat masuk main (butuh PR #19, `48d8c99`, untuk membalikkannya). Ia menyelamatkan, tetapi tidak mencegah.

Bila memilih (b), periksa hasil rebase dengan membandingkan diff terhadap base yang baru sebelum membuka/menyegarkan PR:

```bash
git fetch origin
git rebase origin/main
# diff harus HANYA memuat perubahan milik PR ini, bukan milik PR induk
git diff origin/main...HEAD --stat
```

Bila diff memuat perubahan milik PR induk, stacking belum selesai di-rebase — jangan merge.

### 2. Menentukan file mana yang tidak diperiksa penjaga mana pun

Metodenya bukan menebak, tetapi menanyakan satu pertanyaan per file yang disentuh PR: **perintah §12 mana yang membaca file ini?** Bila jawabannya tidak ada, file itu butuh pembacaan manual setelah merge — dan CI tidak bisa menutupinya, karena CI menjalankan perintah yang sama.

Untuk repo ini jawabannya dapat dihitung dari cwd tiap perintah:

| Wilayah file | Penjaga yang membacanya |
| --- | --- |
| `frontend/**` | `pnpm lint` (`prettier --check . && eslint .` dari `frontend/`), `pnpm check`, `pnpm test`, `pnpm build`, `pnpm test:e2e` |
| `backend/**` | `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` |
| `AGENTS.md`, `README.md`, `CONTEXT.md`, `docs/**`, `.github/**` | **tidak ada** |

Perhatikan bahwa `prettier --check .` berjalan dengan cwd `frontend/`, jadi jangkauannya `frontend/**` — bukan root repo. Markdown root tidak pernah masuk ke prettier, dan tidak ada tes yang membacanya. Konsekuensinya: untuk PR yang menyentuh `AGENTS.md`, `CONTEXT.md`, `docs/adr/**`, atau `docs/solutions/**`, **§12 dan CI sama-sama tidak membuktikan apa pun tentang file itu**. Verifikasinya hanya bisa manual.

### 3. Verifikasi manual setelah merge (dan setelah rebase)

Setelah merge ke main, jalankan pada commit hasil merge — bukan pada branch:

```bash
git checkout main && git pull
git show --stat HEAD                      # file apa saja yang berubah?
grep -c "^### Documented solutions" AGENTS.md   # harus 1
grep -n "^### " AGENTS.md                  # lima section, berurutan, tanpa pengulangan
```

Untuk memastikan cacatnya terbatas pada satu file, bandingkan seluruh wilayah yang tidak diperiksa penjaga terhadap versi sebelumnya — jangan hanya memeriksa file yang Anda duga salah. Pada kasus ini: `docs/solutions/developer-experience/definition-of-done-verification.md` tetap 130 baris dengan satu heading, dan 13 ADR utuh.

Perbaikannya adalah PR biasa, bukan rewrite history — `48d8c99` menghapus tepat 4 baris:

```
$ grep -c "^### Documented solutions" AGENTS.md
2
$ # setelah PR #19
$ grep -c "^### Documented solutions" AGENTS.md
1
```

### 4. Jangan percaya nomor di badan PR

Angka tes yang diklaim di badan PR adalah klaim, bukan pengukuran. Ambil angkanya dari commit yang benar-benar akan di-merge:

```bash
git fetch origin
git rev-parse HEAD origin/<branch>     # head lokal vs head remote
git log --oneline origin/<branch>..HEAD  # commit yang belum ter-push
```

## Why This Matters

Learning sebelumnya di repo ini, `docs/solutions/developer-experience/definition-of-done-verification.md`, mengatakan: jalankan **seluruh** §12 dan baca **exit code**, bukan teks output, karena output yang terpotong bisa membuat kegagalan terlihat seperti lulus. Learning itu tentang **salah membaca sinyal yang ada** — sinyalnya benar-benar ada (`prettier --check` memang gagal dengan exit 1), hanya dibaca keliru.

Learning ini satu tingkat di atasnya: **sinyalnya tidak ada sama sekali**, sehingga tidak ada yang bisa dibaca dengan benar. Duplikat di `AGENTS.md` tidak menghasilkan exit code non-nol di perintah mana pun. Tidak ada urutan perintah §12 yang lebih teliti, tidak ada pembacaan exit code yang lebih disiplin, dan tidak ada pemeriksaan CI yang lebih ketat yang bisa menemukannya. Yang bisa menemukannya hanya satu hal: membaca isi file itu di pohon pasca-merge. Kedua learning saling melengkapi — yang pertama menutup "sinyal ada tetapi dibaca salah", yang kedua menutup "sinyal tidak ada". Menggabungkannya menjadi satu dokumen akan menyembunyikan perbedaan itu, jadi keduanya berdiri sendiri; overlap-nya hanya pada premis bersama "hijau bukan bukti".

Perlu ditegaskan juga: learning ini **memperluas**, bukan menggantikan, learning pertama. Tabel kelas cacat di dokumen itu menyebut `pnpm lint` menangkap "format + aturan lint" tanpa kualifikasi jangkauan; kenyataannya prettier berjalan dari `frontend/` dan tidak pernah melihat markdown root. Jadi baris tabel itu benar untuk kode frontend dan menyesatkan bila dibaca sebagai cakupan repo-wide.

**Bentuk yang sama muncul lebih awal di sesi yang sama, pada PR #16 (stale PR).** Saat memeriksa remote untuk memilih base branch, terlihat bahwa head branch remote PR #16 adalah `db150c8` (1 commit, 37 file) sementara branch lokal punya 7 commit yang belum di-push dengan head `59f04ab`. CI melaporkan 3/3 hijau — hijau untuk `db150c8`, bukan untuk kode yang akan di-merge. Badan PR mendeskripsikan perbaikan yang tidak ada di PR itu, dan mengklaim celah e2e yang sebenarnya sudah ditutup secara lokal. Diff `db150c8..59f04ab` adalah 17 file, +413/-52, termasuk `frontend/src/lib/domains/produk/schemas/produk.schema.ts` — tempat dua bug yang diklaim sudah diperbaiki itu berada. Terverifikasi lewat hitungan: `go test` 175 diklaim vs 177 aktual, `pnpm test` 131 vs 144, e2e 15 (9 produk) vs 17 (11 produk). Diperbaiki dengan push 7 commit tersebut dan menulis ulang badan PR memakai angka yang diukur pada head yang sebenarnya.

**Bentuk yang menghubungkan keduanya: "hijau yang tidak membuktikan apa pun tentang objek yang Anda maksud."** Di #16 subjek sinyalnya adalah commit yang berbeda dari commit yang di-merge; di #18/#19 subjek sinyalnya adalah himpunan file yang tidak mencakup file yang cacat. Dalam kedua kasus tidak ada kegagalan yang perlu dibaca — hanya jarak antara "apa yang diperiksa" dan "apa yang saya pikir diperiksa". Dan dalam kedua kasus biayanya nyata: #16 hampir merge kode yang badan PR-nya berbohong tentang isinya, #18 menaruh cacat di file yang dibaca setiap agen di setiap sesi.

Untuk repo ini khususnya, biayanya asimetris: cacat di `AGENTS.md` bukan cacat kosmetik. File itu adalah instruksi yang dibaca agen sebelum bekerja **dan** salah satu input korpus review dua-sumbu, jadi duplikat di dalamnya berarti setiap sesi berikutnya membaca konfigurasi yang berbeda dari yang dimaksud — tanpa satu pun perintah yang bisa memberitahu.

## When to Apply

- Sebelum menumpuk (stack) satu PR di atas branch PR lain, terutama bila tujuannya menghindari konflik pada file teks (`AGENTS.md`, `CONTEXT.md`, ADR). Periksa dulu apakah stacking benar-benar memberi CI di repo ini; di sini tidak.
- Sebelum meng-merge PR yang ditumpuk, setelah PR induknya masuk main — rebase dulu ke base yang baru, atau lewati stacking sama sekali.
- Setelah merge apa pun yang menyentuh file di luar `frontend/**` dan `backend/**` (`AGENTS.md`, `README.md`, `CONTEXT.md`, `docs/**`, `.github/**`), karena §12 dan CI tidak memeriksa wilayah itu.
- Kapan pun sebuah PR hijau dan Anda hendak menyimpulkan "aman", tanpa bisa menyebut perintah mana yang membaca file yang diubah.
- Kapan pun Anda mengutip angka tes atau status CI: pastikan dulu SHA yang diukur sama dengan SHA yang akan di-merge (`git rev-parse HEAD origin/<branch>`).
- Kapan pun Anda membuat prediksi tentang perilaku merge/diff/tooling dan prediksi itu belum diverifikasi — catat prediksinya sebelum menjalankannya.

## Examples

### Duplikat di `AGENTS.md` — sebelum dan sesudah

Sebelum PR #19, hasil merge #18 di main berisi sisipan yang sama dua kali (byte-identik):

```markdown
15  ### Documented solutions
16
17  `docs/solutions/` — catatan solusi masalah yang sudah lewat (bug, praktik, pola alur kerja), diorganisir per kategori dengan YAML frontmatter (`module`, `tags`, `problem_type`). Relevan saat mengimplementasi atau men-debug di area yang sudah terdokumentasi.
18
19  ### Documented solutions
20
21  `docs/solutions/` — catatan solusi masalah yang sudah lewat (bug, praktik, pola alur kerja), diorganisir per kategori dengan YAML frontmatter (`module`, `tags`, `problem_type`). Relevan saat mengimplementasi atau men-debug di area yang sudah terdokumentasi.
```

Setelah PR #19 (`48d8c99`, menghapus tepat 4 baris):

```markdown
15  ### Documented solutions
16
17  `docs/solutions/` — catatan solusi masalah yang sudah lewat (bug, praktik, pola alur kerja), diorganisir per kategori dengan YAML frontmatter (`module`, `tags`, `problem_type`). Relevan saat mengimplementasi atau men-debug di area yang sudah terdokumentasi.
18
19  ### Code organization
```

Buktinya:

```
$ grep -c "^### Documented solutions" AGENTS.md
2          # sebelum PR #19

$ grep -c "^### Documented solutions" AGENTS.md
1          # sesudah PR #19

$ grep -n "^### " AGENTS.md
15:### Documented solutions
19:### Code organization
...        # lima section, berurutan, tanpa pengulangan
```

Pemeriksaan cakupan (bahwa hanya `AGENTS.md` yang rusak) dilakukan dengan membandingkan seluruh wilayah yang tidak diperiksa penjaga terhadap versi sebelumnya — `docs/solutions/developer-experience/definition-of-done-verification.md` tetap 130 baris dengan satu heading, dan 13 ADR di `docs/adr/` utuh. Tidak ada output perintah persis yang saya simpan dari langkah itu; yang tersimpan adalah hasilnya (hanya `AGENTS.md` yang berdampak). Bila mengulang kasus serupa, simpan outputnya ke file agar angkanya bisa dikutip ulang.

### Mekanisme squash merge — mengapa diff dihitung ulang

```
16708d2  Autentikasi: login, sesi, dan peran Kasir/Admin (#15)   <- merge base lama
   |
   +-- 98f2019  Catat pelajaran verifikasi DoD §12 ... (#17)     <- menambah section di AGENTS.md
   |      |
   |      +-- cc2650d  (branch #18, memuat commit #17 sebagai ancestor)
   |             |
   |             +-- 2fa3409  Tetapkan glosarium tunggal ... (#18)
```

Squash merge #18 menghitung `diff(16708d2..cc2650d)` — dan karena branch #18 **memuat** commit #17, diff itu sudah berisi sisipan `AGENTS.md` milik #17. Diff tersebut diterapkan di atas main yang **sudah** berisi sisipan yang sama (karena #17 sudah di-squash-merge sebagai `98f2019`). Sisipan diterapkan dua kali. Tidak ada langkah di rantai itu yang membandingkan "sisi ini dan sisi itu identik", jadi tidak ada yang bisa membuang salah satunya.

Prediksi yang saya buat sebelum merge adalah sebaliknya — bahwa merge akan mengenali sisipan identik di kedua sisi sebagai satu perubahan. Prediksi itu salah untuk **sisipan**; ia berlaku untuk kasus di mana kedua sisi sudah sama-sama memuat hasil akhirnya, bukan untuk dua diff yang menambahkan baris yang sama. Prediksi ini dicatat di sini justru karena bagian yang paling berguna dari learning ini adalah prediksinya yang salah, bukan langkah perbaikannya.

### Checklist membaca pohon pasca-merge

```bash
git checkout main && git pull
git show --stat HEAD                              # 1. file apa yang sebenarnya berubah?
grep -c "^### Documented solutions" AGENTS.md     # 2. hitung section yang bisa terduplikasi
grep -n "^### " AGENTS.md                         # 3. urutan + tidak ada pengulangan
wc -l docs/solutions/developer-experience/definition-of-done-verification.md   # 4. 130, satu heading
ls docs/adr/ | wc -l                              # 5. 13 ADR utuh
```

Untuk PR yang mungkin stale:

```bash
git fetch origin
git rev-parse HEAD origin/<branch>                # head lokal vs head remote — sama?
git log --oneline origin/<branch>..HEAD          # commit yang belum ter-push
git diff origin/<branch>...HEAD --stat           # apa yang tidak diperiksa CI?
```

Dan untuk memutuskan apakah sebuah PR butuh pembacaan manual sama sekali:

```bash
# perintah §12 mana yang membaca file yang saya ubah?
git diff --name-only origin/main...HEAD
# frontend/**  -> pnpm lint/check/test/build (cwd frontend/)
# backend/**   -> gofmt/go vet/go build/go test (cwd backend/)
# lainnya      -> TIDAK ADA. Baca manual setelah merge.
```

## Related

- `docs/solutions/developer-experience/definition-of-done-verification.md` — learning pertama repo ini; tentang salah **membaca** sinyal yang ada (exit code vs teks output). Dokumen ini adalah kelanjutannya, bukan penggantinya: sinyal yang **tidak ada**. Tabel kelas cacat di sana menyebut `pnpm lint` tanpa kualifikasi jangkauan, dan dokumen ini yang melengkapinya.
- `docs/adr/0008-ci-github-actions.md` — satu workflow, satu job per sisi monorepo; kriteria selesainya ("semua job hijau di push/pull_request") adalah kriteria yang secara literal terpenuhi di main sementara duplikatnya ada. Pemicu `pull_request: branches: [main]` di ADR ini juga yang membuat PR bertumpuk kehilangan CI.
- `docs/adr/0009-e2e-playwright-ci.md` — job `e2e` di CI dan mengapa ia menjalankan proses sungguhan.
- `docs/adr/0007-testing-strategy.md` — piramida per sisi; tidak ada lapisannya yang menyentuh file non-kode.
- `docs/adr/0013-single-glossary-context-md.md` — glosarium tunggal di `CONTEXT.md`; ADR ini juga yang mencatat sinyal tinjau ulang seputar `CONCEPTS.md` yang diuji oleh run penulisan dokumen ini.
- `.github/workflows/ci.yml` — pemicu `push: branches: [main]` dan `pull_request: branches: [main]`, serta `working-directory: frontend` pada job frontend.
- `frontend/package.json` — `"lint": "prettier --check . && eslint ."`, dijalankan dari `frontend/`; inilah alasan `.` tidak berarti root repo.
- `frontend/.prettierignore` — satu-satunya prettierignore di repo, dan ia frontend-relatif; tidak ada prettierignore di root.
- `AGENTS.md` — file yang menjadi korban sekaligus konfigurasi yang menentukan ke mana learning ini ditulis.
- PR #19 (`48d8c99`) — perbaikan: menghapus 4 baris duplikat di `AGENTS.md`.
- PR #18 (`2fa3409`) — PR yang ditumpuk; squash merge-nya menggandakan sisipan #17.
- PR #17 (`98f2019`) — PR induk yang ditumpuk; menambah section `### Documented solutions` di `AGENTS.md`.
- PR #16 (`5305ffa`) — stale PR: CI hijau untuk `db150c8`, sementara head sebenarnya `59f04ab`.
