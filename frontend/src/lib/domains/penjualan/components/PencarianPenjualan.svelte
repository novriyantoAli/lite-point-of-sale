<script lang="ts">
	import {
		CATATAN,
		COMMIT_BARIS,
		ERROR,
		FIELD,
		INPUT_ANGKA,
		JUDUL,
		LABEL,
		MODUL,
		MODUL_BARIS,
		MODUL_JUDUL,
		MODUL_KEPALA,
		PAPAN,
		STRIP
	} from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { cn } from '$lib/utils';
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

	/** Fields & Inputs: field 26px, label 12px/600 di atasnya dengan jarak 3px. */
	const FIELD_CARI = cn(FIELD, 'min-w-0 flex-1');
	/**
	 * Field yang tidak valid: pesannya 12px/600 di bawah field, bertinta — bukan
	 * merah utilitas. Merah hanya milik garis dan outline field-nya (DESIGN.md,
	 * Fields & Inputs).
	 */
	const PESAN = ERROR;
</script>

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya
		(DESIGN.md, Typography). Layar ini tidak punya aksi utama di strip — aksinya
		adalah formulir pencarian di bawahnya.
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Penjualan</h1>
		<p class={CATATAN}>Buka Penjualan yang tersimpan lewat Nomor Struk-nya.</p>
	</div>

	<!--
		Modul Cari Penjualan: field Nomor Struk dan tombol Cari, persis prototipe.
		Kepalanya menyebut di mana Nomor Struk tercetak, badannya menyebut angka yang
		sama dipakai untuk cetak ulang.
	-->
	<section class={MODUL} aria-labelledby="cari-penjualan-judul">
		<div class={MODUL_KEPALA}>
			<h2 id="cari-penjualan-judul" class={MODUL_JUDUL}>Cari Penjualan</h2>
			<span class={CATATAN}>Nomor Struk tercetak di Struk</span>
		</div>

		<form class={MODUL_BARIS} role="search" aria-label="Cari Penjualan" onsubmit={cari} novalidate>
			<div class="flex flex-wrap items-end gap-1.5">
				<div class={FIELD_CARI}>
					<Label for="penjualan-nomor" class={LABEL}>Nomor Struk</Label>
					<Input
						id="penjualan-nomor"
						name="receipt_number"
						inputmode="numeric"
						autocomplete="off"
						autofocus
						class={INPUT_ANGKA}
						bind:value={nomor}
						aria-invalid={fieldError ? 'true' : undefined}
						aria-describedby={fieldError ? 'penjualan-nomor-error' : undefined}
					/>
					{#if fieldError}
						<p id="penjualan-nomor-error" class={PESAN}>{fieldError}</p>
					{/if}
				</div>

				<Button type="submit" class={COMMIT_BARIS}>Cari</Button>
			</div>

			<p class={cn(CATATAN, 'mt-1.5')}>Angka yang sama dipakai untuk cetak ulang Struk.</p>
		</form>
	</section>

	{#if nomorStruk !== null}
		<PenjualanTersimpan {nomorStruk} />
	{:else if !fieldError}
		<!--
			Keadaan kosong menyebut dirinya dengan kata, bukan layar kosong: sebelum
			Nomor Struk apa pun dicari, orang yang membukanya harus tahu apa yang
			dikerjakan layar ini.
		-->
		<p class={cn(CATATAN, 'mt-2 border border-border bg-card px-2 py-1.5')}>
			Ketik Nomor Struk lalu tekan Cari untuk melihat Penjualannya.
		</p>
	{/if}
</div>
