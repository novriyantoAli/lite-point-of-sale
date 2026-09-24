<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { formatRupiah } from '$lib/utils';
	import { METODE_LABEL, punyaKembalian, type Penjualan } from '../schemas/penjualan.schema';

	/**
	 * What the Kasir sees once a keranjang has become a Penjualan: the Nomor Struk
	 * it was recorded under, what was sold at the price it was sold for, and — for
	 * Tunai — the Kembalian that followed from the payment.
	 *
	 * The Penjualan shown is the one the API stored — the total and the Kembalian
	 * are its answers, not this screen's arithmetic. Printing the Struk is #8; this
	 * is the record of the sale the till just wrote.
	 */
	let { sale, onBaru }: { sale: Penjualan; onBaru: () => void } = $props();
</script>

<section class="space-y-4 rounded-lg border p-4" aria-labelledby="kasir-struk-judul">
	<div class="space-y-1">
		<h2 id="kasir-struk-judul" class="font-medium">Penjualan tercatat</h2>
		<p class="text-sm text-muted-foreground" role="status">
			Nomor Struk <span class="font-medium text-foreground tabular-nums">{sale.receipt_number}</span
			>
			· {sale.created_at} · Kasir {sale.cashier_name}
		</p>
	</div>

	<ul class="space-y-1 text-sm">
		{#each sale.items as item (item.product_id)}
			<li class="flex flex-wrap justify-between gap-3">
				<span>
					{item.name}
					<span class="text-muted-foreground">
						· {item.quantity} × {formatRupiah(item.price)}
					</span>
				</span>
				<span class="tabular-nums">{formatRupiah(item.subtotal)}</span>
			</li>
		{/each}
	</ul>

	<div class="space-y-1 border-t pt-3 text-sm">
		<p class="flex justify-between gap-3 font-medium">
			<span>Total</span>
			<span class="tabular-nums">{formatRupiah(sale.total)}</span>
		</p>
		<p class="flex justify-between gap-3">
			<span>Bayar · {METODE_LABEL[sale.payment.method]}</span>
			<span class="tabular-nums">{formatRupiah(sale.payment.amount)}</span>
		</p>
		{#if punyaKembalian(sale.payment.method)}
			<!-- Only Tunai has a Kembalian; a recorded method pays the total exactly
			     (CONTEXT.md, Kembalian). -->
			<p class="flex justify-between gap-3 font-medium">
				<span>Kembalian</span>
				<span class="tabular-nums">{formatRupiah(sale.payment.change)}</span>
			</p>
		{/if}
	</div>

	<Button onclick={onBaru}>Penjualan Baru</Button>
</section>
