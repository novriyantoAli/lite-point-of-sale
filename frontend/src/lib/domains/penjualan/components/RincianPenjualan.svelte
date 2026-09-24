<script lang="ts">
	import { formatRupiah } from '$lib/utils';
	import { METODE_LABEL, punyaKembalian, type Penjualan } from '../schemas/penjualan.schema';

	/**
	 * The record of one Penjualan: its Nomor Struk, when and by whom it was rung
	 * up, what was sold at the price it was sold for, and how it was paid.
	 *
	 * It is presentational — it decides nothing and owns no state. Both the till's
	 * post-checkout panel and the `/penjualan` lookup show exactly this, and a
	 * second copy of the markup would be a second place for the two to drift. The
	 * numbers it prints are the API's answers, not its own arithmetic.
	 */
	let { sale }: { sale: Penjualan } = $props();
</script>

<div class="space-y-4">
	<p class="text-sm text-muted-foreground" role="status">
		Nomor Struk <span class="font-medium text-foreground tabular-nums">{sale.receipt_number}</span>
		· {sale.created_at} · Kasir {sale.cashier_name}
	</p>

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
</div>
