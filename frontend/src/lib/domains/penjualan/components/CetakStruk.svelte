<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { createCetakStrukMutation } from '../queries/penjualan.queries';
	import type { HasilCetak, NomorStruk } from '../schemas/penjualan.schema';

	/**
	 * Printing the Struk of one Penjualan, and the outcome of it.
	 *
	 * It is the same control in both places a Struk is printed from — the panel
	 * after a checkout, where the automatic print's result is passed in, and the
	 * `/penjualan` lookup — because the endpoint behind it is the same one
	 * (ADR-0017, keputusan 5).
	 *
	 * A failed print is shown on the screen with a button that prints again, never
	 * as a toast that passes: a Kasir whose printer jammed has to be able to fix it
	 * from here (ADR-0017, keputusan 1).
	 */
	let { nomorStruk, hasilAwal = null }: { nomorStruk: NomorStruk; hasilAwal?: HasilCetak | null } =
		$props();

	const cetak = createCetakStrukMutation();

	/**
	 * What the last print pressed here answered, or null before one has run. The
	 * result shown is this if it exists, and otherwise the automatic print's outcome
	 * the parent passed in — so the Kasir sees that outcome before pressing anything.
	 */
	let dicetak = $state<HasilCetak | null>(null);

	const hasil = $derived(dicetak ?? hasilAwal);

	/** A Struk that did not come out is retried; one that did is printed again. */
	const teksTombol = $derived(
		cetak.isPending ? 'Mencetak…' : hasil?.printed === false ? 'Cetak ulang Struk' : 'Cetak Struk'
	);

	async function cetakUlang() {
		try {
			dicetak = await cetak.mutateAsync(nomorStruk);
		} catch {
			// `cetak.error` carries the normalized message, rendered below.
		}
	}
</script>

<div class="space-y-2">
	<Button variant="outline" onclick={cetakUlang} disabled={cetak.isPending}>
		{teksTombol}
	</Button>

	{#if hasil}
		{#if hasil.printed}
			<p class="text-sm text-muted-foreground" role="status">Struk tercetak.</p>
		{:else}
			<p class="text-sm text-destructive" role="alert">
				{hasil.message ?? 'Struk gagal dicetak.'}
			</p>
		{/if}
	{/if}

	{#if cetak.error}
		<p class="text-sm text-destructive" role="alert">{cetak.error.message}</p>
	{/if}
</div>
