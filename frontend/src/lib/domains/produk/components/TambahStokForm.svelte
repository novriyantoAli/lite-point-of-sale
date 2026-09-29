<script lang="ts">
	import { COMMIT_BARIS, ERROR, FIELD, INPUT_ANGKA, LABEL } from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { cn, collectFieldErrors } from '$lib/utils';
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

	/** Field jumlah: field angka dunia ini yang menyempit ke `w-32`. */
	const INPUT_JUMLAH = cn(INPUT_ANGKA, 'w-32');
	/**
	 * Tombol berbingkai memakai `variant="ghost"` lebih dulu: varian itu sudah
	 * membawa `hover:bg-muted`, persis latar hover dunia ini, dan tidak membawa
	 * apa pun yang harus dilawan (solution doc §3).
	 */
	const BATAL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground';
</script>

<form
	class="flex flex-wrap items-end gap-1.5"
	aria-label={`Formulir Tambah Stok ${produk.name}`}
	onsubmit={submit}
	novalidate
>
	<div class={FIELD}>
		<Label for={`stok-${produk.id}`} class={LABEL}>Jumlah masuk</Label>
		<Input
			id={`stok-${produk.id}`}
			name="quantity"
			inputmode="numeric"
			autocomplete="off"
			class={INPUT_JUMLAH}
			bind:value={quantity}
			aria-invalid={fieldErrors.quantity ? true : undefined}
		/>
		{#if fieldErrors.quantity}
			<p class={ERROR}>{fieldErrors.quantity}</p>
		{/if}
	</div>

	<Button type="submit" class={COMMIT_BARIS} disabled={pending}>
		{pending ? 'Menyimpan…' : 'Tambah Stok'}
	</Button>
	{#if onCancel}
		<Button type="button" variant="ghost" class={BATAL} onclick={onCancel}>Batal</Button>
	{/if}
</form>

{#if add.error}
	<p class="text-xs font-semibold" role="alert">{add.error.message}</p>
{/if}
