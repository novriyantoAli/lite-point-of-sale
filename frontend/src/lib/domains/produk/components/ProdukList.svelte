<script lang="ts">
	import {
		AKSI_BARIS,
		AKSI_MODUL,
		AKSI_STRIP,
		CATATAN,
		CATATAN_ANGKA,
		COMMIT_BARIS,
		CORET,
		FIELD,
		INPUT,
		JUDUL,
		LABEL,
		MODUL,
		MODUL_JUDUL,
		MODUL_KEPALA,
		PAPAN,
		STRIP,
		TAG
	} from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { cn, formatRupiah } from '$lib/utils';
	import ProdukForm from './ProdukForm.svelte';
	import {
		createDeleteProdukMutation,
		createKategoriListQuery,
		createProdukListQuery,
		createSetProdukActiveMutation,
		createStokMenipisQuery
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
	/**
	 * Produk mana yang Stoknya menipis, menurut aturan domain — bukan menurut ambang
	 * yang dikarang layar ini.
	 *
	 * Ambangnya milik satu toko dan datang bersama daftar itu sendiri, jadi tidak ada
	 * tempat kedua yang perlu diberi tahu kalau Admin mengubahnya di Pengaturan. Dan
	 * karena query yang sama juga yang dipakai layar Stok, katalog ini dan daftar
	 * Stok menipis tidak bisa berbeda pendapat soal Produk yang mana yang menipis
	 * (produk.schema.ts, StokMenipisSchema).
	 */
	const menipis = createStokMenipisQuery();
	const menipisIds = $derived(new Set((menipis.data?.products ?? []).map((p) => p.id)));

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
	 * "Semua Kategori" is a sentinel. These two lines are the only place it is
	 * translated, in either direction, so they cannot drift apart.
	 */
	const SEMUA_KATEGORI = 'semua';
	const kategoriValue = $derived(produkFilterState.category || SEMUA_KATEGORI);

	/**
	 * Status is three states, not two: "Semua" is not the same as "Aktif". One
	 * table holds the whole mapping — the label the trigger shows, and the filter
	 * the state keeps — so the Select's value and the state cannot disagree about
	 * what "nonaktif" means.
	 */
	const STATUS_OPTIONS = [
		{ value: 'semua', label: 'Semua', filter: null },
		{ value: 'aktif', label: 'Aktif', filter: true },
		{ value: 'nonaktif', label: 'Nonaktif', filter: false }
	] as const;

	type StatusValue = (typeof STATUS_OPTIONS)[number]['value'];

	const statusOption = $derived(
		STATUS_OPTIONS.find((option) => option.filter === produkFilterState.active) ?? STATUS_OPTIONS[0]
	);
	const statusValue: StatusValue = $derived(statusOption.value);
	const statusLabel = $derived(statusOption.label);

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

	/** The Produk the catalogue is showing right now. */
	const found = $derived(list.data ?? []);

	/**
	 * Skeleton baris selama memuat: mosaik yang sama, pada jumlah kolom yang sama,
	 * supaya papan tidak melompat saat datanya tiba — dan supaya yang terlihat
	 * bukan layar kosong yang tak bisa dibedakan dari katalog yang memang kosong.
	 */
	const BARIS_HANTU = [0, 1, 2, 3, 4, 5, 6, 7];
	const KOLOM_HANTU = [0, 1, 2, 3, 4, 5, 6];

	function setStatus(value: string) {
		produkFilterState.active =
			STATUS_OPTIONS.find((option) => option.value === value)?.filter ?? null;
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

	/**
	 * What the Stok column writes, in the three states CONTEXT.md gives the word for:
	 * Habis is zero, menipis is below the store's ambang, and anything else is just
	 * the number. Written as one expression on purpose — a chain of `{#if}` blocks
	 * would put the number in a text node of its own, and the e2e suite reads that
	 * cell as one exact name.
	 */
	function stokCell(produk: Produk): string {
		if (produk.stock === 0) return 'Stok habis';
		return menipisIds.has(produk.id) ? `${produk.stock} · menipis` : String(produk.stock);
	}

	/**
	 * Jarak 8px sebelum modul Katalog bukan kelalaian: prototipe layar ini sendiri
	 * yang memisahkan dua modul dengan `margin-top: 8px`, jadi keduanya memang tidak
	 * bersebelahan.
	 */
	const MODUL_KATALOG = cn(MODUL, 'mt-2');
	/**
	 * Kisi field empat kolom; menumpuk di 900px, bukan di titik papan 1080px
	 * (ADR-0020).
	 */
	const FORMGRID = 'grid grid-cols-1 gap-x-3 gap-y-[10px] min-[901px]:grid-cols-4';
	const FIELD_SELECT = cn(
		INPUT,
		'w-full data-[size=default]:h-[26px] [&_svg:not([class*=size-])]:size-3'
	);
	/** Tabel: satu garis rambut bersama antar baris, kepala Wash Grey + garis tinta. */
	const TABEL = 'w-full border-collapse text-[13px] leading-[1.25]';
	const TH =
		'border-b border-foreground bg-muted px-2 py-1.5 text-left text-xs font-semibold whitespace-nowrap';
	const TH_NUM = `${TH} text-right tabular-nums`;
	const BARIS = 'border-b border-border transition-colors last:border-b-0 hover:bg-muted';
	const TD = 'px-2 py-1.5 align-top';
	const TD_NUM = `${TD} text-right tabular-nums`;
	/** Harga merah utilitas dan tabular; tidak pernah monospace (DESIGN.md, Typography). */
	const TD_HARGA = `${TD_NUM} font-semibold text-destructive`;
	const TD_SUB = `${TD} text-xs`;
	const TD_AKSI = `${TD} whitespace-nowrap`;
	/**
	 * Quiet tag yang mati: garis coretnya, bukan opasitasnya (DESIGN.md,
	 * State-Is-Not-Faded).
	 */
	const TAG_NONAKTIF = `${TAG} ${CORET}`;
</script>

<!--
	Kepala tabel ditulis sekali dan dirender dari dua tempat — katalog yang memuat
	dan katalog yang sudah tiba. Dua salinan tujuh kolom akan berbeda sendiri dalam
	dua port (solution doc §5).
-->
{#snippet kepalaTabel()}
	<thead>
		<tr>
			<th scope="col" class={TH}>Nama</th>
			<th scope="col" class={TH}>Kode</th>
			<th scope="col" class={TH}>Kategori</th>
			<th scope="col" class={TH_NUM}>Harga</th>
			<th scope="col" class={TH_NUM}>Stok</th>
			<th scope="col" class={TH}>Status</th>
			<th scope="col" class={TH}>Aksi</th>
		</tr>
	</thead>
{/snippet}

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya, dan
		aksi utama layar ini di baris yang sama (DESIGN.md, Typography).
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Produk</h1>
		<p class={CATATAN}>
			Katalog yang dijual di kasir. Produk Nonaktif tidak muncul di lookup kasir, tetapi riwayatnya
			tetap tersimpan.
		</p>
		<span class={AKSI_STRIP}>
			<Button class={COMMIT_BARIS} onclick={startAdding}>Tambah Produk</Button>
		</span>
	</div>

	<section class={MODUL} role="search" aria-labelledby="saring-judul">
		<div class={MODUL_KEPALA}>
			<h2 id="saring-judul" class={MODUL_JUDUL}>Saring katalog</h2>
			{#if narrowed}
				<!--
					Saringan yang sudah menyempit tidak perlu lagi diberi tahu bahwa ia
					berlaku saat mengetik; yang ia butuhkan adalah jalan kembali.
				-->
				<Button variant="ghost" class={AKSI_MODUL} onclick={() => produkFilterState.reset()}
					>Bersihkan saringan</Button
				>
			{:else}
				<span class={CATATAN}>saringan berlaku saat mengetik</span>
			{/if}
		</div>

		<div class="border-b border-border px-2 py-1.5">
			<div class={FORMGRID}>
				<div class={FIELD}>
					<Label for="saring-nama" class={LABEL}>Nama</Label>
					<Input
						id="saring-nama"
						autocomplete="off"
						class={INPUT}
						bind:value={produkFilterState.name}
					/>
				</div>

				<div class={FIELD}>
					<Label for="saring-kode" class={LABEL}>Kode</Label>
					<Input
						id="saring-kode"
						autocomplete="off"
						class={INPUT}
						bind:value={produkFilterState.code}
					/>
				</div>

				<div class={FIELD}>
					<Label for="saring-kategori" class={LABEL}>Kategori</Label>
					<Select.Root type="single" value={kategoriValue} onValueChange={setKategori}>
						<Select.Trigger id="saring-kategori" class={FIELD_SELECT}>
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
						<Select.Content
							class="rounded-none border border-border shadow-none ring-0 ring-transparent"
						>
							<Select.Item value={SEMUA_KATEGORI} label="Semua Kategori">
								Semua Kategori
							</Select.Item>
							{#each kategori.data ?? [] as name (name)}
								<Select.Item value={name} label={name}>{name}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class={FIELD}>
					<Label for="saring-status" class={LABEL}>Status</Label>
					<Select.Root type="single" value={statusValue} onValueChange={setStatus}>
						<Select.Trigger id="saring-status" class={FIELD_SELECT}>
							<span data-slot="select-value">{statusLabel}</span>
						</Select.Trigger>
						<Select.Content
							class="rounded-none border border-border shadow-none ring-0 ring-transparent"
						>
							{#each STATUS_OPTIONS as option (option.value)}
								<Select.Item value={option.value} label={option.label}>
									{option.label}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
			</div>
		</div>
	</section>

	<section class={MODUL_KATALOG} aria-labelledby="katalog-judul" aria-busy={list.isPending}>
		<div class={MODUL_KEPALA}>
			<h2 id="katalog-judul" class={MODUL_JUDUL}>Katalog</h2>
			<!--
				Cacah modul ini datang dari daftar yang sudah dimuat, bukan permintaan
				kedua: cacah yang belum diputuskan adalah cacah tab di rel, bukan yang
				ini (DESIGN.md, Tabs).
			-->
			<span class={CATATAN_ANGKA}>
				{list.isPending ? 'Memuat katalog…' : `${found.length} Produk`}
			</span>
		</div>

		{#if notice}
			<p class="border-b border-border px-2 py-1.5 text-xs font-semibold" role="status">
				{notice}
			</p>
		{/if}

		{#if list.isPending}
			<!--
				Kerangkanya bukan lingkaran berputar: baris-baris sepanjang kolom yang
				sama, pada jumlah baris yang masuk akal, jadi kepala tabel di atasnya
				tidak bergerak saat datanya tiba.
			-->
			<div class="overflow-x-auto border-b border-border">
				<table class={TABEL}>
					{@render kepalaTabel()}
					<tbody aria-hidden="true">
						{#each BARIS_HANTU as baris (baris)}
							<tr class="border-b border-border last:border-b-0">
								{#each KOLOM_HANTU as kolom (kolom)}
									<td class="px-2 py-1.5">
										<span class="block h-[13px] w-full bg-muted"></span>
									</td>
								{/each}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if list.error}
			<div class="border-b border-border px-2 py-1.5">
				<p class="text-xs font-semibold" role="alert">{list.error.message}</p>
				<Button variant="ghost" class={cn(AKSI_MODUL, 'mt-1.5')} onclick={() => void list.refetch()}
					>Coba lagi</Button
				>
			</div>
		{:else if found.length === 0}
			<p class="border-b border-border px-2 py-1.5 text-xs">
				{narrowed
					? 'Tidak ada Produk yang cocok dengan saringan ini.'
					: 'Belum ada Produk. Tambahkan yang pertama lewat tombol Tambah Produk.'}
			</p>
		{:else}
			<div class="overflow-x-auto border-b border-border">
				<table class={TABEL}>
					{@render kepalaTabel()}
					<tbody>
						{#each found as produk (produk.id)}
							<tr class={BARIS}>
								<td class={`${TD} font-medium`}>{produk.name}</td>

								<!--
									Kode yang kosong tidak boleh terlihat seperti data yang hilang:
									ia menuliskan keadaannya (DESIGN.md, Do's).
								-->
								<td class={TD_SUB}>
									{#if produk.code}
										<span class="tabular-nums">{produk.code}</span>
									{:else}
										tanpa Kode
									{/if}
								</td>

								<td class={TD_SUB}>{produk.category ?? '—'}</td>

								<td class={TD_HARGA}>{formatRupiah(produk.price)}</td>

								<!--
									Angkanya sendiri, atau kata yang menemani keadaannya — prototipe
									menulis "4 · menipis" dan "Stok habis" di kolom yang sama.
								-->
								<td class={TD_NUM}>{stokCell(produk)}</td>

								<td class={TD}>
									<span class={produk.active ? TAG : TAG_NONAKTIF}>
										{produk.active ? 'Aktif' : 'Nonaktif'}
									</span>
								</td>

								<td class={TD_AKSI}>
									{#if confirmingDeleteId === produk.id}
										<!--
											Hapus tidak bisa dibatalkan, jadi ia menempuh dua klik yang
											disengaja; kedua katanya berdiri di barisnya sendiri, bukan
											muncul saat kursor lewat (DESIGN.md, Named-Not-Hidden).
										-->
										<div class="flex flex-wrap items-center gap-1.5">
											<span class="text-xs">Hapus permanen?</span>
											<Button
												variant="ghost"
												class={AKSI_BARIS}
												disabled={remove.isPending}
												onclick={() => void confirmDelete(produk)}
											>
												Ya, hapus
											</Button>
											<Button
												variant="ghost"
												class={AKSI_BARIS}
												onclick={() => (confirmingDeleteId = null)}
												aria-label="Batal menghapus"
											>
												Batal
											</Button>
										</div>
									{:else}
										<div class="flex flex-wrap items-center gap-1.5">
											<Button
												variant="ghost"
												class={AKSI_BARIS}
												onclick={() => startEditing(produk)}>Ubah</Button
											>
											<Button
												variant="ghost"
												class={AKSI_BARIS}
												disabled={setActive.isPending}
												onclick={() => void toggleActive(produk)}
											>
												{produk.active ? 'Nonaktifkan' : 'Aktifkan'}
											</Button>
											{#if produk.sold}
												<!-- A Produk that sold has a history to keep, so the API
												     would refuse a delete; the UI does not offer one. -->
												<span class="text-xs">Pernah terjual</span>
											{:else}
												<Button
													variant="ghost"
													class={AKSI_BARIS}
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
	</section>

	{#if setActive.error}
		<p
			class="border border-t-0 border-border bg-card px-2 py-1.5 text-xs font-semibold"
			role="alert"
		>
			{setActive.error.message}
		</p>
	{/if}
	{#if remove.error}
		<p
			class="border border-t-0 border-border bg-card px-2 py-1.5 text-xs font-semibold"
			role="alert"
		>
			{remove.error.message}
		</p>
	{/if}
</div>

{#if adding || editing}
	{#key editing?.id ?? 'baru'}
		<ProdukForm produk={editing ?? undefined} onSaved={saved} onCancel={closeForm} />
	{/key}
{/if}
