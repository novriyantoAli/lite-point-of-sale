<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { createProdukListQuery, type Produk } from '$lib/domains/produk';
	import { formatRupiah } from '$lib/utils';

	/**
	 * The till's Produk lookup: the Kasir finds a Produk by scanning or typing its
	 * Kode, or by searching its name (CONTEXT.md, Kode).
	 *
	 * The two filters are two fields rather than one because the API narrows a
	 * listing by `code` *and* `name` at once — one field would have to send both
	 * and so require a Produk to match on both. Two fields is also what the AC
	 * asks for: found via Kode, and found via name.
	 *
	 * The filters are component state, not the produk domain's `produkFilterState`:
	 * that belongs to the catalogue screen, and reaching into another domain's
	 * `state/` is exactly what the barrel rule forbids (ADR-0006).
	 */
	let { onAdd }: { onAdd: (produk: Produk) => void } = $props();

	let code = $state('');
	let name = $state('');

	/**
	 * `active: true` is the whole of "the kasir lookup": a Produk that is Nonaktif
	 * is not for sale, so it is not offered (CONTEXT.md, Nonaktif).
	 */
	const list = createProdukListQuery(() => ({ name, code, category: '', active: true }));

	const found = $derived(list.data ?? []);
	const narrowed = $derived(Boolean(code || name));

	function add(produk: Produk) {
		onAdd(produk);
	}

	/**
	 * Enter in either field must not navigate: this form narrows a listing, it does
	 * not submit one. The add itself is the Tambah button — a scan-and-Enter that
	 * guessed at the results would be guessing at a listing that may not have come
	 * back yet.
	 */
	function search(event: SubmitEvent) {
		event.preventDefault();
	}
</script>

<section class="space-y-4" aria-labelledby="kasir-cari-judul">
	<h2 id="kasir-cari-judul" class="font-medium">Cari Produk</h2>

	<form
		class="grid gap-4 rounded-lg border p-4 sm:grid-cols-2"
		role="search"
		aria-label="Cari Produk"
		onsubmit={search}
	>
		<div class="space-y-2">
			<Label for="kasir-kode">Kode</Label>
			<!-- autofocus: a barcode scanner types into whatever has focus, and the
			     Kode field is what narrows the listing to the scanned Produk. -->
			<Input id="kasir-kode" name="code" autocomplete="off" autofocus bind:value={code} />
			<p class="text-sm text-muted-foreground">
				Scan barcode atau ketik Kode — daftar di bawah menyaring ke Produk yang cocok.
			</p>
		</div>

		<div class="space-y-2">
			<Label for="kasir-nama">Nama</Label>
			<Input id="kasir-nama" name="name" autocomplete="off" bind:value={name} />
			<p class="text-sm text-muted-foreground">
				Produk tanpa Kode tetap bisa dijual — cari lewat namanya.
			</p>
		</div>
	</form>

	{#if list.isPending}
		<p class="text-sm text-muted-foreground">Memuat Produk…</p>
	{:else if list.error}
		<div class="space-y-3">
			<p class="text-sm text-destructive">{list.error.message}</p>
			<Button variant="outline" size="sm" onclick={() => void list.refetch()}>Coba lagi</Button>
		</div>
	{:else if found.length === 0}
		<p class="text-sm text-muted-foreground">
			{narrowed
				? 'Tidak ada Produk Aktif yang cocok. Periksa Kode atau namanya.'
				: 'Belum ada Produk yang bisa dijual. Admin perlu menambahkannya lebih dulu.'}
		</p>
	{:else}
		<ul class="space-y-2" aria-label="Hasil pencarian Produk">
			{#each found as produk (produk.id)}
				<li class="flex flex-wrap items-center justify-between gap-3 rounded-lg border p-3">
					<div class="space-y-1">
						<p class="font-medium">
							{produk.name}
							{#if produk.code}
								<span class="text-muted-foreground">· {produk.code}</span>
							{/if}
						</p>
						<p class="text-sm text-muted-foreground">
							{formatRupiah(produk.price)}
							·
							{produk.stock === 0 ? 'Stok habis' : `Stok ${produk.stock}`}
						</p>
					</div>
					<Button
						size="sm"
						disabled={produk.stock === 0}
						onclick={() => add(produk)}
						aria-label={`Tambah ${produk.name} ke keranjang`}
					>
						Tambah
					</Button>
				</li>
			{/each}
		</ul>
	{/if}
</section>
