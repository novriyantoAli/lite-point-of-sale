<script lang="ts">
	import {
		AKSI_MODUL,
		AKSI_STRIP,
		CATATAN,
		CATATAN_ANGKA,
		COMMIT_BARIS,
		DAFTAR,
		JUDUL,
		MODUL,
		MODUL_BARIS,
		MODUL_JUDUL,
		MODUL_KEPALA,
		MODUL_TEKS,
		PAPAN,
		STRIP
	} from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';
	import { createBackupListQuery, createBackupMutation } from '../queries/backup.queries';

	/**
	 * The Backup screen: the Admin's safety net (issue #10). It lists the
	 * snapshots the store has kept and offers a manual export now. The daily
	 * automatic backup runs on the server — this screen does not turn it on or
	 * off, it only takes the same snapshot the scheduler would.
	 *
	 * The database path never crosses this boundary: the browser learns a snapshot
	 * exists, how big it is, and when it was taken, and nothing about where on the
	 * server it lives (backup.schema.ts).
	 */
	const backups = createBackupListQuery();
	const exportBackup = createBackupMutation();

	let notice = $state('');

	async function backupNow() {
		notice = '';
		try {
			const created = await exportBackup.mutateAsync();
			notice = `Backup dibuat: ${created.name}`;
		} catch {
			// exportBackup.error carries the normalized message, rendered below.
		}
	}

	/**
	 * The API answers oldest first; the screen answers the Admin's question —
	 * "what did I take last?" — so the newest snapshot leads (prototipe, catatan
	 * kepala modul: "terbaru di atas"). The list is copied before sorting because
	 * the query cache's array is not ours to reorder.
	 */
	const tersusun = $derived(
		[...(backups.data ?? [])].sort(
			(a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
		)
	);

	/**
	 * Cacah di kepala modul: satu `—` saat daftar gagal dibaca, bukan `0` —
	 * angka yang belum bisa dihitung tidak ditulis nol (DESIGN.md, Figures), dan
	 * LaporanHarian menulis tanda yang sama untuk alasan yang sama.
	 */
	const catatanKepala = $derived(
		backups.isPending
			? 'Memuat…'
			: backups.error
				? '—'
				: `${tersusun.length} file · terbaru di atas`
	);

	/**
	 * The store's local clock, in the one date convention this app already uses:
	 * long month, numeric day and year (LaporanHarian's `DAY_FORMAT`), plus the time
	 * of day. No second date format is introduced (issue #39).
	 */
	const TIME_FORMAT = new Intl.DateTimeFormat('id-ID', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		hour12: false
	});

	/** A snapshot length as a person reads it, not a raw byte count. */
	function formatBytes(bytes: number): string {
		if (bytes < 1024) {
			return `${bytes} B`;
		}
		if (bytes < 1024 * 1024) {
			return `${(bytes / 1024).toFixed(1)} KB`;
		}
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	/** The API's RFC 3339 timestamp as the store's local clock writes it. */
	function formatTime(createdAt: string): string {
		const date = new Date(createdAt);
		if (Number.isNaN(date.getTime())) {
			return createdAt;
		}

		return TIME_FORMAT.format(date);
	}

	/**
	 * Baris pemberitahuan hasil ekspor dan baris galatnya: satu pita selebar papan
	 * di bawah strip, di luar modul daftar. Satu kosakata untuk keduanya supaya
	 * dua baris yang berdampingan tidak berbeda baju; yang menambahkan
	 * `tabular-nums` adalah pemberitahuan yang memuat nama berkas berdigit.
	 */
	const BARIS_PESAN = 'border border-b-0 border-border bg-card px-2 py-1.5 text-xs font-semibold';
	const LIROW = 'flex flex-wrap items-center justify-between gap-x-3 gap-y-1 px-2 py-1.5';
	const LIROW_UTAMA = 'min-w-0';
	/**
	 * Nama berkas memuat digit dan bisa dibandingkan antar-baris, jadi ia
	 * `tabular-nums` seperti seluruh angka lain (prototipe: `lirow__name tnum`).
	 */
	const LIROW_NAMA = 'text-[13px] font-medium tabular-nums';
	const LIROW_META = 'text-xs tabular-nums';
	/** Baris kerangka selama memuat: bentuk baris yang sama, tanpa datanya. */
	const BARIS_HANTU = [0, 1, 2];
</script>

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya, dan
		Commit Button layar ini di baris yang sama (DESIGN.md, Typography).
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Backup</h1>
		<p class={CATATAN}>
			Salinan database toko. Backup harian berjalan otomatis; tombol di samping mengambil salinan
			manual saat ini juga.
		</p>
		<span class={AKSI_STRIP}>
			<!--
				Saat sedang berjalan tombolnya menyebut keadaannya dengan kata, bukan
				berputar tanpa kalimat — dan keadaannya disampaikan dengan kehilangan
				tinta, bukan opasitas (DESIGN.md, State-Is-Not-Faded).
			-->
			<Button class={COMMIT_BARIS} onclick={backupNow} disabled={exportBackup.isPending}>
				{exportBackup.isPending ? 'Membuat backup…' : 'Backup sekarang'}
			</Button>
		</span>
	</div>

	<!--
		Pemberitahuan hasil ekspor berdiri sebagai barisnya sendiri di bawah strip,
		di luar modul daftar: ia menjawab tindakan manual, bukan keadaan daftar.
		Baris ini berbagi garis rambut dengan strip lewat `-mb-px` milik strip
		(StokList, pola yang sama).
	-->
	{#if notice}
		<p class={cn(BARIS_PESAN, 'tabular-nums')} role="status">
			{notice}
		</p>
	{/if}
	{#if exportBackup.error}
		<p class={BARIS_PESAN} role="alert">
			{exportBackup.error.message}
		</p>
	{/if}

	<section class={MODUL} aria-labelledby="salinan-judul" aria-busy={backups.isPending}>
		<div class={MODUL_KEPALA}>
			<h2 id="salinan-judul" class={MODUL_JUDUL}>Salinan tersimpan</h2>
			<!--
				Cacah salinan datang dari daftar yang sudah dimuat, bukan permintaan
				kedua (DESIGN.md, Tabs: cacah tab yang belum diputuskan adalah cacah
				tab di rel, bukan yang ini).
			-->
			<span class={CATATAN_ANGKA}>
				{catatanKepala}
			</span>
		</div>

		{#if backups.isPending}
			<!--
				Kerangkanya bukan lingkaran berputar: baris-baris pada bentuk yang sama,
				jadi kepala modul di atasnya tidak bergerak saat datanya tiba.
			-->
			<div aria-hidden="true">
				{#each BARIS_HANTU as baris (baris)}
					<div class={MODUL_BARIS}>
						<span class="block h-[13px] w-1/3 bg-muted"></span>
						<span class="mt-1 block h-[12px] w-1/4 bg-muted"></span>
					</div>
				{/each}
			</div>
		{:else if backups.error}
			<div class={MODUL_BARIS}>
				<p class="text-xs font-semibold" role="alert">{backups.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void backups.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if tersusun.length === 0}
			<!--
				Modul kosong menjelaskan apa yang akan muncul dan apa yang harus
				dilakukan — bukan daftar kosong yang tak bisa dibedakan dari kegagalan
				(DESIGN.md, Do's).
			-->
			<p class={MODUL_TEKS}>
				Belum ada backup. Tekan “Backup sekarang” untuk membuat yang pertama — salinan akan muncul
				di sini beserta nama, ukuran, dan waktunya.
			</p>
		{:else}
			<ul class={DAFTAR}>
				{#each tersusun as backup (backup.name)}
					<li class={LIROW}>
						<div class={LIROW_UTAMA}>
							<p class={LIROW_NAMA}>{backup.name}</p>
							<p class={LIROW_META}>
								{formatTime(backup.created_at)} · {formatBytes(backup.size)}
							</p>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</div>
