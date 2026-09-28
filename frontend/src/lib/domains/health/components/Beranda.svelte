<script lang="ts">
	import HealthStatus from './HealthStatus.svelte';

	/**
	 * Beranda: layar pertama setelah Masuk. Ia menyusun dua modul — Status layanan
	 * (dari domain health) dan Dari satu layar (kalimat yang menunjuk rel navigasi)
	 * — di bawah satu strip judul. Susunannya hidup di sini, bukan di route, supaya
	 * route tetap tipis: hanya menaruh komponen di sebuah URL (ADR-0006).
	 */
	/**
	 * Papan mosaik (DESIGN.md, Layout): satu kolom selebar papan — prototipe
	 * Beranda memakai `board--full`, bukan papan tiga kolom milik Kasir. Padding
	 * luarnya milik rel (`(app)/+layout`), jadi di sini tidak ada padding lagi.
	 */
	const PAPAN = 'grid grid-cols-1 gap-0';
	/**
	 * Strip judul: satu-satunya tempat ukuran 24px muncul di layar ini. `-mb-px`
	 * plus `z-[2]` menariknya turun satu piksel, jadi garis tintanya yang menutup
	 * modul di bawahnya alih-alih bertumpuk dengan garis rambut (DESIGN.md,
	 * Shared-Hairline).
	 */
	const STRIP =
		'z-[2] col-span-full -mb-px flex flex-wrap items-baseline gap-x-3 gap-y-1 border border-border border-b-foreground bg-card px-2 py-2';
	const JUDUL = 'text-2xl leading-none font-bold tracking-[-0.015em]';
	/** Seluruh teks 12px; lantai huruf dunia ini (DESIGN.md, Legibility Floor). */
	const CATATAN = 'text-xs';

	/**
	 * Modul: petak dengan garis rambut, dan hanya garis atasnya yang digambar —
	 * dua modul bersebelahan berbagi satu garis, bukan dua (DESIGN.md,
	 * Shared-Hairline).
	 */
	const MODUL = 'border border-border border-b-0 bg-card';
	/**
	 * Jarak 8px sebelum modul Dari satu layar bukan kelalaian: prototipe layar ini
	 * sendiri yang memisahkan dua modul dengan `margin-top: 8px`, jadi keduanya
	 * memang tidak bersebelahan.
	 */
	const MODUL_KEDUA = `${MODUL} mt-2`;
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/** Satu baris isi di dalam modul; garis bawahnya menutup modulnya. */
	const MODUL_TEKS = 'border-b border-border px-2 py-1.5 text-xs';
</script>

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya
		(DESIGN.md, Typography).
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Beranda</h1>
		<p class={CATATAN}>Kasir untuk satu toko, satu terminal.</p>
	</div>

	<!--
		Modul pertama adalah Status layanan; domain health yang menggambarnya
		(`HealthStatus`), dan Beranda menulis kelas modulnya sendiri seperti setiap
		layar lain menulisnya — yang tidak diduplikasi adalah bentuk statusnya, bukan
		bungkusnya.
	-->
	<HealthStatus />

	<!--
		Modul Dari satu layar: satu kalimat yang menyebut setiap tab di rel atas,
		dengan nama tabnya sebagai Strong di dalam teks (DESIGN.md, Typography).
	-->
	<section class={MODUL_KEDUA} aria-labelledby="mulai-judul">
		<div class={MODUL_KEPALA}>
			<h2 id="mulai-judul" class={MODUL_JUDUL}>Dari satu layar</h2>
			<span class={CATATAN}>tanpa menu tersembunyi</span>
		</div>
		<p class={MODUL_TEKS}>
			Semua pekerjaan toko ada di rel atas: <strong>Kasir</strong> untuk melayani pembeli,
			<strong>Produk</strong> dan <strong>Stok</strong> untuk barang, <strong>Laporan</strong> untuk
			omzet harian, lalu <strong>Pengguna</strong>, <strong>Pengaturan</strong>, dan
			<strong>Backup</strong>.
		</p>
	</section>
</div>
