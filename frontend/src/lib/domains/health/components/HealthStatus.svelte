<script lang="ts">
	import {
		AKSI_MODUL,
		CATATAN,
		ERROR,
		MODUL,
		MODUL_BARIS,
		MODUL_JUDUL,
		MODUL_KEPALA,
		MODUL_TEKS,
		TAG_TENANG,
		TAG_TINTA
	} from '$lib/components/shared/mosaik';
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

	/**
	 * Baris pesan bertinta, bukan merah utilitas: merah milik tab dan harga, jadi
	 * pesan galat pun ditulis sebagai tinta (DESIGN.md, Fields & Inputs).
	 */
	const PESAN = ERROR;
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
