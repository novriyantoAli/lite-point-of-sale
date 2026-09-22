<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { collectFieldErrors } from '$lib/utils';
	import { createAddStokMutation } from '../queries/produk.queries';
	import { TambahStokInputSchema, type Produk } from '../schemas/produk.schema';

	/**
	 * Records a restock for one Produk: how many units arrived. It asks for the
	 * amount received, never the new total — the Stok already on the Produk is the
	 * API's to add to, so two deliveries cannot overwrite each other.
	 *
	 * `produk` is the record being restocked, which the parent renders this
	 * alongside; the parent remounts the form per row, so the fields below start
	 * empty and there is no prop to keep in step.
	 */
	let {
		produk,
		onAdded,
		onCancel
	}: {
		produk: Produk;
		onAdded?: (produk: Produk) => void;
		onCancel?: () => void;
	} = $props();

	const add = createAddStokMutation();

	/**
	 * The field stays a string: a form submits text, and the schema is the one
	 * place that turns that text into the integer the API takes — the same rule
	 * the Go use case enforces, so a field error and a request error cannot
	 * disagree.
	 */
	let quantity = $state('');
	let fieldErrors = $state<Partial<Record<'quantity', string>>>({});

	const pending = $derived(add.isPending);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};

		const parsed = TambahStokInputSchema.safeParse({ quantity });
		if (!parsed.success) {
			fieldErrors = collectFieldErrors(parsed.error, ['quantity']);
			return;
		}

		try {
			const updated = await add.mutateAsync({ id: produk.id, quantity: parsed.data.quantity });
			quantity = '';
			onAdded?.(updated);
		} catch {
			// `add.error` carries the normalized message, rendered below.
		}
	}
</script>

<form
	class="flex flex-wrap items-end gap-2"
	aria-label={`Formulir Tambah Stok ${produk.name}`}
	onsubmit={submit}
	novalidate
>
	<div class="space-y-2">
		<Label for={`stok-${produk.id}`}>Jumlah masuk</Label>
		<Input
			id={`stok-${produk.id}`}
			name="quantity"
			inputmode="numeric"
			autocomplete="off"
			class="w-32"
			bind:value={quantity}
			aria-invalid={fieldErrors.quantity ? true : undefined}
		/>
		{#if fieldErrors.quantity}
			<p class="text-sm text-destructive">{fieldErrors.quantity}</p>
		{/if}
	</div>

	<Button type="submit" size="sm" disabled={pending}>
		{pending ? 'Menyimpan…' : 'Tambah Stok'}
	</Button>
	{#if onCancel}
		<Button type="button" variant="outline" size="sm" onclick={onCancel}>Batal</Button>
	{/if}
</form>

{#if add.error}
	<p class="text-sm text-destructive" role="alert">{add.error.message}</p>
{/if}
