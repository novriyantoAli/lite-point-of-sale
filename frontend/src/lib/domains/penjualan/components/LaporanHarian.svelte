<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { cn, formatRupiah } from '$lib/utils';
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
	 * The day the report is over, and whether the field holds something the API
	 * could not read. An empty date field reads as today — the same answer the API
	 * gives for a missing `date` — so clearing the field is not a broken state,
	 * just the default day. A day that is neither empty nor a real calendar date is
	 * a broken field: it is refused with a message instead of a blank screen, and the
	 * report falls back to today while the message stands.
	 */
	const tanggalMentah = $derived(laporanState.tanggal.trim());
	const tanggalTerpilih = $derived(TanggalLaporanSchema.safeParse(tanggalMentah));
	const tanggalRusak = $derived(tanggalMentah !== '' && !tanggalTerpilih.success);
	const tanggal = $derived(tanggalTerpilih.data ?? tanggalHariIni());

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

	/**
	 * The chosen day as a person reads it ("25 September 2026"). Parsed as UTC and
	 * formatted in UTC, so a store east or west of Greenwich cannot see the day
	 * before or after the one the report is over.
	 */
	const DAY_FORMAT = new Intl.DateTimeFormat('id-ID', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		timeZone: 'UTC'
	});

	const tanggalPanjang = $derived(DAY_FORMAT.format(new Date(`${tanggal}T00:00:00Z`)));

	/**
	 * Papan mosaik (DESIGN.md, Layout): satu kolom selebar papan — layar Laporan
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
	 * Cacah dan tanggal di kepala modul. `tabular-nums` bukan hiasan di sini: angka
	 * yang berubah tanpa kelasnya menggeser tetangganya (DESIGN.md, Typography:
	 * "setiap angka uang, jumlah, dan Stok", tanpa kecuali).
	 */
	const CATATAN_ANGKA = `${CATATAN} tabular-nums`;
	/**
	 * Field Tanggal berdiri di ujung strip: `margin-left: auto` milik prototipe
	 * (`.strip__act`). Labelnya di ATAS field dengan jarak 3px, seperti setiap field
	 * lain di dunia ini (DESIGN.md, Fields & Inputs) — bukan di sebelahnya seperti
	 * prototipe menggambarnya.
	 */
	const FIELD_GRUP = 'ml-auto flex flex-col gap-[3px]';
	/** Fields & Inputs: field 26px, label 12px/600 di sebelahnya. */
	const FIELD_LABEL = 'text-xs leading-[1.2] font-semibold';
	const FIELD_INPUT =
		'h-[26px] w-[150px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]';
	/**
	 * Baris pesan — galat maupun keadaan — bertinta, bukan merah utilitas: merah
	 * hanya dua pekerjaan, tab dan harga (DESIGN.md, Secondary & Three-Percent
	 * Rule). Merah yang menandai field rusak milik garis dan outline field-nya.
	 */
	const PESAN = 'text-xs font-semibold';

	/**
	 * Modul: latar petak dengan garis rambut, dan hanya garis bawahnya yang
	 * digambar — dua modul bersebelahan berbagi satu garis, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const MODUL = 'border border-border border-b-0 bg-card';
	/**
	 * Jarak 8px sebelum modul Daftar bukan kelalaian: prototipe layar ini sendiri
	 * yang memisahkan dua modul dengan `margin-top: 8px`, jadi keduanya memang tidak
	 * bersebelahan.
	 */
	const MODUL_DAFTAR = `${MODUL} mt-2`;
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/**
	 * Sub-judul di dalam satu modul. Ia bergaris rambut, bukan Wash Grey ditutup
	 * garis tinta — supaya rincian yang berdiri di dalam modul tidak membaca
	 * sebagai modul ketiga (prototipe layar ini punya dua kepala).
	 */
	const SUB_JUDUL = 'border-b border-border px-2 py-1.5 text-[13px] font-bold tracking-[0.01em]';
	/** Satu baris catatan atau keadaan di dalam modul, selebar modulnya. */
	const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
	const MODUL_CATATAN = `${MODUL_BARIS} text-xs`;

	/**
	 * Figures (DESIGN.md): angka yang jadi jawaban layar, 20px/700 `tabular-nums`,
	 * selalu didahului label 12px/600 yang menyebut apa angka itu. Dua selnya
	 * dipisah satu garis rambut, bukan dua garis (Shared-Hairline).
	 */
	const STAT_GRID = 'grid grid-cols-2 border-b border-border';
	const STAT = 'flex flex-col gap-1 px-2 py-2.5';
	const STAT_KEDUA = `${STAT} border-l border-border`;
	const STAT_LABEL = 'text-xs leading-[1.2] font-semibold';
	const STAT_NILAI = 'text-[20px] leading-[1.1] font-bold tabular-nums';

	/** Tabel: satu garis rambut bersama antar baris, kepala Wash Grey + garis tinta. */
	const TABEL = 'w-full border-collapse text-[13px] leading-[1.25]';
	const TH =
		'border-b border-foreground bg-muted px-2 py-1.5 text-left text-xs font-semibold whitespace-nowrap';
	const TH_NUM = `${TH} text-right tabular-nums`;
	const BARIS = 'border-b border-border transition-colors last:border-b-0 hover:bg-muted';
	const TD = 'px-2 py-1.5 align-top';
	const TD_NUM = `${TD} text-right tabular-nums`;
	/** Kotak tabel: gulir mendatar di dalam modul kalau layarnya sempit. */
	const TABEL_BUNGKUS = 'overflow-x-auto border-b border-border';

	/** Daftar Penjualan: baris nama + keterangan + aksi, satu garis rambut bersama. */
	const DAFTAR = 'divide-y divide-border border-b border-border';
	const LIROW = 'flex flex-wrap items-center justify-between gap-x-3 gap-y-1.5 px-2 py-1.5';
	const LIROW_NAMA = 'text-[13px] font-medium';
	const LIROW_META = 'text-xs tabular-nums';
	const LIROW_AKSI = 'flex items-center gap-1.5';
	/**
	 * Tombol berbingkai memakai `variant="ghost"` lebih dulu: varian itu sudah
	 * membawa `hover:bg-muted`, persis latar hover dunia ini, dan tidak membawa apa
	 * pun yang harus dilawan (solution doc §3). Ukurannya `.lirow__acts .btn`
	 * prototipe: 26px, 13px/600.
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';

	/**
	 * A method nobody used never had a number to show, so the Figures rule writes
	 * `—` instead of `0` (DESIGN.md, Figures; issue #37). The row still stands — a
	 * missing method would read as a method that does not exist.
	 */
	function tanpaTransaksi(transactions: number): boolean {
		return transactions === 0;
	}
</script>

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya, dan field
		Tanggal di ujungnya (DESIGN.md, Typography; prototipe, `.strip__act`).
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Laporan</h1>
		<p class={CATATAN}>Omzet harian dan daftar Penjualan, dibaca dalam waktu lokal toko.</p>
		<div class={FIELD_GRUP}>
			<Label for="laporan-tanggal" class={FIELD_LABEL}>Tanggal</Label>
			<Input
				id="laporan-tanggal"
				type="date"
				class={FIELD_INPUT}
				bind:value={laporanState.tanggal}
				aria-invalid={tanggalRusak ? 'true' : undefined}
				aria-describedby={tanggalRusak ? 'laporan-tanggal-error' : undefined}
			/>
			<!--
				Tanggal yang rusak ditolak dengan pesan, bukan layar kosong: pesannya
				12px/600 di bawah fieldnya, bertinta (DESIGN.md, Fields & Inputs:
				warna baris pesan galat selalu tinta).
			-->
			{#if tanggalRusak}
				<p id="laporan-tanggal-error" class={PESAN} role="alert">
					Tanggal laporan tidak valid. Menampilkan hari ini.
				</p>
			{/if}
		</div>
	</div>

	<!--
		Modul Omzet harian: angka jawaban layar lebih dulu — total dan cacah
		transaksi sebagai Figures — lalu rinciannya sebagai tabel.
	-->
	<section class={MODUL} aria-labelledby="omzet-judul" aria-busy={omzet.isPending}>
		<div class={MODUL_KEPALA}>
			<h2 id="omzet-judul" class={MODUL_JUDUL}>Omzet harian</h2>
			<!-- The day spelled out, the way the prototype's head note reads it. -->
			<span class={CATATAN_ANGKA}>{tanggalPanjang}</span>
		</div>

		{#if omzet.isPending}
			<p class={MODUL_CATATAN}>Memuat omzet…</p>
		{:else if omzet.error}
			<div class={MODUL_BARIS}>
				<p class={PESAN} role="alert">{omzet.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void omzet.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if omzet.data}
			<div class={STAT_GRID}>
				<div class={STAT}>
					<span class={STAT_LABEL}>Total omzet</span>
					<span class={STAT_NILAI}>{formatRupiah(omzet.data.total)}</span>
				</div>
				<div class={STAT_KEDUA}>
					<span class={STAT_LABEL}>Jumlah transaksi</span>
					<span class={STAT_NILAI}>{omzet.data.transactions}</span>
				</div>
			</div>

			<!--
				Hari kosong bukan hari bernilai nol: keadaannya menjelaskan dirinya
				dengan kalimat, dan tabel metode di bawahnya menulis `—`, bukan `0`
				(DESIGN.md, Figures).
			-->
			{#if tanpaTransaksi(omzet.data.transactions)}
				<p class={MODUL_CATATAN}>Belum ada Penjualan pada tanggal ini.</p>
			{/if}

			<div class={TABEL_BUNGKUS}>
				<table class={TABEL}>
					<caption class="sr-only">Omzet per metode Pembayaran</caption>
					<thead>
						<tr>
							<th scope="col" class={TH}>Metode</th>
							<th scope="col" class={TH_NUM}>Transaksi</th>
							<th scope="col" class={TH_NUM}>Omzet</th>
						</tr>
					</thead>
					<tbody>
						{#each metodeUrut as metode (metode.method)}
							<tr class={BARIS}>
								<td class={TD}>{METODE_LABEL[metode.method]}</td>
								<!-- A method nobody used shows `—`, not `0`: there was no
								     number to answer with (DESIGN.md, Figures). -->
								<td class={TD_NUM}>
									{tanpaTransaksi(metode.transactions) ? '—' : metode.transactions}
								</td>
								<td class={TD_NUM}>
									{tanpaTransaksi(metode.transactions) ? '—' : formatRupiah(metode.total)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!--
				Omzet per Kasir adalah rincian di dalam modul Omzet harian yang sama,
				jadi judulnya sub-judul bergaris rambut — bukan modul kedua yang menjauh
				dari e2e `region "Omzet harian"` yang membacanya, dan bukan kepala modul
				ketiga yang prototipe tidak punya.
			-->
			<h3 class={SUB_JUDUL}>Omzet per Kasir</h3>
			{#if omzet.data.by_cashier.length === 0}
				<p class={MODUL_CATATAN}>Belum ada Penjualan, jadi belum ada yang bisa diatribusikan.</p>
			{:else}
				<div class={TABEL_BUNGKUS}>
					<table class={TABEL}>
						<caption class="sr-only">Omzet per Kasir</caption>
						<thead>
							<tr>
								<th scope="col" class={TH}>Kasir</th>
								<th scope="col" class={TH_NUM}>Transaksi</th>
								<th scope="col" class={TH_NUM}>Omzet</th>
							</tr>
						</thead>
						<tbody>
							{#each omzet.data.by_cashier as kasir (kasir.cashier_id)}
								<tr class={BARIS}>
									<td class={TD}>{kasir.cashier_name}</td>
									<td class={TD_NUM}>{kasir.transactions}</td>
									<td class={TD_NUM}>{formatRupiah(kasir.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		{/if}
	</section>

	<!--
		Modul Daftar Penjualan: setiap Penjualan hari itu sebagai satu baris, dengan
		aksinya sebagai kata (DESIGN.md, Named-Not-Hidden).
	-->
	<section class={MODUL_DAFTAR} aria-labelledby="daftar-judul" aria-busy={daftar.isPending}>
		<div class={MODUL_KEPALA}>
			<h2 id="daftar-judul" class={MODUL_JUDUL}>Daftar Penjualan</h2>
			<span class={CATATAN_ANGKA}>
				{daftar.isPending
					? 'memuat…'
					: daftar.error
						? '—'
						: `${daftar.data?.length ?? 0} transaksi`}
			</span>
		</div>

		{#if daftar.isPending}
			<p class={MODUL_CATATAN}>Memuat Penjualan…</p>
		{:else if daftar.error}
			<div class={MODUL_BARIS}>
				<p class={PESAN} role="alert">{daftar.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void daftar.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if daftar.data?.length === 0}
			<p class={MODUL_CATATAN}>Belum ada Penjualan pada tanggal ini.</p>
		{:else}
			<ul class={DAFTAR}>
				{#each daftar.data ?? [] as penjualan (penjualan.receipt_number)}
					<li>
						<div class={LIROW}>
							<div class="min-w-0">
								<p class={LIROW_NAMA}>
									Nomor Struk <span class="tabular-nums">{penjualan.receipt_number}</span>
								</p>
								<p class={LIROW_META}>
									{penjualan.created_at} · Kasir {penjualan.cashier_name} ·
									{METODE_LABEL[penjualan.method]}
								</p>
							</div>
							<div class={LIROW_AKSI}>
								<span class="text-[13px] font-semibold tabular-nums">
									{formatRupiah(penjualan.total)}
								</span>
								<Button variant="ghost" class={AKSI_MODUL} onclick={() => buka(penjualan)}>
									{terbuka === penjualan.receipt_number ? 'Tutup' : 'Buka'}
								</Button>
							</div>
						</div>

						{#if terbuka === penjualan.receipt_number}
							<!-- The same record and reprint the lookup shows: opening a sale from
							     the list is the same read, and the reprint is the same endpoint. -->
							<PenjualanTersimpan nomorStruk={penjualan.receipt_number} />
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</div>
