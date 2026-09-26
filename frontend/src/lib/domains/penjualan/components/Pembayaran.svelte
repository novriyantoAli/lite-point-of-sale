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
	 *
	 * Total yang harus dibayar tidak diulang di sini: pitanya sudah berdiri tepat
	 * di atas form ini, di kolom yang sama (DESIGN.md, One-Screen).
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

	/** Satu metode: radio asli yang disembunyikan, sel persegi yang terlihat. */
	const METODE =
		'flex h-[26px] cursor-pointer items-center justify-center border border-border bg-card text-[13px] font-medium transition-colors hover:bg-muted focus-within:outline-2 focus-within:outline-solid focus-within:-outline-offset-2 focus-within:outline-foreground has-[:checked]:border-foreground has-[:checked]:bg-foreground has-[:checked]:font-semibold has-[:checked]:text-primary-foreground';
	const FIELD =
		'h-[26px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]';
	/** Angka yang jadi jawaban layar: 20px/700, selalu tabular. */
	const FIGURE = 'text-[20px] leading-[1.1] font-bold tabular-nums';
</script>

<!--
	Formnya sendiri tidak menggambar garis apa pun: ia deretan modul, dan tiap
	modul membawa garis bawahnya sendiri, jadi dua modul bersebelahan berbagi satu
	garis rambut alih-alih menggambar dua (DESIGN.md, Shared-Hairline).
-->
<form aria-label="Pembayaran" onsubmit={submit} novalidate>
	<div class="border-b border-border px-2 py-1.5">
		<p class="text-xs font-semibold">Metode pembayaran</p>
		<div class="mt-1 grid grid-cols-2 gap-1">
			{#each METODE_URUT as pilihan (pilihan)}
				<!--
					A native radio, so arrow keys move between the methods and the group
					is announced as one choice. The input is only hidden from sight: the
					label is what shows the method and what carries the focus ring.
				-->
				<label class={METODE}>
					<input type="radio" name="method" value={pilihan} bind:group={method} class="sr-only" />
					{METODE_LABEL[pilihan]}
				</label>
			{/each}
		</div>
	</div>

	{#if tunai}
		<div class="border-b border-border px-2 py-1.5">
			<div class="flex flex-col gap-[3px]">
				<Label for="kasir-bayar" class="text-xs font-semibold">Jumlah bayar</Label>
				<Input
					id="kasir-bayar"
					name="amount"
					inputmode="numeric"
					autocomplete="off"
					class={FIELD}
					bind:value={amount}
					aria-invalid={amountError ? true : undefined}
				/>
			</div>
			{#if amountError}
				<p class="mt-1 text-xs font-semibold">{amountError}</p>
			{/if}
		</div>

		<div class="border-b border-border px-2 py-1.5">
			<p class="text-xs font-semibold">Kembalian</p>
			<p
				class={FIGURE}
				role="status"
				aria-label={`Kembalian ${change === null ? 'belum bisa dihitung' : formatRupiah(change)}`}
			>
				{change === null ? '—' : formatRupiah(change)}
			</p>
		</div>
	{:else}
		<!--
			Nothing to type and nothing to hand back: the recorded method pays the
			total, and the API records exactly that.
		-->
		<div class="border-b border-border px-2 py-1.5">
			<p class="text-xs font-semibold">Dibayar</p>
			<p
				class={FIGURE}
				role="status"
				aria-label={`Dibayar dengan ${METODE_LABEL[method]} ${formatRupiah(total)}`}
			>
				{METODE_LABEL[method]} · {formatRupiah(total)}
			</p>
		</div>
	{/if}

	<div class="border-b border-border px-2 py-1.5">
		{#if blocked}
			<p class="mb-1.5 text-xs font-semibold">{blocked}</p>
		{/if}

		{#if checkout.error}
			<p class="mb-1.5 text-xs font-semibold" role="alert">{checkout.error.message}</p>
		{/if}

		<!--
			Aksi utama: satu-satunya bidang bertinta penuh di papan ini. Saat mati ia
			kehilangan tintanya (bukan dipudarkan), jadi putih bergaris putus-putus.
		-->
		<Button
			type="submit"
			class="h-10 w-full text-[15px] font-semibold hover:bg-primary disabled:border-dashed disabled:border-border disabled:bg-card disabled:text-foreground disabled:opacity-100"
			disabled={pending || Boolean(blocked)}
		>
			{pending ? 'Menyimpan…' : 'Bayar & Simpan Penjualan'}
		</Button>
	</div>
</form>
