# Cetak struk via ESC/POS dari Go, tanpa JVM/JRXML

Struk dicetak oleh **backend Go** sebagai perintah **ESC/POS mentah** ke printer thermal (USB), dengan **template blok sederhana** yang isinya dapat diatur (header, daftar item, total, footer; lebar kertas 58/80 mm). Editor visual (**JasperSoft Studio / JRXML**) sengaja **ditunda**; JVM tidak dipakai di MVP.

## Considered Options

- **JasperReports/JRXML** — editor visual matang (JasperSoft Studio), tapi butuh JVM di samping Go (tidak ada renderer JRXML yang matang untuk Go) dan berorientasi halaman, kurang pas untuk struk rol thermal.

## Consequences

- Template struk MVP berupa blok sederhana tanpa drag-drop bebas; editor visual diputuskan ulang di phase 2.
- Browser tidak mencetak langsung; frontend memanggil Go untuk mencetak.
