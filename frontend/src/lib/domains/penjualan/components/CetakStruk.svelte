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
	 * from here (ADR-0017, keputusan 1). Pesannya ditulis dengan tinta, bukan abu
	 * abu dan bukan merah: yang menyatakan keadaannya adalah katanya, dan yang
	 * membawa tanda merah adalah field yang tidak valid (DESIGN.md, Zero-Grey).
	 */
	let { nomorStruk, hasilAwal = null }: { nomorStruk: NomorStruk; hasilAwal?: HasilCetak | null } =
		$props();

	const cetak = createCetakStrukMutation();

	/**
	 * What the last print answered: the result the retry got, or the automatic
	 * print's outcome the parent passed in. The mutation already holds its own
	 * answer, so nothing is copied into runes (Golden Rule 6, skill §3) — and only
	 * one of the two can be showing at a time.
	 */
	const hasil = $derived(cetak.data ?? hasilAwal);

	/** A Struk that did not come out is retried; one that did is printed again. */
	const teksTombol = $derived(
		cetak.isPending ? 'Mencetak…' : hasil?.printed === false ? 'Cetak ulang Struk' : 'Cetak Struk'
	);

	async function cetakUlang() {
		try {
			await cetak.mutateAsync(nomorStruk);
		} catch {
			// `cetak.error` carries the normalized message, rendered below.
		}
	}
</script>

<div class="border-b border-border px-2 py-1.5">
	<Button
		variant="ghost"
		size="sm"
		class="h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground"
		onclick={cetakUlang}
		disabled={cetak.isPending}
	>
		{teksTombol}
	</Button>

	<!--
		One outcome, in one place: the request that could not be made at all is the
		latest news, and it takes the screen over the print result it never replaced.
	-->
	{#if cetak.error}
		<p class="mt-1 text-xs font-semibold" role="alert">{cetak.error.message}</p>
	{:else if hasil?.printed === false}
		<p class="mt-1 text-xs font-semibold" role="alert">
			{hasil.message ?? 'Struk gagal dicetak.'}
		</p>
	{:else if hasil?.printed}
		<p class="mt-1 text-xs" role="status">Struk tercetak.</p>
	{/if}
</div>
