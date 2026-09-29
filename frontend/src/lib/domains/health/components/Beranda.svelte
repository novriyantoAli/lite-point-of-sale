<script lang="ts">
	import {
		CATATAN,
		JUDUL,
		MODUL,
		MODUL_JUDUL,
		MODUL_KEPALA,
		MODUL_TEKS,
		PAPAN,
		STRIP
	} from '$lib/components/shared/mosaik';
	import { cn } from '$lib/utils';
	import HealthStatus from './HealthStatus.svelte';

	/**
	 * Beranda: layar pertama setelah Masuk. Ia menyusun dua modul — Status layanan
	 * (dari domain health) dan Dari satu layar (kalimat yang menunjuk rel navigasi)
	 * — di bawah satu strip judul. Susunannya hidup di sini, bukan di route, supaya
	 * route tetap tipis: hanya menaruh komponen di sebuah URL (ADR-0006).
	 */
	/**
	 * Jarak 8px sebelum modul Dari satu layar bukan kelalaian: prototipe layar ini
	 * sendiri yang memisahkan dua modul dengan `margin-top: 8px`, jadi keduanya
	 * memang tidak bersebelahan.
	 */
	const MODUL_KEDUA = cn(MODUL, 'mt-2');
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
