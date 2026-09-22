<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { formatRupiah } from '$lib/utils';
	import ProdukForm from './ProdukForm.svelte';
	import {
		createDeleteProdukMutation,
		createKategoriListQuery,
		createProdukListQuery,
		createSetProdukActiveMutation
	} from '../queries/produk.queries';
	import type { Produk } from '../schemas/produk.schema';
	import { produkFilterState } from '../state/produk.state.svelte';

	/**
	 * The catalogue screen. It reads the filters from `produkFilterState` — UI
	 * state the Admin owns, not server data — and asks for the matching Produk.
	 * The thunk keeps the query reactive to that state, so the list follows the
	 * filter bar as it is typed into.
	 */
	const list = createProdukListQuery(() => produkFilterState.filter);
	const kategori = createKategoriListQuery();
	const setActive = createSetProdukActiveMutation();
	const remove = createDeleteProdukMutation();

	/** Which form is open: adding, changing one Produk, or neither. */
	let adding = $state(false);
	let editing = $state<Produk | null>(null);
	/**
	 * The row that has been clicked once and is waiting for a second click.
	 * Deleting a Produk cannot be undone, so it takes two deliberate clicks.
	 */
	let confirmingDeleteId = $state<number | null>(null);
	let notice = $state('');

	/**
	 * The empty string the filter state uses is not a value a Select can hold, so
	 * "Semua Kategori" is a sentinel that is mapped back to '' on the way in.
	 */
	const SEMUA_KATEGORI = 'semua';
	const kategoriValue = $derived(produkFilterState.category || SEMUA_KATEGORI);

	/** Status is three states, not two: "Semua" is not the same as "Aktif". */
	const STATUS_OPTIONS = [
		{ value: 'semua', label: 'Semua' },
		{ value: 'aktif', label: 'Aktif' },
		{ value: 'nonaktif', label: 'Nonaktif' }
	] as const;

	type StatusValue = (typeof STATUS_OPTIONS)[number]['value'];

	const statusValue = $derived<StatusValue>(
		produkFilterState.active === null ? 'semua' : produkFilterState.active ? 'aktif' : 'nonaktif'
	);
	const statusLabel = $derived(
		STATUS_OPTIONS.find((option) => option.value === statusValue)?.label ?? 'Semua'
	);

	/**
	 * Whether the Admin narrowed anything. It decides which empty state to show:
	 * an empty catalogue is a different thing from a filter that matched nothing.
	 */
	const narrowed = $derived(
		Boolean(
			produkFilterState.name ||
			produkFilterState.code ||
			produkFilterState.category ||
			produkFilterState.active !== null
		)
	);

	function setStatus(value: string) {
		produkFilterState.active = value === 'aktif' ? true : value === 'nonaktif' ? false : null;
	}

	function setKategori(value: string) {
		produkFilterState.category = value === SEMUA_KATEGORI ? '' : value;
	}

	function startAdding() {
		adding = true;
		editing = null;
		notice = '';
	}

	function startEditing(produk: Produk) {
		editing = produk;
		adding = false;
		notice = '';
	}

	function closeForm() {
		adding = false;
		editing = null;
	}

	function saved(produk: Produk) {
		notice = `Produk ${produk.name} disimpan.`;
		closeForm();
	}

	async function toggleActive(produk: Produk) {
		notice = '';
		try {
			await setActive.mutateAsync({ id: produk.id, active: !produk.active });
		} catch {
			// setActive.error carries the normalized message, rendered below.
		}
	}

	async function confirmDelete(produk: Produk) {
		notice = '';
		try {
			await remove.mutateAsync(produk.id);
			notice = `Produk ${produk.name} dihapus.`;
		} catch {
			// remove.error carries the normalized message, rendered below.
		} finally {
			confirmingDeleteId = null;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex flex-wrap items-end justify-between gap-4">
		<div class="space-y-1">
			<h1 class="text-2xl font-semibold">Produk</h1>
			<p class="text-sm text-muted-foreground">
				Katalog yang dijual di kasir. Produk Nonaktif tidak muncul di lookup kasir, tetapi
				riwayatnya tetap tersimpan.
			</p>
		</div>

		<Button onclick={startAdding}>Tambah Produk</Button>
	</div>

	<div
		class="grid gap-4 rounded-lg border p-4 sm:grid-cols-4"
		role="search"
		aria-label="Saring katalog"
	>
		<div class="space-y-2">
			<Label for="saring-nama">Nama</Label>
			<Input id="saring-nama" autocomplete="off" bind:value={produkFilterState.name} />
		</div>

		<div class="space-y-2">
			<Label for="saring-kode">Kode</Label>
			<Input id="saring-kode" autocomplete="off" bind:value={produkFilterState.code} />
		</div>

		<div class="space-y-2">
			<Label for="saring-kategori">Kategori</Label>
			<Select.Root type="single" value={kategoriValue} onValueChange={setKategori}>
				<Select.Trigger id="saring-kategori" class="w-full">
					<!--
						The text is written here rather than left to `Select.Value`: the
						label registry is filled by items as they mount, and the content is
						lazily mounted, so before the dropdown is ever opened the trigger
						would show the raw value — including the `semua` sentinel above.
					-->
					<span data-slot="select-value">
						{produkFilterState.category || 'Semua Kategori'}
					</span>
				</Select.Trigger>
				<Select.Content>
					<Select.Item value={SEMUA_KATEGORI} label="Semua Kategori">Semua Kategori</Select.Item>
					{#each kategori.data ?? [] as name (name)}
						<Select.Item value={name} label={name}>{name}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<div class="space-y-2">
			<Label for="saring-status">Status</Label>
			<Select.Root type="single" value={statusValue} onValueChange={setStatus}>
				<Select.Trigger id="saring-status" class="w-full">
					<span data-slot="select-value">{statusLabel}</span>
				</Select.Trigger>
				<Select.Content>
					{#each STATUS_OPTIONS as option (option.value)}
						<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		{#if narrowed}
			<div class="sm:col-span-4">
				<Button variant="outline" size="sm" onclick={() => produkFilterState.reset()}>
					Bersihkan saringan
				</Button>
			</div>
		{/if}
	</div>

	{#if adding || editing}
		{#key editing?.id ?? 'baru'}
			<ProdukForm produk={editing ?? undefined} onSaved={saved} onCancel={closeForm} />
		{/key}
	{/if}

	{#if notice}
		<p class="text-sm text-muted-foreground" role="status">{notice}</p>
	{/if}

	{#if list.isPending}
		<p class="text-sm text-muted-foreground">Memuat katalog…</p>
	{:else if list.error}
		<div class="space-y-3">
			<p class="text-sm text-destructive">{list.error.message}</p>
			<Button variant="outline" size="sm" onclick={() => void list.refetch()}>Coba lagi</Button>
		</div>
	{:else if list.data?.length === 0}
		<p class="text-sm text-muted-foreground">
			{narrowed
				? 'Tidak ada Produk yang cocok dengan saringan ini.'
				: 'Belum ada Produk. Tambahkan yang pertama lewat tombol Tambah Produk.'}
		</p>
	{:else}
		<div class="overflow-x-auto rounded-lg border">
			<table class="w-full text-sm">
				<thead class="border-b bg-muted/50 text-left">
					<tr>
						<th scope="col" class="p-3 font-medium">Nama</th>
						<th scope="col" class="p-3 font-medium">Kode</th>
						<th scope="col" class="p-3 font-medium">Kategori</th>
						<th scope="col" class="p-3 text-right font-medium">Harga</th>
						<th scope="col" class="p-3 text-right font-medium">Stok</th>
						<th scope="col" class="p-3 font-medium">Status</th>
						<th scope="col" class="p-3 font-medium">Aksi</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each list.data ?? [] as produk (produk.id)}
						<tr>
							<td class="p-3 font-medium">{produk.name}</td>
							<td class="p-3 text-muted-foreground">{produk.code ?? '—'}</td>
							<td class="p-3 text-muted-foreground">{produk.category ?? '—'}</td>
							<td class="p-3 text-right tabular-nums">{formatRupiah(produk.price)}</td>
							<td class="p-3 text-right tabular-nums">{produk.stock}</td>
							<td class="p-3">
								{#if produk.active}
									<Badge variant="secondary">Aktif</Badge>
								{:else}
									<Badge variant="outline">Nonaktif</Badge>
								{/if}
							</td>
							<td class="p-3">
								{#if confirmingDeleteId === produk.id}
									<div class="flex items-center gap-2">
										<span class="text-muted-foreground">Hapus permanen?</span>
										<Button
											variant="destructive"
											size="sm"
											disabled={remove.isPending}
											onclick={() => void confirmDelete(produk)}
										>
											Ya, hapus
										</Button>
										<Button variant="outline" size="sm" onclick={() => (confirmingDeleteId = null)}>
											Batal
										</Button>
									</div>
								{:else}
									<div class="flex flex-wrap items-center gap-2">
										<Button variant="outline" size="sm" onclick={() => startEditing(produk)}>
											Ubah
										</Button>
										<Button
											variant="outline"
											size="sm"
											disabled={setActive.isPending}
											onclick={() => void toggleActive(produk)}
										>
											{produk.active ? 'Nonaktifkan' : 'Aktifkan'}
										</Button>
										{#if produk.sold}
											<!-- A Produk that sold has a history to keep, so the API
											     would refuse a delete; the UI does not offer one. -->
											<span class="text-muted-foreground">Pernah terjual</span>
										{:else}
											<Button
												variant="outline"
												size="sm"
												onclick={() => (confirmingDeleteId = produk.id)}
											>
												Hapus
											</Button>
										{/if}
									</div>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}

	{#if setActive.error}
		<p class="text-sm text-destructive" role="alert">{setActive.error.message}</p>
	{/if}
	{#if remove.error}
		<p class="text-sm text-destructive" role="alert">{remove.error.message}</p>
	{/if}
</div>
