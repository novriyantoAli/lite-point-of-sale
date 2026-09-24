<script lang="ts">
	import Keranjang from './Keranjang.svelte';
	import Pembayaran from './Pembayaran.svelte';
	import PencarianProduk from './PencarianProduk.svelte';
	import StrukPenjualan from './StrukPenjualan.svelte';
	import type { HasilCheckout } from '../schemas/penjualan.schema';
	import { keranjangState } from '../state/keranjang.state.svelte';

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
	}

	function startNew() {
		hasil = null;
	}
</script>

<div class="space-y-8">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Kasir</h1>
		<p class="text-sm text-muted-foreground">
			Temukan Produk, susun keranjang, lalu bayar. Stok berkurang sendiri saat Penjualan tersimpan.
		</p>
	</div>

	{#if hasil}
		<StrukPenjualan sale={hasil.sale} cetak={hasil.print} onBaru={startNew} />
	{:else}
		<PencarianProduk onAdd={(produk) => keranjangState.add(produk)} />
		<Keranjang />
		<Pembayaran onCheckedOut={checkedOut} />
	{/if}
</div>
