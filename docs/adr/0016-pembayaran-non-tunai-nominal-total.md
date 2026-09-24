# Pembayaran non-tunai: dicatat sebesar total, tanpa Kembalian, tanpa gateway

#6 menambahkan checkout, tetapi hanya menerima **Tunai** (ADR-0015). #7 membuka tiga metode sisanya — **QRIS**, **Debit**, **Transfer** — sebagai Pembayaran yang **dicatat**: tidak ada gateway yang dipanggil, tidak ada uang yang dipindahkan oleh aplikasi. Yang mudah dipertanyakan lagi nanti bukan "apakah dicatat", melainkan **berapa nominalnya** dan **siapa yang menentukannya**, jadi itu yang dicatat di sini.

## Keputusan

### 1. Nominal Pembayaran non-tunai adalah total Penjualan — tepat, bukan sekadar cukup

Satu Penjualan punya satu Pembayaran (CONTEXT.md, Pembayaran; split payment di luar scope PRD), dan hanya **Tunai** yang punya **Kembalian** (CONTEXT.md, Kembalian). Karena itu tidak ada tempat untuk selisih:

| Metode              | Nominal yang diterima            | Kembalian                        |
| ------------------- | -------------------------------- | -------------------------------- |
| Tunai               | ≥ total                          | `bayar − total`                  |
| QRIS/Debit/Transfer | **= total** (selain itu ditolak) | selalu `0`                       |

Menerima `nominal > total` untuk non-tunai lalu memaksa Kembaliannya `0` ditolak: barisnya akan menyimpan Pembayaran yang nominalnya **tidak sama dengan Penjualan yang dibayarnya**, dan selisihnya tidak muncul di mana pun — tidak di Kembalian (yang nol), tidak di laporan, tidak di Struk. Menerima `nominal < total` juga ditolak: itu Pembayaran sebagian, dan satu Penjualan tetap satu Pembayaran penuh.

Aturan ini hidup di **usecase**, tempat aturan checkout yang lain sudah tinggal (ADR-0015):

```go
var change int64
if method == domainpenjualan.PaymentCash {
    if input.Payment.Amount < total { … }   // "Jumlah bayar kurang dari total Penjualan."
    change = input.Payment.Amount - total
} else if input.Payment.Amount != total {
    …                                       // "Pembayaran non-tunai harus sebesar total Penjualan: %d."
}
```

### 2. Kasir memilih metode; nominal non-tunai diisi layar, bukan diketik

Form Pembayaran menawarkan keempat metode sebagai satu pilihan radio (Tunai default — metode yang butuh angka diketik, dan yang paling sering dipakai). Untuk Tunai, form menampilkan field **Jumlah bayar** dan Kembalian yang terus dihitung. Untuk metode tercatat, keduanya **tidak dirender**: yang tampil hanya "Dibayar · QRIS Rp 36.000", dan layar mengirim `amount = total` — angka yang sudah ia tampilkan.

Alternatif yang ditolak: field "nominal" untuk non-tunai yang harus diketik Kasir sebesar total. Itu meminta Kasir mengetik ulang angka yang sudah ada di layar, dengan satu-satunya hasil yang benar adalah angka itu — dan setiap salah ketik menjadi 400 dari API, bukan koreksi di tempat.

### 3. Satu bentuk body untuk keempat metode; API yang memutuskan

`POST /api/penjualan` tetap menerima `{"payment":{"method":…,"amount":…}}` untuk semua metode. `amount` tidak dibuat opsional untuk non-tunai: Pembayaran yang tersimpan tetap butuh nominalnya, dan dua bentuk body berarti dua jalur validasi. Yang berubah dari #6 hanya **apa yang diterima**: `paymentMethod` di usecase kini menerima keempat metode yang dikenal domain, dan menolak yang tidak dikenal (`400 invalid_input`); `CheckoutInputSchema` di frontend memakai `MetodePembayaranSchema`, bukan lagi `z.literal('cash')`.

`METODE_URUT` di `penjualan.schema.ts` menjadi satu sumber untuk urutan tampil **dan** keanggotaan (`z.enum(METODE_URUT)`), supaya daftar yang dirender layar tidak bisa menyimpang dari yang diterima schema.

### 4. Tidak ada migrasi: kolomnya sudah menerima keempatnya

`0004_penjualan.sql` menulis `method TEXT NOT NULL CHECK (method IN ('cash','qris','debit','transfer'))` dan `domainpenjualan` sudah mendeklarasikan keempat konstanta sejak #6, dengan komentar yang menyebut #7 sebagai alasan. Taruhan desain itu yang membuat #7 tidak menyentuh skema sama sekali.

## Considered Options

- **Non-tunai menerima nominal berapa pun ≥ total, Kembalian dipaksa 0** — ditolak: barisnya menyimpan nominal yang tidak sama dengan Penjualan yang dibayarnya, dan selisihnya hilang tanpa jejak.
- **Non-tunai menerima nominal < total sebagai pembayaran sebagian** — ditolak: satu Penjualan satu Pembayaran; split payment di luar scope PRD.
- **API mengabaikan `amount` untuk non-tunai dan memakai total sendiri** — ditolak: klien yang mengirim angka salah tidak akan pernah tahu, dan kontraknya berbohong (field yang ada tapi tidak dibaca). Menolak selisih memberi pesan yang bisa dibaca Kasir.
- **Field "nominal" untuk non-tunai di form** — ditolak: meminta Kasir mengetik ulang total yang sudah tampil, dengan satu jawaban benar.
- **`amount` opsional untuk non-tunai** — ditolak: dua bentuk body, dua jalur validasi, dan Pembayaran tetap butuh nominal saat disimpan.
- **Memproses QRIS lewat gateway** — ditolak: di luar scope PRD; non-tunai hanya dicatat.
- **Memisahkan Pembayaran ke domain/route sendiri** — ditolak: Pembayaran selalu satu dengan Penjualannya dan tidak pernah dibaca tanpa Penjualannya (sama seperti alasan ADR-0015 memilih satu slice `penjualan`).

## Consequences

- **Struk tidak menampilkan baris Kembalian untuk non-tunai** (CONTEXT.md, Struk: jumlah bayar & Kembalian hanya "bila Tunai"). Form dan Struk sama-sama bertanya lewat `punyaKembalian` di `penjualan.schema.ts`, bukan membandingkan sendiri dengan literal `cash` — satu aturan, satu jawaban.
- **Non-tunai yang totalnya basi tidak bisa dibetulkan dari form.** Layar mengirim total yang ia hitung dari katalog yang dibacanya, dan API menolaknya bila harga sudah berubah; pesannya menyebut total yang benar, tetapi tidak ada field untuk memperbaikinya — Kasir harus membuang lalu menambahkan Itemnya lagi. Ini konsekuensi yang diterima dari keputusan 1, bukan yang terlewat: satu terminal dan satu Pengguna (ADR-0002) membuat harga yang berubah di tengah checkout menjadi balapan yang jarang, dan "nominal bebas" yang menutupinya justru yang ditolak di atas.
- Kembalian yang tersimpan untuk non-tunai **selalu 0**, jadi laporan (#9) boleh menjumlahkan `change` tanpa memfilter metode — tetapi tidak boleh menampilkan Kembalian untuk metode tercatat.
- Kalau nanti split payment atau pembayaran sebagian masuk, **keputusan 1 inilah yang berubah lebih dulu**: `amount = total` adalah aturan yang menutup pintu itu.
- Tidak ada editor template, printer, maupun gateway di sini — itu #8. Yang mendarat di #7 adalah datanya: metode dan nominal yang benar-benar tercatat.
- Tes yang menjaga keputusan ini: `backend/internal/usecase/penjualan/checkout_test.go` (tiap metode tercatat dengan Kembalian 0; nominal ≠ total ditolak; metode tak dikenal ditolak), `backend/tests/e2e/penjualan_test.go` (seam REST: tiap metode lewat HTTP lalu dibaca ulang dengan `GET /api/penjualan/{nomorStruk}`, Stok tetap keluar, nominal meleset & metode tak dikenal → 400), `frontend/src/lib/domains/penjualan/schemas/penjualan.schema.test.ts`, `frontend/src/lib/domains/penjualan/api/penjualan.api.test.ts`, `frontend/src/lib/domains/penjualan/components/Kasir.svelte.test.ts`, dan `frontend/tests/e2e/penjualan.spec.ts` (browser sungguhan: jual dengan QRIS/Debit/Transfer, Struk menyebut metodenya dan tidak menampilkan Kembalian).
