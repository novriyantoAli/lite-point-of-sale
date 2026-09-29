<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { createPengaturanQuery } from '$lib/domains/pengaturan';
	import { pencarianState } from '../state/pencarian.state.svelte';

	/**
	 * Kolom kiri kasir: dua field yang menyaring mosaik katalog di kolom tengah,
	 * plus satu baris yang menyebut ambang Stok menipis.
	 *
	 * Dua field, bukan satu, karena API menyaring listing dengan `code` *dan*
	 * `name` sekaligus — satu field harus mengirim keduanya dan menuntut sebuah
	 * Produk cocok di keduanya. Dua field juga yang diminta AC: ditemukan lewat
	 * Kode, dan ditemukan lewat nama.
	 *
	 * Saringannya komponen state milik domain penjualan, bukan `produkFilterState`
	 * milik domain produk: itu milik layar katalog, dan mencapai `state/` domain
	 * lain persis yang dilarang aturan barrel (ADR-0006).
	 */

	/**
	 * Ambang Stok menipis dibaca dari Pengaturan lewat barrel-nya (ADR-0006): kata
	 * "menipis" tercetak di petak katalog, dan pemakainya belum pernah memakai
	 * sistem kasir — angka yang membuat kata itu berlaku harus bisa dibacanya di
	 * layar yang sama (PRODUCT.md, Prinsip 1).
	 */
	const pengaturan = createPengaturanQuery();
</script>

<section class="flex flex-col" aria-labelledby="kasir-cari-judul">
	<div
		class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
	>
		<h2 id="kasir-cari-judul" class="text-[13px] font-bold tracking-[0.01em]">Cari Produk</h2>
		<span class="text-xs">Kasir</span>
	</div>

	<!--
		Petak pertama kolom, jadi ia yang membawa garis atas: modul di bawahnya
		berbagi satu garis rambut dengan petak di atasnya, tidak menggambar sendiri.
	-->
	<div class="border-b border-border px-2 py-1.5">
		<div class="flex flex-col gap-[3px]">
			<Label for="kasir-kode" class="text-xs font-semibold">Kode</Label>
			<!-- autofocus: a barcode scanner types into whatever has focus, and the
			     Kode field is what narrows the listing to the scanned Produk. -->
			<Input
				id="kasir-kode"
				name="code"
				autocomplete="off"
				autofocus
				class="h-[26px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]"
				bind:value={pencarianState.code}
			/>
		</div>
		<p class="mt-1 text-xs">Scan barcode atau ketik Kode.</p>
	</div>

	<div class="border-b border-border px-2 py-1.5">
		<div class="flex flex-col gap-[3px]">
			<Label for="kasir-nama" class="text-xs font-semibold">Nama</Label>
			<Input
				id="kasir-nama"
				name="name"
				autocomplete="off"
				class="h-[26px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]"
				bind:value={pencarianState.name}
			/>
		</div>
		<p class="mt-1 text-xs">Produk tanpa Kode dicari lewat namanya.</p>
	</div>

	<!--
		Angka yang belum bisa dihitung ditulis "—", bukan 0 (DESIGN.md, Figures):
		sebelum Pengaturan datang, ambangnya belum diketahui, dan menulis 0 akan
		menandai setiap Produk sebagai menipis.
	-->
	<div class="border-b border-border px-2 py-1.5">
		<p class="text-xs">
			Ambang Stok menipis
			<span class="font-semibold tabular-nums">
				{#if pengaturan.isPending || pengaturan.error}—{:else}{pengaturan.data
						.low_stock_threshold}{/if}
			</span>. Produk di angka itu ke bawah ditandai.
		</p>
	</div>
</section>
