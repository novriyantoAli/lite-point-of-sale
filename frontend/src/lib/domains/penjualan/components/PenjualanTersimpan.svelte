<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { createPenjualanDetailQuery } from '../queries/penjualan.queries';
	import type { NomorStruk } from '../schemas/penjualan.schema';
	import RincianPenjualan from './RincianPenjualan.svelte';

	/**
	 * One stored Penjualan, read by its Nomor Struk. The three states of §11 are
	 * explicit: a skeleton while Go is answering, the message Go gave on a failure
	 * (a Nomor Struk that names nothing is its readable 404, "Penjualan tidak
	 * ditemukan."), and the record itself.
	 */
	let { nomorStruk }: { nomorStruk: NomorStruk } = $props();

	const detail = createPenjualanDetailQuery(() => nomorStruk);
</script>

{#if detail.isPending}
	<div class="space-y-3" role="status">
		<span class="sr-only">Mencari Penjualan…</span>
		<div class="h-4 w-2/3 animate-pulse rounded bg-muted"></div>
		<div class="h-4 w-1/2 animate-pulse rounded bg-muted"></div>
		<div class="h-4 w-3/4 animate-pulse rounded bg-muted"></div>
	</div>
{:else if detail.error}
	<div class="space-y-3">
		<p class="text-sm text-destructive" role="alert">{detail.error.message}</p>
		<Button variant="outline" size="sm" onclick={() => void detail.refetch()}>Coba lagi</Button>
	</div>
{:else if detail.data}
	<section class="rounded-lg border p-4" aria-label="Penjualan tersimpan">
		<RincianPenjualan sale={detail.data} />
	</section>
{/if}
