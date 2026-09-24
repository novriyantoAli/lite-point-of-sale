# Struk: dicetak setelah Penjualan tersimpan, dari template Pengaturan

#8 menambahkan **cetak Struk ESC/POS**, **template Struk yang diatur Admin**, dan **cetak ulang dari Penjualan lama** (ADR-0003). Ini slice pertama yang menyentuh **perangkat keras**, dan yang pertama memperkenalkan **Pengaturan** sebagai tempat menyimpan setelan toko — ADR-0014 sudah menjanjikan ambang **Stok menipis** pindah ke sana saat layar itu mendarat.

Enam keputusan diambil sebelum satu baris kode ditulis, dan dicatat di sini karena masing-masing sulit dibalik, mengejutkan tanpa konteks, atau keduanya.

## Keputusan

### 1. Cetak tidak pernah menggagalkan Penjualan

`CetakStruk` berjalan **setelah** checkout selesai, bukan di dalam transaksinya. Kalau printer tidak ada atau gagal, Penjualan tetap tersimpan, dan **hasil cetaknya ikut di jawaban checkout** — Kasir harus tahu, bukan menemukannya di log.

| Opsi | Perilaku | Kenapa ditolak |
| --- | --- | --- |
| Cetak di dalam transaksi | Gagal cetak → 500, Penjualan di-rollback | Uang sudah berpindah tapi Penjualannya hilang; Kasir harus mengulang seluruh penjualan — risiko double-sale atau penjualan tak tercatat |
| **Cetak setelah simpan, hasil di jawaban** | Penjualan selalu tersimpan; `printed: true/false` + alasan | — (dipilih) |
| Cetak async tanpa hasil | Kasir tidak tahu Struk keluar atau tidak | Butuh UI pemantau sendiri; kompleksitas tanpa pemakai di MVP |
| Tidak ada cetak otomatis | Hanya tombol cetak di detail | Bertentangan dengan AC #8 |

Dua kegagalan itu tidak sederajat: Penjualan adalah uang yang sudah berpindah dan harus tercatat apa pun yang terjadi; Struk adalah bukti yang bisa diterbitkan ulang — justru itulah kenapa cetak ulang ada di slice yang sama. Kegagalan cetak **tidak boleh senyap**: layar menampilkan pesannya beserta tombol mengulang, bukan toast yang lewat.

### 2. Transport: path device tanpa default, dan printer null yang jujur

`POS_PRINTER_DEVICE` menunjuk path device (ditulis dengan `os.OpenFile`), dan **tidak ada nilai default** — `/dev/usb/lp0` hanya benar di Linux, dan tebakan yang salah lebih buruk daripada keadaan "belum diatur" yang jujur. Kosong berarti **printer null**: adapter yang menjawab "printer belum diatur" sebagai **kegagalan cetak**, bukan crash dan bukan sukses senyap. Mesin dev dan CI menyala tanpa setup, dan jalur gagal keputusan 1 benar-benar terlatih.

Ditolak: shell ke `lp`/CUPS (dependensi eksternal di terminal toko; ADR-0003 memilih Go mengirim byte sendiri), TCP `:9100` (ADR-0003 menyebut USB), dan menulis byte ke file untuk spooler (Struk tidak keluar saat checkout).

Port `Printer` berdiri di samping domain (`Print(ctx, lines) error`), di-wire di `app.New` — pola yang sama dengan `ProductRepository` — dan fake-nya dipakai tes `usecase` (ADR-0007). **Encoder murni**: domain menghasilkan **baris teks** (mana yang dicetak, urutannya, pembungkusan pada lebar kertas), adapter ESC/POS menerjemahkan baris → byte. Tes domain memeriksa *apa yang dikatakan Struk*; tes adapter memeriksa byte untuk baris yang diketahui — bukan sebaliknya, yaitu tes aturan yang mengunci urutan byte ajaib.

### 3. Pengaturan: satu baris bertipe, dan ambang menipis pindah ke sana

Migrasi `0005` membuat **satu baris** `pengaturan(id = 1, header, footer, paper_width, low_stock_threshold)`, disemai dengan nilai hari ini (ambang 5, lebar 80 mm, header/footer kosong). ADR-0002 (satu toko, satu terminal) membuat satu baris itu jujur; kolom bertipe menjaga jalur baca tetap bertipe. Ditolak: kantong key-value (mengundang UI generik dan pembacaan tanpa tipe), dan memperluas `app_meta` — tabel itu pembukuan infrastruktur (`schema_version`), bukan data domain.

**`domainproduk.LowStockThreshold` dihapus**, nilainya pindah ke baris semai: satu sumber kebenaran, dan tidak ada keadaan "belum ada baris" yang harus ditangani di Go. Tiga hal yang mengikutinya:

- `usecaseproduk.LowStock` menjawab **laporannya**, bukan cuma daftarnya — `{Threshold, Products}` — supaya handler tetap tidak memegang aturan dan tidak membaca setting sendiri.
- `usecaseproduk` mendeklarasikan port satu-method `LowStockSettings` **di package-nya sendiri**, bukan mengimpor domain `pengaturan`; dependensi tetap mengarah ke dalam (ADR-0004).
- **`GET /api/produk/stok-menipis` dan layar `/stok` tidak berubah.** Endpoint itu sudah menjawab `threshold` bersama daftarnya, dan `/stok` merender angka yang dijawab API — yang berubah hanya dari mana angka 5 itu datang.

### 4. Template: blok teks bebas, lebar kertas dalam mm

- **`header` dan `footer` adalah blok teks bebas** (baris dipisah newline, diedit lewat textarea), bukan field `nama_toko`/`alamat`/`telepon`. Struk adalah artefak cetak dan `CONTEXT.md` menyebut "blok sederhana"; field terstruktur akan mengarang istilah domain yang tidak ada di glosarium **sekaligus** memaksa tata letak yang tidak dipilih Admin.
- **`paper_width` disimpan sebagai mm (58 atau 80), divalidasi.** Jumlah kolom (58 mm → 32, 80 mm → 48) adalah **turunan** yang dihitung encoder, bukan nilai kedua yang bisa menyimpang. Istilah domain dan label UI tetap "58/80 mm".
- **Isi Struk mengikuti daftar `CONTEXT.md`**: blok header → Nomor Struk → waktu → Kasir → Item (nama, `qty × harga`, subtotal) → total → metode → jumlah bayar + Kembalian (**hanya Tunai**) → blok footer.
- **Baris panjang dibungkus pada lebar kertas, tidak dipotong.** Nama Produk yang panjang kehilangan informasi kalau dipotong; Struk yang salah lebih buruk daripada Struk yang lebih tinggi.

Ditolak: pratinjau/editor visual Struk di layar Pengaturan (ADR-0003 menundanya; memperluas slice ke arah editor tata letak).

### 5. Satu usecase cetak, dan template saat ini untuk cetak ulang

`CetakStruk(nomorStruk)` dipakai **dua jalur**: otomatis setelah checkout, dan lewat `POST /api/penjualan/{nomorStruk}/struk` untuk cetak ulang — sehingga bentuk jawaban "hasil cetak" identik di keduanya, dan tombol ulang di panel setelah checkout memakai endpoint yang sama. Sub-resource-nya `/struk` (benda yang dihasilkan), sejajar dengan `POST /api/produk/{id}/stok`. Rutenya **tanpa role guard**, seperti `GET` Penjualan: `CONTEXT.md` memberi **Kasir** "cetak Struk", dan kedua Peran menjual di toko satu terminal.

Cetak ulang memakai **template Pengaturan saat ini**, bukan salinan saat Penjualan terjadi. Data Penjualan tetap snapshot seperti sekarang (Item, harga, total, nama Kasir, waktu, Nomor Struk — ADR-0015); yang dibaca saat cetak hanyalah header, footer, dan lebar kertas. Struk adalah dokumen **toko**, bukan salinan letterhead historis, dan menyimpan salinan template per Penjualan menambah kolom yang tidak dibaca siapa pun.

Pintu masuk cetak ulang adalah **layar `/penjualan`** yang kecil: ketik Nomor Struk → tampilkan Penjualannya → tombol cetak. Ia bisa diakses kedua Peran, dan #9 (laporan) menumbuhkan layar ini atau menambah `/laporan` Admin-only untuk omzet. Ditolak: menunggu daftar Penjualan #9 (AC #8 "cetak ulang dari Penjualan lama" jadi tidak terpenuhi oleh #8, dan Kasir kehilangan jalan cetak ulang).

### 6. Dipecah tiga issue, dengan A dan B paralel

| | Issue | Bergantung pada |
| --- | --- | --- |
| **A** | Pengaturan: template Struk & ambang Stok menipis (migrasi 0005, domain, layar `/pengaturan`, pindahnya ambang) | — |
| **B** | Penjualan: baca satu Penjualan lewat Nomor Struk (proxy BFF `GET`, api, query detail, layar `/penjualan`) | — |
| **C** | Struk: cetak ESC/POS & cetak ulang (domain `struk`, port + adapter printer, encoder, `CetakStruk`, hasil cetak di jawaban checkout, tombol di panel & `/penjualan`) | **A dan B** |

A dan B tidak saling menyentuh — A menyentuh `produk` dan slice baru, B hanya menambah pembaca di slice `penjualan` — jadi keduanya bisa jalan bersamaan. C menyentuh keduanya **dan** bagian paling berisiko (perangkat), sehingga lebih baik mendarat setelah dua potongan yang mudah dipastikan.

## Considered Options

Selain yang sudah disebut di tiap keputusan:

- **Menyimpan salinan template pada setiap Penjualan** — ditolak: kolom yang tidak dibaca siapa pun, dan Struk adalah dokumen toko, bukan letterhead historis (keputusan 5).
- **Mencetak dari browser** (`window.print`) — ditolak: ADR-0003 memilih Go mengirim ESC/POS ke printer thermal, bukan halaman browser.
- **Memakai kembali `StrukPenjualan.svelte` sebagai pratinjau Struk** — ditolak: memperluas #8 ke arah editor visual.
- **Ambang menipis tetap konstanta dan pindah nanti** — ditolak: ADR-0014 sudah menjanjikan pemindahannya, layar Pengaturan yang isinya cuma field Struk sebenarnya editor template yang memakai nama Pengaturan, dan menundanya berarti migrasi kedua menyentuh tabel yang sama.

## Consequences

- **Reprint memakai nama toko saat ini.** Setelah toko mengganti namanya, Struk yang dicetak ulang menampilkan nama baru; Nomor Struk yang mengidentifikasi struk aslinya. Konsekuensi yang diterima dari keputusan 5.
- **Tanpa `POS_PRINTER_DEVICE`, cetak selalu gagal** — dengan pesan yang bisa dibaca Kasir, bukan senyap. Deployment pertama harus menyetelnya.
- **`playwright.config.ts` menyetel `POS_PRINTER_DEVICE` ke file sementara** untuk proses Go-nya, sehingga e2e browser menguji **jalur sukses** dengan membaca byte yang tercetak dari file itu — memakai jalur device sungguhan, tanpa perangkat. Tanpa itu, jalur sukses tidak pernah berjalan di CI dan AC "tes memverifikasi konten Struk" hanya terbukti di tes encoder.
- **`CONTEXT.md` diperbarui saat A mendarat**: entri **Stok menipis** berhenti menyebut ambang sebagai konstanta, entri **Struk** menyebut blok template dan lebar kertas, dan **Pengaturan** menjadi istilahnya sendiri. Glosarium tidak mendahului kode.
- **Tanpa editor visual** (ADR-0003): template MVP adalah blok teks; toko yang ingin tata letak lain belum bisa.
- Tes yang menjaga keputusan ini: `backend/internal/domain/struk/*_test.go` (isi Struk: urutan baris, pembungkusan, Kembalian hanya Tunai, blok template), `backend/internal/adapter/escpos/*_test.go` (byte untuk baris yang diketahui; menulis ke file `t.TempDir()`), `backend/internal/usecase/penjualan/cetak_test.go` (cetak gagal tidak menggagalkan Penjualan, printer null melaporkan kegagalan), `backend/internal/usecase/produk/stok_test.go` (ambang dari setting, bukan konstanta), `backend/internal/adapter/sqlite/pengaturan_repository_test.go`, `backend/tests/e2e/{pengaturan,struk}_test.go` (seam REST: guard Admin, ambang tersimpan, hasil cetak di jawaban checkout, cetak ulang), dan `frontend/tests/e2e/*.spec.ts` (browser sungguhan: atur template, cetak ulang dari `/penjualan`, byte Struk terbaca dari file printer).
