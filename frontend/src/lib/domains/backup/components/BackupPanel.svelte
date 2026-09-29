<script lang="ts">
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
	 * Papan mosaik (DESIGN.md, Layout): satu kolom selebar papan — layar Backup
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
	 * Cacah salinan di kepala modul. `tabular-nums` bukan hiasan: DESIGN.md
	 * (Typography) menuntutnya untuk "setiap angka uang, jumlah, dan Stok", tanpa
	 * kecuali, dan prototipe menulis kelas `tnum` di catatan kepala.
	 */
	const CATATAN_ANGKA = `${CATATAN} tabular-nums`;
	/** Aksi strip mengambil sisa baris ke kanan, seperti `strip__act` prototipe. */
	const AKSI_STRIP = 'ml-auto flex items-center gap-1.5';
	/**
	 * Modul: latar petak dengan garis rambut, dan hanya garis bawahnya yang
	 * digambar — modul terakhir menutup papan, jadi garis bawahnya milik papan
	 * (DESIGN.md, Shared-Hairline).
	 */
	const MODUL = 'border border-border border-b-0 bg-card';
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/** Satu baris catatan atau keadaan di dalam modul, selebar modulnya. */
	const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
	const MODUL_CATATAN = `${MODUL_BARIS} text-xs`;
	/**
	 * Baris pemberitahuan hasil ekspor dan baris galatnya: satu pita selebar papan
	 * di bawah strip, di luar modul daftar. Satu kosakata untuk keduanya supaya
	 * dua baris yang berdampingan tidak berbeda baju; yang menambahkan
	 * `tabular-nums` adalah pemberitahuan yang memuat nama berkas berdigit.
	 */
	const BARIS_PESAN = 'border border-b-0 border-border bg-card px-2 py-1.5 text-xs font-semibold';
	/** Daftar salinan: baris nama + keterangan, satu garis rambut bersama. */
	const DAFTAR = 'divide-y divide-border border-b border-border';
	const LIROW = 'flex flex-wrap items-center justify-between gap-x-3 gap-y-1 px-2 py-1.5';
	const LIROW_UTAMA = 'min-w-0';
	/**
	 * Nama berkas memuat digit dan bisa dibandingkan antar-baris, jadi ia
	 * `tabular-nums` seperti seluruh angka lain (prototipe: `lirow__name tnum`).
	 */
	const LIROW_NAMA = 'text-[13px] font-medium tabular-nums';
	const LIROW_META = 'text-xs tabular-nums';
	/**
	 * Commit Button baris dunia ini: 26px, selebar katanya, huruf 13px/600, bidang
	 * tinta penuh saat hidup dan — saat mati — kehilangan tintanya jadi putih
	 * bergaris putus-putus, bukan dipudar (DESIGN.md, Commit Button;
	 * State-Is-Not-Faded). "Backup sekarang" adalah Commit Button layar ini karena
	 * ia satu-satunya bidang bertinta penuh di sini.
	 */
	const AKSI_UTAMA =
		'h-[26px] border-foreground bg-primary px-2 text-[13px] font-semibold text-primary-foreground hover:border-foreground hover:bg-primary hover:text-primary-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:border-dashed disabled:border-border disabled:bg-card disabled:text-foreground disabled:opacity-100';
	/**
	 * Tombol berbingkai memakai `variant="ghost"` lebih dulu: varian itu sudah
	 * membawa `hover:bg-muted` — persis latar hover dunia ini — dan tidak membawa
	 * apa pun yang harus dilawan (solution doc §3). Ukurannya `.btn` dunia ini:
	 * 26px, 13px/600, keadaan mati tidak dipudarkan.
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
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
			<Button class={AKSI_UTAMA} onclick={backupNow} disabled={exportBackup.isPending}>
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
			<p class={MODUL_CATATAN}>
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
