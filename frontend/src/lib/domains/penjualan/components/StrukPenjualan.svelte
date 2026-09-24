<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import type { Penjualan } from '../schemas/penjualan.schema';
	import RincianPenjualan from './RincianPenjualan.svelte';

	/**
	 * What the Kasir sees once a keranjang has become a Penjualan: the record of
	 * the sale the till just wrote, and the way back to a new one.
	 *
	 * Printing the Struk is #8; this is the record of the sale the API stored.
	 */
	let { sale, onBaru }: { sale: Penjualan; onBaru: () => void } = $props();
</script>

<section class="space-y-4 rounded-lg border p-4" aria-labelledby="kasir-struk-judul">
	<h2 id="kasir-struk-judul" class="font-medium">Penjualan tercatat</h2>

	<RincianPenjualan {sale} />

	<Button onclick={onBaru}>Penjualan Baru</Button>
</section>
