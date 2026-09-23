<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import TambahStokForm from './TambahStokForm.svelte';
	import { createProdukListQuery, createStokMenipisQuery } from '../queries/produk.queries';
	import type { Produk } from '../schemas/produk.schema';

	/**
	 * The Stok screen. Two lists answer two different questions — "what do I have
	 * to restock now" and "where does every Produk stand" — and both offer the same
	 * restock form, because an Admin who came for one of them should not have to
	 * walk over to the other to act on it.
	 *
	 * Every Produk, unfiltered: this is not the catalogue and does not own the
	 * catalogue's filters, so it asks for the whole list rather than reading
	 * `produkFilterState`. An Admin who narrowed the catalogue to one Kategori must
	 * not find half their Stok missing here.
	 */
	const catalogue = createProdukListQuery(() => ({ name: '', code: '', category: '' }));
	/** What to restock, with the threshold that selected it (the domain's rule). */
	const menipis = createStokMenipisQuery();

	/**
	 * The row whose restock form is open. The Produk alone is not enough to name it:
	 * the same Produk appears in both lists, and a form that opened in both at once
	 * would be two inputs for one delivery — the Admin would type into one and the
	 * other would keep a stale amount.
	 */
	type RestockPlace = 'menipis' | 'katalog';
	let restocking = $state<{ place: RestockPlace; id: number } | null>(null);
	let notice = $state('');

	function isRestocking(produk: Produk, place: RestockPlace): boolean {
		return restocking?.place === place && restocking.id === produk.id;
	}

	function startRestocking(produk: Produk, place: RestockPlace) {
		restocking = { place, id: produk.id };
		notice = '';
	}

	function restocked(produk: Produk) {
		notice = `Stok ${produk.name} sekarang ${produk.stock}.`;
		restocking = null;
	}
</script>

{#snippet restock(produk: Produk, place: RestockPlace)}
	{#if isRestocking(produk, place)}
		<TambahStokForm {produk} onAdded={restocked} onCancel={() => (restocking = null)} />
	{:else}
		<Button variant="outline" size="sm" onclick={() => startRestocking(produk, place)}>
			Tambah Stok
		</Button>
	{/if}
{/snippet}

<div class="space-y-8">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Stok</h1>
		<p class="text-sm text-muted-foreground">
			Catat barang masuk dan lihat Produk mana yang perlu ditambah. Stok berkurang sendiri saat
			Produk terjual.
		</p>
	</div>

	{#if notice}
		<p class="text-sm text-muted-foreground" role="status">{notice}</p>
	{/if}

	<section class="space-y-3" aria-labelledby="stok-menipis-judul">
		<div class="space-y-1">
			<h2 id="stok-menipis-judul" class="font-medium">Stok menipis</h2>
			<p class="text-sm text-muted-foreground">
				{#if menipis.data}
					Produk Aktif dengan Stok di bawah {menipis.data.threshold}, yang paling sedikit di atas.
					Produk Nonaktif tidak dihitung — ia tidak sedang dijual.
				{:else}
					Produk Aktif dengan Stok paling sedikit, yang perlu ditambah.
				{/if}
			</p>
		</div>

		{#if menipis.isPending}
			<p class="text-sm text-muted-foreground">Memuat Stok menipis…</p>
		{:else if menipis.error}
			<div class="space-y-3">
				<p class="text-sm text-destructive">{menipis.error.message}</p>
				<Button variant="outline" size="sm" onclick={() => void menipis.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if menipis.data?.products.length === 0}
			<p class="text-sm text-muted-foreground">
				Tidak ada Produk dengan Stok menipis. Semua Produk Aktif masih punya Stok di ambang atau
				lebih.
			</p>
		{:else}
			<ul class="space-y-2">
				{#each menipis.data?.products ?? [] as produk (produk.id)}
					<li class="flex flex-wrap items-center justify-between gap-3 rounded-lg border p-3">
						<div class="space-y-1">
							<p class="font-medium">
								{produk.name}
								{#if produk.code}
									<span class="text-muted-foreground">· {produk.code}</span>
								{/if}
							</p>
							<p class="text-sm text-muted-foreground">
								{produk.stock === 0 ? 'Stok habis' : `Sisa ${produk.stock}`}
							</p>
						</div>
						{@render restock(produk, 'menipis')}
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section class="space-y-3" aria-labelledby="stok-produk-judul">
		<h2 id="stok-produk-judul" class="font-medium">Stok per Produk</h2>

		{#if catalogue.isPending}
			<p class="text-sm text-muted-foreground">Memuat Produk…</p>
		{:else if catalogue.error}
			<div class="space-y-3">
				<p class="text-sm text-destructive">{catalogue.error.message}</p>
				<Button variant="outline" size="sm" onclick={() => void catalogue.refetch()}>
					Coba lagi
				</Button>
			</div>
		{:else if catalogue.data?.length === 0}
			<p class="text-sm text-muted-foreground">
				Belum ada Produk. Tambahkan yang pertama di layar Produk.
			</p>
		{:else}
			<div class="overflow-x-auto rounded-lg border">
				<table class="w-full text-sm">
					<thead class="border-b bg-muted/50 text-left">
						<tr>
							<th scope="col" class="p-3 font-medium">Nama</th>
							<th scope="col" class="p-3 font-medium">Kode</th>
							<th scope="col" class="p-3 text-right font-medium">Stok</th>
							<th scope="col" class="p-3 font-medium">Keadaan</th>
							<th scope="col" class="p-3 font-medium">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each catalogue.data ?? [] as produk (produk.id)}
							<tr>
								<td class="p-3 font-medium">{produk.name}</td>
								<td class="p-3 text-muted-foreground">{produk.code ?? '—'}</td>
								<td class="p-3 text-right tabular-nums">{produk.stock}</td>
								<td class="p-3">
									{#if !produk.active}
										<!-- A Produk that is not for sale is not one that runs out, so
										     it is not called menipis here either (CONTEXT.md, Nonaktif). -->
										<Badge variant="outline">Nonaktif</Badge>
									{:else if produk.stock === 0}
										<Badge variant="destructive">Habis</Badge>
									{:else if menipis.data === undefined}
										<!--
											The threshold is what "menipis" means, so without it this screen
											cannot call anything aman — that would be a guess, and the section
											above is already saying the list could not be read.
										-->
										<span class="text-muted-foreground">—</span>
									{:else if produk.stock < menipis.data.threshold}
										<Badge variant="secondary">Menipis</Badge>
									{:else}
										<span class="text-muted-foreground">Aman</span>
									{/if}
								</td>
								<td class="p-3">{@render restock(produk, 'katalog')}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>
</div>
