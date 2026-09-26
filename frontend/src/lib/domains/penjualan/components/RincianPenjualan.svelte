<script lang="ts">
	import { cn, formatRupiah } from '$lib/utils';
	import { METODE_LABEL, punyaKembalian, type Penjualan } from '../schemas/penjualan.schema';

	/**
	 * The record of one Penjualan: its Nomor Struk, when and by whom it was rung
	 * up, what was sold at the price it was sold for, and how it was paid.
	 *
	 * It is presentational — it decides nothing and owns no state. Both the till's
	 * post-checkout panel and the `/penjualan` lookup show exactly this, and a
	 * second copy of the markup would be a second place for the two to drift. The
	 * numbers it prints are the API's answers, not its own arithmetic.
	 *
	 * Tiap bagian adalah satu modul yang membawa garis bawahnya sendiri; yang
	 * memanggilnya menaruhnya di dalam kolom, dan kolom itu yang menentukan
	 * lebarnya (DESIGN.md, Modules).
	 */
	let { sale }: { sale: Penjualan } = $props();
</script>

<div>
	<p class="border-b border-border px-2 py-1.5 text-xs tabular-nums" role="status">
		Nomor Struk {sale.receipt_number} · {sale.created_at} · Kasir {sale.cashier_name}
	</p>

	<ul>
		{#each sale.items as item (item.product_id)}
			<li class="grid grid-cols-[1fr_auto] gap-x-2 border-b border-border px-2 py-1.5">
				<span class="text-[13px] font-medium">{item.name}</span>
				<span class="text-right text-[13px] font-semibold tabular-nums">
					{formatRupiah(item.subtotal)}
				</span>
				<span class="col-span-2 text-xs tabular-nums">
					· {item.quantity} × {formatRupiah(item.price)}
				</span>
			</li>
		{/each}
	</ul>

	<p
		class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
	>
		<span class="text-[13px] font-bold">Total</span>
		<span class="text-[15px] font-bold tabular-nums">{formatRupiah(sale.total)}</span>
	</p>

	<!--
		Bayar dan Kembalian berdiri sebagai dua sel bersebelahan: keduanya angka yang
		menjawab pertanyaan yang sama — berapa yang diterima, berapa yang kembali. Label
		dan angkanya satu paragraf, karena pembacanya membaca keduanya sebagai satu
		jawaban (DESIGN.md, Figures).
	-->
	<div
		class={cn(
			'grid border-b border-border',
			punyaKembalian(sale.payment.method) ? 'grid-cols-2' : 'grid-cols-1'
		)}
	>
		<p class="border-border px-2 py-1.5 text-xs font-semibold">
			Bayar · {METODE_LABEL[sale.payment.method]}
			<span class="block text-[20px] leading-[1.1] font-bold tabular-nums">
				{formatRupiah(sale.payment.amount)}
			</span>
		</p>

		{#if punyaKembalian(sale.payment.method)}
			<!-- Only Tunai has a Kembalian; a recorded method pays the total exactly
			     (CONTEXT.md, Kembalian). -->
			<p class="border-l border-border px-2 py-1.5 text-xs font-semibold">
				Kembalian
				<span class="block text-[20px] leading-[1.1] font-bold tabular-nums">
					{formatRupiah(sale.payment.change)}
				</span>
			</p>
		{/if}
	</div>
</div>
