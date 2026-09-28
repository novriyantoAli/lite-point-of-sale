<script lang="ts">
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

	/**
	 * Papan mosaik (DESIGN.md, Layout): satu kolom selebar papan — layar Penjualan
	 * bukan papan tiga kolom milik Kasir. Padding luarnya milik rel
	 * (`(app)/+layout`), jadi di sini tidak ada padding lagi.
	 */
	const PAPAN = 'grid grid-cols-1 gap-0';
	/**
	 * Strip judul: satu-satunya tempat ukuran 24px muncul di layar ini. `-mb-px`
	 * plus `z-[2]` menariknya turun satu piksel, jadi garis tintanya yang menutup
	 * modul di bawahnya alih-alih bertumpuk dengan garis rambut (DESIGN.md,
	 * Shared-Hairline).
	 */
	const STRIP =
		'z-[2] col-span-full -mb-px flex flex-wrap items-baseline gap-x-3 gap-y-1 border border-border border-b-foreground bg-card px-2 py-2';
	const JUDUL = 'text-2xl leading-none font-bold tracking-[-0.015em]';
	/** Seluruh teks 12px; lantai huruf dunia ini (DESIGN.md, Legibility Floor). */
	const CATATAN = 'text-xs';
	/**
	 * Modul: latar petak dengan garis rambut, dan hanya garis bawahnya yang
	 * digambar — dua modul bersebelahan berbagi satu garis, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const MODUL = 'border border-border border-b-0 bg-card';
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/** Satu blok isi di dalam modul, dan garis rambut bawahnya milik modul itu. */
	const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
	/** Fields & Inputs: field 26px, label 12px/600 di atasnya dengan jarak 3px. */
	const FIELD = 'flex min-w-0 flex-1 flex-col gap-[3px]';
	const FIELD_LABEL = 'text-xs leading-[1.2] font-semibold';
	/**
	 * Field yang isinya angka: Nomor Struk. `tabular-nums` menahan lebar digit yang
	 * sedang diketik, seperti setiap angka lain di dunia ini (DESIGN.md, Typography:
	 * "tanpa kecuali").
	 */
	const FIELD_INPUT =
		'h-[26px] border-border bg-card px-1.5 py-0 text-[13px] tabular-nums shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]';
	/**
	 * Field yang tidak valid: pesannya 12px/600 di bawah field, bertinta — bukan
	 * merah utilitas. Merah hanya milik garis dan outline field-nya (DESIGN.md,
	 * Fields & Inputs).
	 */
	const PESAN = 'text-xs font-semibold';
	/**
	 * Commit Button: satu-satunya bidang bertinta penuh di layar ini. DESIGN.md
	 * menggambarnya 40px selebar kolom keranjang; di dalam baris aksi ia memakai
	 * tinggi `.btn` dunia ini, 26px, karena yang ia warisi adalah tintanya, bukan
	 * lebarnya — kosakata yang sama dengan tombol Tambah Stok (#34) dan Tambah
	 * Pengguna (#35).
	 */
	const COMMIT =
		'h-[26px] border-foreground bg-primary px-2 text-[13px] font-semibold text-primary-foreground hover:border-foreground hover:bg-primary hover:text-primary-foreground focus-visible:border-foreground';
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
				<div class={FIELD}>
					<Label for="penjualan-nomor" class={FIELD_LABEL}>Nomor Struk</Label>
					<Input
						id="penjualan-nomor"
						name="receipt_number"
						inputmode="numeric"
						autocomplete="off"
						autofocus
						class={FIELD_INPUT}
						bind:value={nomor}
						aria-invalid={fieldError ? 'true' : undefined}
						aria-describedby={fieldError ? 'penjualan-nomor-error' : undefined}
					/>
					{#if fieldError}
						<p id="penjualan-nomor-error" class={PESAN}>{fieldError}</p>
					{/if}
				</div>

				<Button type="submit" class={COMMIT}>Cari</Button>
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
		<p class={cn(CATATAN, 'mt-2 border border-t-0 border-border bg-card px-2 py-1.5')}>
			Ketik Nomor Struk lalu tekan Cari untuk melihat Penjualannya.
		</p>
	{/if}
</div>
