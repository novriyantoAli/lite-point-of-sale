<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { formatRupiah } from '$lib/utils';
	import { createCheckoutMutation } from '../queries/penjualan.queries';
	import { JumlahBayarSchema, type Penjualan } from '../schemas/penjualan.schema';
	import { keranjangState } from '../state/keranjang.state.svelte';

	/**
	 * The Pembayaran Tunai of the checkout: the nominal the buyer hands over, the
	 * Kembalian that follows from it, and the button that turns the keranjang into
	 * a Penjualan (CONTEXT.md, Tunai, Kembalian).
	 *
	 * Non-tunai methods are #7's; the API refuses them until then, so this form
	 * offers Tunai only rather than a dropdown that would fail on three of its four
	 * options.
	 */
	let { onCheckedOut }: { onCheckedOut: (sale: Penjualan) => void } = $props();

	const checkout = createCheckoutMutation();

	/**
	 * The amount stays a string: a form submits text, and `JumlahBayarSchema` is
	 * the one place that turns that text into the integer the API takes. Parsing
	 * here as well would be a second set of rules to keep in step.
	 */
	let amount = $state('');
	let amountError = $state('');

	const total = $derived(keranjangState.total);

	/**
	 * What the Kasir will hand back, worked out as the amount is typed — and `null`
	 * while the amount is not yet a number that covers the total, so the preview
	 * never shows a Kembalian the checkout would refuse.
	 */
	const parsedAmount = $derived(JumlahBayarSchema.safeParse(amount));
	const change = $derived(
		parsedAmount.success && parsedAmount.data >= total ? parsedAmount.data - total : null
	);

	/**
	 * Why the checkout cannot go ahead yet, or the empty string when it can. The
	 * keranjang is what knows: an empty one has nothing to sell, and a line above
	 * its Stok is the refusal of CONTEXT.md, Stok.
	 */
	const blocked = $derived(
		keranjangState.items.length === 0
			? 'Keranjang masih kosong.'
			: keranjangState.exceedsStock
				? 'Ada Item yang melebihi Stok. Kurangi jumlahnya lebih dulu.'
				: ''
	);

	const pending = $derived(checkout.isPending);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		amountError = '';

		if (blocked) {
			amountError = blocked;
			return;
		}

		const parsed = JumlahBayarSchema.safeParse(amount);
		if (!parsed.success) {
			amountError = parsed.error.issues[0]?.message ?? 'Jumlah bayar tidak valid.';
			return;
		}
		// The rule of CONTEXT.md, Kembalian: bayar − total, and it has to be at
		// least zero. The API refuses an underpayment too; this is so the Kasir
		// reads it before the request is sent.
		if (parsed.data < total) {
			amountError = 'Jumlah bayar kurang dari total.';
			return;
		}

		try {
			const sale = await checkout.mutateAsync({
				items: keranjangState.items.map((item) => ({
					product_id: item.produk.id,
					quantity: item.qty
				})),
				payment: { method: 'cash', amount: parsed.data }
			});
			onCheckedOut(sale);
		} catch {
			// `checkout.error` carries the normalized message, rendered below.
		}
	}
</script>

<form
	class="space-y-4 rounded-lg border p-4"
	aria-label="Pembayaran Tunai"
	onsubmit={submit}
	novalidate
>
	<h2 class="font-medium">Pembayaran Tunai</h2>

	<p class="text-sm text-muted-foreground">
		Total yang harus dibayar
		<span class="font-medium text-foreground tabular-nums">{formatRupiah(total)}</span>
	</p>

	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<Label for="kasir-bayar">Jumlah bayar</Label>
			<Input
				id="kasir-bayar"
				name="amount"
				inputmode="numeric"
				autocomplete="off"
				bind:value={amount}
				aria-invalid={amountError ? true : undefined}
			/>
			{#if amountError}
				<p class="text-sm text-destructive">{amountError}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<span class="text-sm font-medium">Kembalian</span>
			<p
				class="text-2xl font-semibold tabular-nums"
				role="status"
				aria-label={`Kembalian ${change === null ? 'belum bisa dihitung' : formatRupiah(change)}`}
			>
				{change === null ? '—' : formatRupiah(change)}
			</p>
		</div>
	</div>

	{#if blocked}
		<p class="text-sm text-muted-foreground">{blocked}</p>
	{/if}

	{#if checkout.error}
		<p class="text-sm text-destructive" role="alert">{checkout.error.message}</p>
	{/if}

	<Button type="submit" disabled={pending || Boolean(blocked)}>
		{pending ? 'Menyimpan…' : 'Bayar & Simpan Penjualan'}
	</Button>
</form>
