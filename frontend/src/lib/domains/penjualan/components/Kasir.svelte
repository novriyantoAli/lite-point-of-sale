<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HasilCheckout } from '../schemas/penjualan.schema';
	import { keranjangState } from '../state/keranjang.state.svelte';
	import { pencarianState } from '../state/pencarian.state.svelte';
	import KatalogProduk from './KatalogProduk.svelte';
	import Keranjang from './Keranjang.svelte';
	import Pembayaran from './Pembayaran.svelte';
	import PencarianProduk from './PencarianProduk.svelte';
	import StrukPenjualan from './StrukPenjualan.svelte';

	/**
	 * The till: find a Produk, build a keranjang, take a Pembayaran — Tunai, or QRIS,
	 * Debit or Transfer recorded — and record the Penjualan (CONTEXT.md). It is the
	 * one screen a Kasir needs; an Admin sells from the same screen.
	 *
	 * It owns nothing but which of its two faces is showing — the running sale or the
	 * one just recorded. The keranjang belongs to the domain's own state, so the
	 * lookup and the payment form both work on the same draft without props threaded
	 * between them.
	 */
	let hasil = $state<HasilCheckout | null>(null);

	function checkedOut(checkedOut: HasilCheckout) {
		hasil = checkedOut;
		// The keranjang is emptied only once the Penjualan is stored: a refused
		// checkout leaves the Kasir's work exactly where it was.
		keranjangState.clear();
		// Saringan pembeli sebelumnya tidak boleh menjadi saringan pembeli
		// berikutnya: keranjang berikutnya mulai dari katalog yang utuh.
		pencarianState.reset();
	}

	function startNew() {
		hasil = null;
		pencarianState.reset();
	}

	/**
	 * Papan mosaik: tiga kolom dengan lebar tetap di tepinya, dan strip judul
	 * selebar papan di atasnya. Di bawah 1081px papan menumpuk jadi satu kolom —
	 * laptop di meja kasir adalah perangkatnya, jadi ini hanya supaya tidak rusak
	 * (DESIGN.md, Layout).
	 */
	const PAPAN = 'grid grid-cols-1 gap-0';
	/**
	 * Kolom: tumpukan modul, dan tiap modul membawa garis bawahnya sendiri — dua
	 * modul bersebelahan berbagi satu garis rambut (DESIGN.md, Shared-Hairline).
	 * Garis tegaknya dipasang per kolom supaya garis antara dua kolom digambar satu
	 * kali, bukan dua.
	 */
	const KOLOM = 'flex flex-col border-border bg-card border-r';
	const KOLOM_SISI = `${KOLOM} border-l`;
	// Kolom tengah dan kanan hanya butuh garis kirinya saat papannya menumpuk.
	const KOLOM_TENGAH = `${KOLOM} border-l min-[1081px]:border-l-0`;
</script>

<!--
	Strip judul: satu-satunya tempat ukuran 24px muncul di layar ini, dan garis
	tintanya yang menutup kepala papan (DESIGN.md, Typography).
-->
<div
	class={cn(
		PAPAN,
		hasil ? 'min-[1081px]:grid-cols-1' : 'min-[1081px]:grid-cols-[232px_minmax(0,1fr)_300px]'
	)}
>
	<div
		class="col-span-full flex flex-wrap items-baseline gap-x-3 gap-y-1 border border-border border-b-foreground bg-card px-2 py-2"
	>
		<h1 class="text-2xl leading-none font-bold tracking-[-0.015em]">Kasir</h1>
		<p class="text-xs">
			Temukan Produk, susun keranjang, lalu bayar. Stok berkurang sendiri saat Penjualan tersimpan.
		</p>
	</div>

	{#if hasil}
		<StrukPenjualan sale={hasil.sale} cetak={hasil.print} onBaru={startNew} />
	{:else}
		<div class={KOLOM_SISI}>
			<PencarianProduk />
		</div>

		<div class={KOLOM_TENGAH}>
			<KatalogProduk onAdd={(produk) => keranjangState.add(produk)} />
		</div>

		<!--
			Keranjang dan Pembayaran adalah satu kolom, bukan dua: yang memisahkan
			mereka adalah garis tinta di bawah total berjalan, dan pita itu berada di
			kolom yang sama dengan total yang dimaksudnya.
		-->
		<div class={KOLOM_TENGAH}>
			<Keranjang />
			<Pembayaran onCheckedOut={checkedOut} />
		</div>
	{/if}
</div>
