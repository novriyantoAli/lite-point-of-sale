<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { formatRupiah } from '$lib/utils';
	import { createOmzetHarianQuery, createPenjualanListQuery } from '../queries/penjualan.queries';
	import {
		METODE_LABEL,
		METODE_URUT,
		TanggalLaporanSchema,
		tanggalHariIni,
		type PenjualanRingkas
	} from '../schemas/penjualan.schema';
	import { laporanState } from '../state/laporan.state.svelte';
	import PenjualanTersimpan from './PenjualanTersimpan.svelte';

	/**
	 * The Laporan screen (#9): the omzet of one store-local day and the Penjualan
	 * that made it. It is Admin-only — revenue is not the till's screen — and it
	 * reuses the lookup's record and reprint for a row, because a Penjualan opened
	 * from the list is the same Penjualan the lookup shows.
	 *
	 * The day is the only input, and it is UI state (`laporanState`): both queries
	 * read it through a thunk, so changing the date re-derives them.
	 */
	/**
	 * The day the report is over. An empty date field reads as today — the same
	 * answer the API gives for a missing `date` — so clearing the field is not a
	 * broken state, just the default day.
	 */
	const tanggal = $derived(
		TanggalLaporanSchema.safeParse(laporanState.tanggal).data ?? tanggalHariIni()
	);

	const omzet = createOmzetHarianQuery(() => tanggal);
	const daftar = createPenjualanListQuery(() => tanggal);

	/**
	 * The method breakdown in the till's order. `METODE_URUT` is the one source for
	 * that order (ADR-0016): the API's answer carries the numbers, and this lookup
	 * keeps the screen from introducing a second order that could drift from it.
	 */
	const metodeUrut = $derived(
		METODE_URUT.map(
			(method) =>
				omzet.data?.by_method.find((row) => row.method === method) ?? {
					method,
					total: 0,
					transactions: 0
				}
		)
	);

	/**
	 * The Nomor Struk whose record is open under the list. One at a time: opening a
	 * second row closes the first, so the screen never grows two reprint buttons
	 * that could be pressed for the wrong sale.
	 */
	let terbuka = $state<number | null>(null);

	function buka(penjualan: PenjualanRingkas) {
		terbuka = terbuka === penjualan.receipt_number ? null : penjualan.receipt_number;
	}
</script>

<div class="space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Laporan</h1>
		<p class="text-sm text-muted-foreground">
			Omzet harian dan daftar Penjualan, dibaca dalam waktu lokal toko.
		</p>
	</div>

	<div class="max-w-xs space-y-2">
		<Label for="laporan-tanggal">Tanggal</Label>
		<Input id="laporan-tanggal" type="date" bind:value={laporanState.tanggal} />
		<p class="text-sm text-muted-foreground">Kosong berarti hari ini.</p>
	</div>

	<section class="space-y-3" aria-label="Omzet harian">
		<h2 class="text-lg font-semibold">Omzet harian</h2>

		{#if omzet.isPending}
			<p class="text-sm text-muted-foreground">Memuat omzet…</p>
		{:else if omzet.error}
			<div class="space-y-3">
				<p class="text-sm text-destructive" role="alert">{omzet.error.message}</p>
				<Button variant="outline" size="sm" onclick={() => void omzet.refetch()}>Coba lagi</Button>
			</div>
		{:else if omzet.data}
			<div class="grid gap-3 sm:grid-cols-2">
				<div class="rounded-lg border p-4">
					<p class="text-sm text-muted-foreground">Total omzet</p>
					<p class="text-2xl font-semibold tabular-nums">{formatRupiah(omzet.data.total)}</p>
				</div>
				<div class="rounded-lg border p-4">
					<p class="text-sm text-muted-foreground">Jumlah transaksi</p>
					<p class="text-2xl font-semibold tabular-nums">{omzet.data.transactions}</p>
				</div>
			</div>

			{#if omzet.data.transactions === 0}
				<p class="text-sm text-muted-foreground">Belum ada Penjualan pada tanggal ini.</p>
			{/if}

			<!-- Every method the API answers, in the till's order: a method nobody used
			     is a zero row rather than a missing one, so the breakdown is complete. -->
			<div class="overflow-x-auto rounded-lg border">
				<table class="w-full text-sm">
					<caption class="sr-only">Omzet per metode Pembayaran</caption>
					<thead class="border-b bg-muted/50 text-left">
						<tr>
							<th scope="col" class="p-3 font-medium">Metode</th>
							<th scope="col" class="p-3 text-right font-medium">Transaksi</th>
							<th scope="col" class="p-3 text-right font-medium">Omzet</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each metodeUrut as metode (metode.method)}
							<tr>
								<td class="p-3">{METODE_LABEL[metode.method]}</td>
								<td class="p-3 text-right tabular-nums">{metode.transactions}</td>
								<td class="p-3 text-right tabular-nums">{formatRupiah(metode.total)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<h3 class="text-base font-semibold">Omzet per Kasir</h3>
			{#if omzet.data.by_cashier.length === 0}
				<p class="text-sm text-muted-foreground">
					Belum ada Penjualan, jadi belum ada yang bisa diatribusikan.
				</p>
			{:else}
				<div class="overflow-x-auto rounded-lg border">
					<table class="w-full text-sm">
						<caption class="sr-only">Omzet per Kasir</caption>
						<thead class="border-b bg-muted/50 text-left">
							<tr>
								<th scope="col" class="p-3 font-medium">Kasir</th>
								<th scope="col" class="p-3 text-right font-medium">Transaksi</th>
								<th scope="col" class="p-3 text-right font-medium">Omzet</th>
							</tr>
						</thead>
						<tbody class="divide-y">
							{#each omzet.data.by_cashier as kasir (kasir.cashier_id)}
								<tr>
									<td class="p-3">{kasir.cashier_name}</td>
									<td class="p-3 text-right tabular-nums">{kasir.transactions}</td>
									<td class="p-3 text-right tabular-nums">{formatRupiah(kasir.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		{/if}
	</section>

	<section class="space-y-3" aria-label="Daftar Penjualan">
		<h2 class="text-lg font-semibold">Daftar Penjualan</h2>

		{#if daftar.isPending}
			<p class="text-sm text-muted-foreground">Memuat Penjualan…</p>
		{:else if daftar.error}
			<div class="space-y-3">
				<p class="text-sm text-destructive" role="alert">{daftar.error.message}</p>
				<Button variant="outline" size="sm" onclick={() => void daftar.refetch()}>Coba lagi</Button>
			</div>
		{:else if daftar.data?.length === 0}
			<p class="text-sm text-muted-foreground">Belum ada Penjualan pada tanggal ini.</p>
		{:else}
			<ul class="space-y-2">
				{#each daftar.data ?? [] as penjualan (penjualan.receipt_number)}
					<li class="rounded-lg border">
						<div class="flex flex-wrap items-center justify-between gap-3 p-3 text-sm">
							<div class="space-y-0.5">
								<p class="font-medium">
									Nomor Struk <span class="tabular-nums">{penjualan.receipt_number}</span>
								</p>
								<p class="text-muted-foreground">
									{penjualan.created_at} · Kasir {penjualan.cashier_name} ·
									{METODE_LABEL[penjualan.method]}
								</p>
							</div>
							<div class="flex items-center gap-3">
								<span class="font-medium tabular-nums">{formatRupiah(penjualan.total)}</span>
								<Button variant="outline" size="sm" onclick={() => buka(penjualan)}>
									{terbuka === penjualan.receipt_number ? 'Tutup' : 'Buka'}
								</Button>
							</div>
						</div>

						{#if terbuka === penjualan.receipt_number}
							<div class="border-t p-3">
								<!-- The same record and reprint the lookup shows: opening a sale from
								     the list is the same read, and the reprint is the same endpoint. -->
								<PenjualanTersimpan nomorStruk={penjualan.receipt_number} />
							</div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</div>
