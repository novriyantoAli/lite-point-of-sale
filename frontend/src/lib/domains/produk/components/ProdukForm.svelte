<script lang="ts">
	import { untrack } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { collectFieldErrors } from '$lib/utils';
	import { createProdukMutation, createUpdateProdukMutation } from '../queries/produk.queries';
	import { ProdukInputSchema, type Produk } from '../schemas/produk.schema';

	/**
	 * Adds a Produk or changes one. The same fields serve both, because an edit
	 * replaces the whole editable record — there is no second shape for it.
	 *
	 * `produk` is the record being changed, or undefined to add a new one. The
	 * parent remounts this component per record (`{#key}`), and that is what gives
	 * the fields below their starting values: initializing from a prop that later
	 * changes would need an `$effect` writing state, which is how a form ends up
	 * fighting the person typing into it.
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
	 * The fields stay strings: a form submits text, and the schema is the one
	 * place that turns that text into the integers the API takes (§11, and the
	 * reason `ProdukInputSchema` owns the parsing). Parsing here as well would be a
	 * second set of rules to keep in step.
	 */
	// The parent remounts this component per record, so the prop is deliberately
	// read only once, when the fields are initialized: `untrack` states that
	// instead of leaving the next reader to wonder whether it was an oversight.
	let name = $state(untrack(() => produk?.name ?? ''));
	let code = $state(untrack(() => produk?.code ?? ''));
	let price = $state(untrack(() => (produk === undefined ? '' : String(produk.price))));
	let category = $state(untrack(() => produk?.category ?? ''));
	let stock = $state(untrack(() => (produk === undefined ? '' : String(produk.stock))));
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

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};

		const parsed = ProdukInputSchema.safeParse({
			name,
			code,
			price,
			category,
			stock,
			// Status is a create-only decision, so an edit leaves the field out
			// entirely rather than sending a value the API would ignore.
			...(editing ? {} : { active: statusOption.active })
		});
		if (!parsed.success) {
			fieldErrors = collectFieldErrors(parsed.error, fields);
			return;
		}

		try {
			const saved =
				produk === undefined
					? await create.mutateAsync(parsed.data)
					: await update.mutateAsync({ id: produk.id, input: parsed.data });
			onSaved(saved);
		} catch {
			// `error` carries the normalized message, rendered below.
		}
	}
</script>

<form
	class="space-y-4 rounded-lg border p-4"
	aria-label="Formulir Produk"
	onsubmit={submit}
	novalidate
>
	<h2 class="font-medium">{produk ? `Ubah ${produk.name}` : 'Tambah Produk'}</h2>

	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<Label for="produk-nama">Nama</Label>
			<Input
				id="produk-nama"
				name="name"
				autocomplete="off"
				bind:value={name}
				aria-invalid={fieldErrors.name ? true : undefined}
			/>
			{#if fieldErrors.name}
				<p class="text-sm text-destructive">{fieldErrors.name}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label for="produk-kode">Kode</Label>
			<Input id="produk-kode" name="code" autocomplete="off" bind:value={code} />
			<p class="text-sm text-muted-foreground">
				Boleh dikosongkan. Kode yang ada harus unik — barcode atau kode internal.
			</p>
		</div>

		<div class="space-y-2">
			<Label for="produk-harga">Harga</Label>
			<Input
				id="produk-harga"
				name="price"
				inputmode="numeric"
				autocomplete="off"
				bind:value={price}
				aria-invalid={fieldErrors.price ? true : undefined}
			/>
			{#if fieldErrors.price}
				<p class="text-sm text-destructive">{fieldErrors.price}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label for="produk-stok">Stok</Label>
			<Input
				id="produk-stok"
				name="stock"
				inputmode="numeric"
				autocomplete="off"
				bind:value={stock}
				aria-invalid={fieldErrors.stock ? true : undefined}
			/>
			{#if fieldErrors.stock}
				<p class="text-sm text-destructive">{fieldErrors.stock}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label for="produk-kategori">Kategori</Label>
			<Input id="produk-kategori" name="category" autocomplete="off" bind:value={category} />
			<p class="text-sm text-muted-foreground">
				Opsional, satu tingkat. Hanya untuk pengelompokan.
			</p>
		</div>

		{#if !editing}
			<div class="space-y-2">
				<Label for="produk-status">Status</Label>
				<Select.Root type="single" value={status} onValueChange={setStatus}>
					<Select.Trigger id="produk-status" class="w-full">
						<!-- Written here rather than left to `Select.Value`: the label
						     registry is filled as items mount, and the content mounts
						     lazily, so the trigger would show the raw value first. -->
						<span data-slot="select-value">{statusOption.label}</span>
					</Select.Trigger>
					<Select.Content>
						{#each STATUS_OPTIONS as option (option.value)}
							<Select.Item value={option.value} label={option.label}>
								{option.label}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
				<p class="text-sm text-muted-foreground">Produk Nonaktif tidak muncul di lookup kasir.</p>
			</div>
		{/if}
	</div>

	{#if error}
		<p class="text-sm text-destructive" role="alert">{error.message}</p>
	{/if}

	<div class="flex gap-2">
		<Button type="submit" disabled={pending}>
			{pending ? 'Menyimpan…' : editing ? 'Simpan Perubahan' : 'Tambah'}
		</Button>
		{#if onCancel}
			<Button type="button" variant="outline" onclick={onCancel}>Batal</Button>
		{/if}
	</div>
</form>
