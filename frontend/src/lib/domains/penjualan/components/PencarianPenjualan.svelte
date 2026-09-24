<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { NomorStrukSchema, type NomorStruk } from '../schemas/penjualan.schema';
	import PenjualanTersimpan from './PenjualanTersimpan.svelte';

	/**
	 * The `/penjualan` lookup: type a Nomor Struk, see the Penjualan that was
	 * stored under it. It is the way back to a sale that already happened — what
	 * the reprint (#8) prints and what the sales list (#9) links to. Both Peran
	 * open it: CONTEXT.md gives the Kasir "cetak Struk", so a lookup cannot be
	 * Admin-only.
	 */
	let nomor = $state('');
	let nomorStruk = $state<NomorStruk | null>(null);
	let fieldError = $state('');

	/**
	 * Only a submitted search asks the API. The field is parsed with the same rule
	 * Go applies to the path, so "abc" is refused here instead of coming back as a
	 * 400 — while a number that names no Penjualan is left for Go's readable 404
	 * to answer.
	 */
	function cari(event: SubmitEvent) {
		event.preventDefault();

		const parsed = NomorStrukSchema.safeParse(nomor);
		if (!parsed.success) {
			fieldError = parsed.error.issues[0]?.message ?? 'Nomor Struk tidak valid.';
			nomorStruk = null;
			return;
		}

		fieldError = '';
		nomorStruk = parsed.data;
	}
</script>

<div class="space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Penjualan</h1>
		<p class="text-sm text-muted-foreground">
			Buka Penjualan yang tersimpan lewat Nomor Struk-nya.
		</p>
	</div>

	<form
		class="space-y-4 rounded-lg border p-4"
		role="search"
		aria-label="Cari Penjualan"
		onsubmit={cari}
		novalidate
	>
		<div class="space-y-2">
			<Label for="penjualan-nomor">Nomor Struk</Label>
			<Input
				id="penjualan-nomor"
				name="receipt_number"
				inputmode="numeric"
				autocomplete="off"
				autofocus
				bind:value={nomor}
				aria-invalid={fieldError ? 'true' : undefined}
				aria-describedby={fieldError ? 'penjualan-nomor-error' : undefined}
			/>
			{#if fieldError}
				<p id="penjualan-nomor-error" class="text-sm text-destructive">{fieldError}</p>
			{:else}
				<p class="text-sm text-muted-foreground">
					Nomor Struk tercetak di Struk — angka yang sama dipakai untuk cetak ulang.
				</p>
			{/if}
		</div>

		<Button type="submit">Cari</Button>
	</form>

	{#if nomorStruk !== null}
		<PenjualanTersimpan {nomorStruk} />
	{:else if !fieldError}
		<p class="text-sm text-muted-foreground">
			Ketik Nomor Struk lalu tekan Cari untuk melihat Penjualannya.
		</p>
	{/if}
</div>
