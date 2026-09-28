<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';
	import { createHealthQuery } from '../queries/health.queries';

	/**
	 * Status layanan: modul Beranda yang menjawab "API Go-nya hidup?" lewat
	 * `/api/health` (BFF same-origin, ADR-0001). Ia menggambar satu modul mosaik,
	 * bukan satu kartu: kepala Wash Grey ditutup garis tinta, badan petak dengan
	 * garis rambut, radius nol, dan tanpa bayangan (DESIGN.md, Modules).
	 */
	const health = createHealthQuery();

	const isHealthy = $derived(health.data?.status === 'ok');

	/** Modul: petak dengan garis rambut, dan hanya garis atasnya yang digambar. */
	const MODUL = 'border border-border border-b-0 bg-card';
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/** Satu baris isi di dalam modul; garis bawahnya menutup modulnya. */
	const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
	const MODUL_TEKS = `${MODUL_BARIS} text-xs`;
	/** Kepala modul membawa catatannya sendiri: 12px, lantai huruf dunia ini. */
	const CATATAN = 'text-xs';
	/**
	 * Quiet tag untuk keadaan sehat: isian Wash Grey dengan garis rambut. Merah
	 * utilitas tidak dipakai di sini — ia hanya dua pekerjaan, tab dan harga
	 * (DESIGN.md, Secondary; Three-Percent Rule; Tags), dan keadaan sehat adalah
	 * keadaan tenang.
	 */
	const TAG_TENANG =
		'inline-flex h-[15px] items-center border border-border bg-muted px-[5px] text-xs font-semibold';
	/**
	 * Tag bergaris tinta untuk keadaan tidak sehat: garis dan kata, bukan merah
	 * (DESIGN.md, Do's: "write every state's word next to its mark").
	 */
	const TAG_TINTA =
		'inline-flex h-[15px] items-center border border-foreground bg-card px-[5px] text-xs font-semibold';
	/**
	 * Baris pesan bertinta, bukan merah utilitas: merah milik tab dan harga, jadi
	 * pesan galat pun ditulis sebagai tinta (DESIGN.md, Fields & Inputs).
	 */
	const PESAN = 'text-xs font-semibold';
	/**
	 * Tombol berbingkai memakai `variant="ghost"` lebih dulu: varian itu sudah
	 * membawa `hover:bg-muted`, persis latar hover dunia ini, dan tidak membawa apa
	 * pun yang harus dilawan (solution doc §3). Ukurannya `.btn` dunia ini: 26px,
	 * 13px/600, keadaan mati tidak dipudarkan.
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
</script>

<section class={MODUL} aria-labelledby="status-layanan-judul" aria-busy={health.isPending}>
	<div class={MODUL_KEPALA}>
		<h2 id="status-layanan-judul" class={MODUL_JUDUL}>Status layanan</h2>
		<span class={CATATAN}>diperiksa dari API Go</span>
	</div>

	{#if health.isPending}
		<p class={MODUL_TEKS}>Memeriksa…</p>
	{:else if health.error}
		<div class={MODUL_BARIS}>
			<p class={PESAN} role="alert">{health.error.message}</p>
			<Button variant="ghost" class={cn(AKSI_MODUL, 'mt-1.5')} onclick={() => void health.refetch()}
				>Coba lagi</Button
			>
		</div>
	{:else if health.data}
		<div class={MODUL_BARIS}>
			<!--
				Keadaannya membawa katanya: `OK` atau `DEGRADED` di tag, lalu satu
				kalimat yang menjelaskan apa artinya (DESIGN.md, Do's: "write every
				state's word next to its mark").
			-->
			<span class={isHealthy ? TAG_TENANG : TAG_TINTA}>
				{isHealthy ? 'OK' : 'DEGRADED'}
			</span>
			<p class="mt-1.5 text-xs">
				{#if isHealthy}
					Basis data: {health.data.database} — layanan siap dipakai.
				{:else}
					Basis data: {health.data.database} — layanan tidak sehat, periksa basis data.
				{/if}
			</p>
		</div>
	{/if}
</section>
