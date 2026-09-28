<script lang="ts">
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
	const MODUL = 'mt-2 border border-border border-b-0 bg-card';
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/**
	 * Red tag `tersegel` (DESIGN.md, Tags): Penjualan yang sudah tersimpan bersifat
	 * final — tidak ada void dan tidak ada refund — dan tab merah inilah yang
	 * mengatakannya, bukan kalimat penjelas (CONTEXT.md). Merah bukan warna
	 * sendirian: katanya ikut tertulis.
	 */
	const TAG_MERAH =
		'inline-flex h-[15px] items-center bg-destructive px-[5px] text-xs font-semibold text-primary-foreground';
	/**
	 * Tombol berbingkai memakai `variant="ghost"` lebih dulu: varian itu sudah
	 * membawa `hover:bg-muted`, persis latar hover dunia ini, dan tidak membawa
	 * apa pun yang harus dilawan (solution doc §3).
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
	/**
	 * Baris pesan — galat maupun keadaan — bertinta, bukan merah utilitas: merah
	 * hanya dua pekerjaan, tab dan harga (DESIGN.md, Secondary & Three-Percent
	 * Rule).
	 */
	const PESAN = 'text-xs font-semibold';
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
	<section class={MODUL} aria-label="Penjualan tersimpan">
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
