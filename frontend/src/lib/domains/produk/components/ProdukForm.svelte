<script lang="ts">
	import { untrack } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { cn, collectFieldErrors } from '$lib/utils';
	import { createProdukMutation, createUpdateProdukMutation } from '../queries/produk.queries';
	import {
		CreateProdukInputSchema,
		UpdateProdukInputSchema,
		type Produk
	} from '../schemas/produk.schema';

	/**
	 * Adds a Produk or changes one. Adding fills in the Stok awal; changing edits the
	 * editable record only — nama, Kode, harga, Kategori — because an edit cannot
	 * move Stok (ADR-0014). The two shapes are two schemas, so the form never sends
	 * a field the API would have to drop.
	 *
	 * `produk` is the record being changed, or undefined to add a new one. The
	 * parent remounts this component per record (`{#key}`), and that is what gives
	 * the fields below their starting values: initializing from a prop that later
	 * changes would need an `$effect` writing state, which is how a form ends up
	 * fighting the person typing into it.
	 *
	 * Formulir ini berdiri di dalam `Dialog` karena ia bukan bagian dari papan yang
	 * sedang dibaca — ia pekerjaan yang mengambil alih layar sebentar, lalu pergi.
	 * Prototipe tidak menggambar satu pun dialog, jadi hiasannya diturunkan dari
	 * DESIGN.md: radius nol, tanpa bayangan, kepala Wash Grey ditutup garis tinta,
	 * Fields & Inputs, dan Commit Button (#33).
	 */
	let {
		produk,
		onSaved,
		onCancel
	}: {
		produk?: Produk;
		onSaved: (produk: Produk) => void;
		onCancel?: () => void;
	} = $props();

	const create = createProdukMutation();
	const update = createUpdateProdukMutation();

	/**
	 * The fields stay strings: a form submits text, and the schemas are the one
	 * place that turn that text into the integers the API takes (§11, and the reason
	 * `CreateProdukInputSchema` owns the parsing). Parsing here as well would be a
	 * second set of rules to keep in step.
	 */
	// The parent remounts this component per record, so the prop is deliberately
	// read only once, when the fields are initialized: `untrack` states that
	// instead of leaving the next reader to wonder whether it was an oversight.
	let name = $state(untrack(() => produk?.name ?? ''));
	let code = $state(untrack(() => produk?.code ?? ''));
	let price = $state(untrack(() => (produk === undefined ? '' : String(produk.price))));
	let category = $state(untrack(() => produk?.category ?? ''));
	/**
	 * The Stok awal of a Produk being added. It is not seeded from `produk.stock`
	 * and is not rendered when changing a Produk: an edit has no Stok to change, so
	 * there is no number here to send back over a delivery that arrived meanwhile.
	 */
	let stock = $state('');
	let fieldErrors = $state<Partial<Record<'name' | 'price' | 'stock', string>>>({});

	/**
	 * The Status a new Produk starts with. Aktif is the default because adding a
	 * Produk is how an Admin puts something on sale; Nonaktif is for entering
	 * something that is not for sale yet.
	 *
	 * The field is offered only when adding. An edit keeps the Produk's Status —
	 * changing it is what the Aktifkan/Nonaktifkan button in the catalogue is for —
	 * so this form never sends a Status it would not be allowed to decide.
	 */
	const STATUS_OPTIONS = [
		{ value: 'aktif', label: 'Aktif', active: true },
		{ value: 'nonaktif', label: 'Nonaktif', active: false }
	] as const;

	type StatusValue = (typeof STATUS_OPTIONS)[number]['value'];

	let status = $state<StatusValue>('aktif');
	const statusOption = $derived(
		STATUS_OPTIONS.find((option) => option.value === status) ?? STATUS_OPTIONS[0]
	);

	function setStatus(value: string) {
		status = STATUS_OPTIONS.find((option) => option.value === value)?.value ?? 'aktif';
	}

	/** The fields a person can get wrong; Kode and Kategori accept anything. */
	const fields = ['name', 'price', 'stock'] as const;

	const pending = $derived(create.isPending || update.isPending);
	const error = $derived(create.error ?? update.error);
	const editing = $derived(produk !== undefined);

	/**
	 * Escape and the overlay are the dialog's own ways out, and they mean the same
	 * thing as Batal. The parent owns whether this component is mounted at all, so
	 * there is no local `open` to flip: a request to close that nobody can answer
	 * would leave the form open with nothing able to clear it.
	 */
	function handleOpenChange(next: boolean) {
		if (!next) {
			onCancel?.();
		}
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};

		try {
			if (produk === undefined) {
				const parsed = CreateProdukInputSchema.safeParse({
					name,
					code,
					price,
					category,
					stock,
					// Status is a create-only decision, which is why the field is not
					// rendered when changing a Produk.
					active: statusOption.active
				});
				if (!parsed.success) {
					fieldErrors = collectFieldErrors(parsed.error, fields);
					return;
				}

				onSaved(await create.mutateAsync(parsed.data));
				return;
			}

			// The edit body carries no Stok and no Status: the schema has no field for
			// either, so there is nothing here to send that the API would have to drop.
			const parsed = UpdateProdukInputSchema.safeParse({ name, code, price, category });
			if (!parsed.success) {
				fieldErrors = collectFieldErrors(parsed.error, fields);
				return;
			}

			onSaved(await update.mutateAsync({ id: produk.id, input: parsed.data }));
		} catch {
			// `error` carries the normalized message, rendered below.
		}
	}

	/**
	 * Fields & Inputs (DESIGN.md): field 26px dengan label 12px/600 di atasnya,
	 * jaraknya 3px. Ukurannya literal, bukan token baru — itulah kosakata yang
	 * sudah dipakai layar Kasir, dan dua kosakata di tengah port lebih mahal
	 * daripada pengulangan (solution doc §1).
	 */
	const FIELD = 'flex flex-col gap-[3px]';
	const LABEL = 'text-xs leading-[1.2] font-semibold';
	const INPUT =
		'h-[26px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]';
	/**
	 * Field yang isinya angka: Harga dan Stok, bukan Nama, Kode, atau Kategori.
	 *
	 * Angkanya tabular, seperti setiap angka lain di dunia ini (DESIGN.md,
	 * Typography: "tanpa kecuali"), dan itu bukan hiasan: pada Inter Variable 13px
	 * sepuluh digit `1111111111` selebar 50px sementara `0000000000` selebar 80px,
	 * jadi nominal yang sedang diketik menggeser dirinya sendiri di dalam fieldnya
	 * tanpa kelas ini. Kasir memakai pembagian yang sama untuk field Jumlah bayar.
	 */
	const INPUT_ANGKA = `${INPUT} tabular-nums`;
	const SELECT = cn(
		INPUT,
		'w-full data-[size=default]:h-[26px] [&_svg:not([class*=size-])]:size-3'
	);
	/**
	 * Field yang tidak valid: pesannya 12px/600 di bawah field. Warnanya tinta, bukan
	 * merah utilitas — merah hanya milik garis dan outline field-nya (DESIGN.md,
	 * Fields & Inputs).
	 */
	const ERROR = 'text-xs font-semibold';
	/**
	 * `.btn` 26px/13px — petak berbingkai rambut, dan keadaan mati kehilangan
	 * kursor alih-alih tintanya (DESIGN.md, Commit Button).
	 *
	 * Dipakai dengan `variant="ghost"`: varian itu sudah membawa `hover:bg-muted`,
	 * persis latar hover dunia ini. Varian `default` membawa
	 * `text-primary-foreground`, jadi tombol berbingkai tanpa varian berakhir putih
	 * di atas putih (solution doc §3).
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
	/**
	 * Commit Button: satu-satunya bidang bertinta penuh di layar ini, dan saat mati
	 * ia kehilangan tintanya lalu jadi putih bergaris putus-putus — bukan pudar
	 * (DESIGN.md, State-Is-Not-Faded).
	 */
	const COMMIT =
		'h-10 w-full text-[15px] font-semibold hover:bg-primary disabled:pointer-events-auto disabled:cursor-not-allowed disabled:border-dashed disabled:border-border disabled:bg-card disabled:text-foreground disabled:opacity-100';
</script>

<!--
	Dialog-nya sendiri tidak membawa hiasan: yang menggambar adalah tiga pitanya —
	kepala Wash Grey bergaris tinta, badan, lalu kaki bergaris rambut tempat satu-
	satunya bidang bertinta penuh berdiri. Radius nol dan tanpa bayangan, seperti
	seluruh permukaan lain di dunia ini (DESIGN.md, No-Shadow & Square-Corner).
-->
<Dialog.Root open onOpenChange={handleOpenChange}>
	<Dialog.Content
		class="grid-cols-1 gap-0 rounded-none border border-border bg-card p-0 text-[13px] ring-0 ring-transparent"
		showCloseButton={false}
	>
		<Dialog.Header
			class="flex-row items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
		>
			<Dialog.Title class="text-[13px] leading-none font-bold tracking-[0.01em]">
				{produk ? `Ubah ${produk.name}` : 'Tambah Produk'}
			</Dialog.Title>
			{#if onCancel}
				<!--
					Jalan keluar berdiri sebagai kata di kepala dialog, bukan sebagai ikon
					silang: setiap kontrol di dunia ini membawa katanya sendiri
					(DESIGN.md, Named-Not-Hidden).
				-->
				<Button variant="ghost" class={AKSI_MODUL} onclick={onCancel}>Batal</Button>
			{/if}
		</Dialog.Header>

		<!--
			Deskripsi untuk pembaca layar saja: ia menyebut apa yang sedang diubah
			tanpa menambah satu baris pun ke papan.
		-->
		<Dialog.Description class="sr-only">
			{produk ? `Ubah data ${produk.name}.` : 'Tambahkan Produk baru ke katalog toko.'}
		</Dialog.Description>

		<form aria-label="Formulir Produk" onsubmit={submit} novalidate>
			<!--
				Kisi field menumpuk di 900px, sama seperti kisi field mana pun di dunia ini
				(ADR-0020) — dialog pun sebuah kisi field, bukan permukaan yang diatur papan.
			-->
			<div class="grid grid-cols-1 gap-x-3 gap-y-[10px] px-2 py-2 min-[901px]:grid-cols-2">
				<div class={FIELD}>
					<Label for="produk-nama" class={LABEL}>Nama</Label>
					<Input
						id="produk-nama"
						name="name"
						autocomplete="off"
						class={INPUT}
						bind:value={name}
						aria-invalid={fieldErrors.name ? true : undefined}
					/>
					{#if fieldErrors.name}
						<p class={ERROR}>{fieldErrors.name}</p>
					{/if}
				</div>

				<div class={FIELD}>
					<Label for="produk-kode" class={LABEL}>Kode</Label>
					<Input id="produk-kode" name="code" autocomplete="off" class={INPUT} bind:value={code} />
					<p class="text-xs">
						Boleh dikosongkan. Kode yang ada harus unik — barcode atau kode internal.
					</p>
				</div>

				<div class={FIELD}>
					<Label for="produk-harga" class={LABEL}>Harga</Label>
					<Input
						id="produk-harga"
						name="price"
						inputmode="numeric"
						autocomplete="off"
						class={INPUT_ANGKA}
						bind:value={price}
						aria-invalid={fieldErrors.price ? true : undefined}
					/>
					{#if fieldErrors.price}
						<p class={ERROR}>{fieldErrors.price}</p>
					{/if}
				</div>

				{#if !editing}
					<div class={FIELD}>
						<Label for="produk-stok" class={LABEL}>Stok</Label>
						<Input
							id="produk-stok"
							name="stock"
							inputmode="numeric"
							autocomplete="off"
							class={INPUT_ANGKA}
							bind:value={stock}
							aria-invalid={fieldErrors.stock ? true : undefined}
						/>
						{#if fieldErrors.stock}
							<p class={ERROR}>{fieldErrors.stock}</p>
						{/if}
					</div>
				{/if}

				<div class={FIELD}>
					<Label for="produk-kategori" class={LABEL}>Kategori</Label>
					<Input
						id="produk-kategori"
						name="category"
						autocomplete="off"
						class={INPUT}
						bind:value={category}
					/>
					<p class="text-xs">Opsional, satu tingkat. Hanya untuk pengelompokan.</p>
				</div>

				{#if !editing}
					<div class={FIELD}>
						<Label for="produk-status" class={LABEL}>Status</Label>
						<Select.Root type="single" value={status} onValueChange={setStatus}>
							<Select.Trigger id="produk-status" class={SELECT}>
								<!-- Written here rather than left to `Select.Value`: the label
								     registry is filled as items mount, and the content mounts
								     lazily, so the trigger would show the raw value first. -->
								<span data-slot="select-value">{statusOption.label}</span>
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
						<p class="text-xs">Produk Nonaktif tidak muncul di lookup kasir.</p>
					</div>
				{/if}
			</div>

			{#if error}
				<p class="border-t border-border px-2 py-1.5 text-xs font-semibold" role="alert">
					{error.message}
				</p>
			{/if}

			<div class="border-t border-border px-2 py-2">
				<Button type="submit" class={COMMIT} disabled={pending}>
					{pending ? 'Menyimpan…' : editing ? 'Simpan Perubahan' : 'Tambah'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
