<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import type { HasilCetak, Penjualan } from '../schemas/penjualan.schema';
	import CetakStruk from './CetakStruk.svelte';
	import RincianPenjualan from './RincianPenjualan.svelte';

	/**
	 * What the Kasir sees once a keranjang has become a Penjualan: the record of
	 * the sale the till just wrote, the outcome of the Struk the API printed for
	 * it, and the way back to a new one.
	 *
	 * A print that failed is shown here with a button that prints again — the
	 * Penjualan is stored either way, and the retry goes through the same endpoint
	 * as the automatic print (ADR-0017, keputusan 1 and 5).
	 */
	let { sale, cetak, onBaru }: { sale: Penjualan; cetak: HasilCetak; onBaru: () => void } =
		$props();
</script>

<section class="space-y-4 rounded-lg border p-4" aria-labelledby="kasir-struk-judul">
	<h2 id="kasir-struk-judul" class="font-medium">Penjualan tercatat</h2>

	<RincianPenjualan {sale} />

	<CetakStruk nomorStruk={sale.receipt_number} hasilAwal={cetak} />

	<Button onclick={onBaru}>Penjualan Baru</Button>
</section>
