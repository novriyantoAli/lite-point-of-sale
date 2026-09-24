<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { formatRupiah } from '$lib/utils';
	import { createCheckoutMutation } from '../queries/penjualan.queries';
	import {
		JumlahBayarSchema,
		METODE_LABEL,
		METODE_URUT,
		punyaKembalian,
		type HasilCheckout,
		type MetodePembayaran
	} from '../schemas/penjualan.schema';
	import { keranjangState } from '../state/keranjang.state.svelte';

	/**
	 * The Pembayaran of the checkout: which method the sale is paid by, and — for
	 * Tunai — the nominal the buyer hands over and the Kembalian that follows from
	 * it. The button turns the keranjang into a Penjualan (CONTEXT.md, Pembayaran,
	 * Tunai, Kembalian).
	 *
	 * QRIS, Debit and Transfer are only recorded: no gateway is called, and the
	 * nominal is the total of the sale rather than something the Kasir types
	 * (CONTEXT.md, Pembayaran). That is why the amount field and the Kembalian
	 * belong to Tunai alone.
	 */
	let { onCheckedOut }: { onCheckedOut: (hasil: HasilCheckout) => void } = $props();

	const checkout = createCheckoutMutation();

	/**
	 * The method the Kasir picked. Tunai is the default: it is the one that needs
	 * an amount typed, and the one a till takes most of the day.
	 */
	let method = $state<MetodePembayaran>('cash');

	/**
	 * The amount stays a string: a form submits text, and `JumlahBayarSchema` is
	 * the one place that turns that text into the integer the API takes. Parsing
	 * here as well would be a second set of rules to keep in step.
	 */
	let amount = $state('');
	let amountError = $state('');

	const total = $derived(keranjangState.total);

	/** Whether the method being paid is Tunai, the only one with a Kembalian. */
	const tunai = $derived(punyaKembalian(method));

	/**
	 * What the Kasir will hand back, worked out as the amount is typed — and `null`
	 * while the amount is not yet a number that covers the total, so the preview
	 * never shows a Kembalian the checkout would refuse. A recorded method has no
	 * Kembalian at all, so it stays `null` there.
	 */
	const parsedAmount = $derived(JumlahBayarSchema.safeParse(amount));
	const change = $derived(
		tunai && parsedAmount.success && parsedAmount.data >= total ? parsedAmount.data - total : null
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

	/**
	 * The nominal this Pembayaran is recorded for. For Tunai it is what the Kasir
	 * typed, and it has to cover the total — the rule of CONTEXT.md, Kembalian: the
	 * Kembalian is `bayar − total` and must be at least zero. For a recorded method
	 * there is nothing to type and nothing to hand back, so it is the total, which
	 * is also what the API requires.
	 *
	 * Answers the empty string when Tunai's field does not hold a usable amount,
	 * after writing the message to show.
	 */
	function nominal(): number | '' {
		if (!tunai) {
			return total;
		}

		const parsed = JumlahBayarSchema.safeParse(amount);
		if (!parsed.success) {
			amountError = parsed.error.issues[0]?.message ?? 'Jumlah bayar tidak valid.';
			return '';
		}
		if (parsed.data < total) {
			amountError = 'Jumlah bayar kurang dari total.';
			return '';
		}

		return parsed.data;
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		amountError = '';

		if (blocked) {
			amountError = blocked;
			return;
		}

		const paid = nominal();
		if (paid === '') {
			return;
		}

		try {
			const hasil = await checkout.mutateAsync({
				items: keranjangState.items.map((item) => ({
					product_id: item.produk.id,
					quantity: item.qty
				})),
				payment: { method, amount: paid }
			});
			onCheckedOut(hasil);
		} catch {
			// `checkout.error` carries the normalized message, rendered below.
		}
	}
</script>

<form class="space-y-4 rounded-lg border p-4" aria-label="Pembayaran" onsubmit={submit} novalidate>
	<h2 class="font-medium">Pembayaran</h2>

	<p class="text-sm text-muted-foreground">
		Total yang harus dibayar
		<span class="font-medium text-foreground tabular-nums">{formatRupiah(total)}</span>
	</p>

	<fieldset class="space-y-2">
		<legend class="text-sm font-medium">Metode pembayaran</legend>
		<div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
			{#each METODE_URUT as pilihan (pilihan)}
				<!--
					A native radio, so arrow keys move between the methods and the group
					is announced as one choice. The input is only hidden from sight: the
					label is what shows the method and what carries the focus ring.
				-->
				<label
					class="flex cursor-pointer items-center justify-center rounded-md border px-3 py-2 text-sm font-medium transition-colors focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2 hover:bg-accent has-[:checked]:border-primary has-[:checked]:bg-primary/10"
				>
					<input type="radio" name="method" value={pilihan} bind:group={method} class="sr-only" />
					{METODE_LABEL[pilihan]}
				</label>
			{/each}
		</div>
	</fieldset>

	{#if tunai}
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
	{:else}
		<!--
			Nothing to type and nothing to hand back: the recorded method pays the
			total, and the API records exactly that.
		-->
		<div class="space-y-2">
			<span class="text-sm font-medium">Dibayar</span>
			<p
				class="text-2xl font-semibold tabular-nums"
				role="status"
				aria-label={`Dibayar dengan ${METODE_LABEL[method]} ${formatRupiah(total)}`}
			>
				{METODE_LABEL[method]} · {formatRupiah(total)}
			</p>
		</div>
	{/if}

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
