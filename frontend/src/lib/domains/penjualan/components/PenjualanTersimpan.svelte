<script lang="ts">
	import {
		AKSI_MODUL,
		ERROR,
		MODUL,
		MODUL_JUDUL,
		MODUL_KEPALA,
		TAG_MERAH
	} from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';
	import { createPenjualanDetailQuery } from '../queries/penjualan.queries';
	import type { NomorStruk } from '../schemas/penjualan.schema';
	import CetakStruk from './CetakStruk.svelte';
	import RincianPenjualan from './RincianPenjualan.svelte';

	/**
	 * One stored Penjualan, read by its Nomor Struk. The three states of §11 are
	 * explicit: a word while Go is answering, the message Go gave on a failure
	 * (a Nomor Struk that names nothing is its readable 404, "Penjualan tidak
	 * ditemukan."), and the record itself.
	 *
	 * The record carries the reprint: the way back to a Struk that was lost, and the
	 * one the Kasir presses after a print that failed (ADR-0017, keputusan 5).
	 */
	let { nomorStruk }: { nomorStruk: NomorStruk } = $props();

	const detail = createPenjualanDetailQuery(() => nomorStruk);

	/**
	 * Modul Penjualan tersimpan: prototipe memisahkannya dari modul Cari Penjualan
	 * dengan `margin-top: 8px` — keduanya memang tidak bersebelahan, jadi keduanya
	 * tidak berbagi garis (DESIGN.md, Shared-Hairline).
	 */
	const MODUL_PENJUALAN = cn(MODUL, 'mt-2');
	/**
	 * Baris pesan — galat maupun keadaan — bertinta, bukan merah utilitas: merah
	 * hanya dua pekerjaan, tab dan harga (DESIGN.md, Secondary & Three-Percent
	 * Rule).
	 */
	const PESAN = ERROR;
</script>

{#if detail.isPending}
	<div class="mt-2 border border-border bg-card px-2 py-1.5" aria-busy="true">
		<span class="sr-only" role="status">Mencari Penjualan…</span>
		<p class="text-xs">Memuat Penjualan…</p>
	</div>
{:else if detail.error}
	<div class="mt-2 border border-border bg-card px-2 py-1.5">
		<p class={PESAN} role="alert">{detail.error.message}</p>
		<Button variant="ghost" class={cn(AKSI_MODUL, 'mt-1.5')} onclick={() => void detail.refetch()}
			>Coba lagi</Button
		>
	</div>
{:else if detail.data}
	<section class={MODUL_PENJUALAN} aria-label="Penjualan tersimpan">
		<div class={MODUL_KEPALA}>
			<h2 class={MODUL_JUDUL}>Penjualan tersimpan</h2>
			<!--
				Tag merahnya ditaruh di kepala modul, persis prototipe: satu kata di
				sebelah judul, bukan baris penjelas yang memakan tinggi baris.
			-->
			<span class={TAG_MERAH}>tersegel</span>
		</div>

		<!--
			Rincian yang dibuka memakai komponen yang sudah diporting saat layar
			Kasir, bukan markah baru yang mirip — satu bentuk, dua layar.
		-->
		<RincianPenjualan sale={detail.data} />

		<!--
			Keyed by the Nomor Struk: a second lookup is a different sale, and the
			print result of the first must not follow it onto the screen.
		-->
		{#key nomorStruk}
			<CetakStruk {nomorStruk} />
		{/key}
	</section>
{/if}
