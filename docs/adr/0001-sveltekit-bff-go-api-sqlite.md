# SvelteKit BFF + Go API + SQLite

Sistem kasir satu terminal ini memisahkan lapisan web dari API bisnis: **SvelteKit** (Node) melayani UI dan memegang login/session (cookie) sekaligus meneruskan panggilan ke Go sebagai backend-for-frontend; **Go** adalah API bisnis murni (Produk, Stok, Penjualan, laporan, cetak struk) yang mengakses **SQLite** dan mengirim perintah ESC/POS ke printer. Browser hanya bicara ke SvelteKit, tidak pernah langsung ke Go.

## Considered Options

- **Satu binary** — Go meng-embed frontend Svelte statis + SQLite, tanpa Node di produksi. Paling "ringan" (satu proses, satu file); ditolak karena tim memilih pola full-stack SvelteKit yang lebih familiar.
- **Browser langsung ke Go** — JWT + CORS, dua origin, duplikasi proteksi; ditolak demi satu origin via BFF.

## Consequences

- Di produksi ada **dua proses** (server SvelteKit + server Go), bukan satu binary.
- SQLite tetap dipegang Go (satu penulis data) — SvelteKit tidak menyentuh database langsung.
