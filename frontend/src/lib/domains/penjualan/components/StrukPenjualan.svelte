<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import type { HasilCetak, Penjualan } from '../schemas/penjualan.schema';
	import CetakStruk from './CetakStruk.svelte';
	import RincianPenjualan from './RincianPenjualan.svelte';

	/**
	 * What the Kasir sees once a keranjang has become a Penjualan: the record of
	 * the sale the till just wrote, the outcome of the Struk the API printed for
	 * it, and the way back to a new one.
	 *
	 * A print that failed is shown here with a button that prints again — the
	 * Penjualan is stored either way, and the retry goes through the same endpoint
	 * as the automatic print (ADR-0017, keputusan 1 and 5).
	 *
	 * Layarnya menggantikan kasir, bukan menemani mosaiknya: aksi katalog yang tetap
	 * hidup di sebelah struk akan menambah keranjang yang tidak terlihat — aksi
	 * tersembunyi, dan itu yang dunia ini tolak (DESIGN.md, Named-Not-Hidden).
	 */
	let { sale, cetak, onBaru }: { sale: Penjualan; cetak: HasilCetak; onBaru: () => void } =
		$props();
</script>

<section
	class="grid grid-cols-1 border-x border-b border-border bg-card min-[1081px]:grid-cols-[minmax(0,1fr)_300px]"
	aria-labelledby="kasir-struk-judul"
>
	<div class="flex flex-col border-border min-[1081px]:border-r">
		<div
			class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
		>
			<h2 id="kasir-struk-judul" class="text-[13px] font-bold tracking-[0.01em]">
				Penjualan tercatat
			</h2>
			<!--
				Penjualan bersifat final — tidak ada void dan tidak ada refund — dan tab
				merah inilah yang mengatakannya, bukan kalimat penjelas (CONTEXT.md).
			-->
			<span
				class="inline-flex h-[15px] items-center bg-destructive px-[5px] text-xs font-semibold text-primary-foreground"
			>
				tersegel
			</span>
		</div>

		<RincianPenjualan {sale} />
	</div>

	<div class="flex flex-col">
		<CetakStruk nomorStruk={sale.receipt_number} hasilAwal={cetak} />

		<!-- Satu-satunya bidang bertinta penuh di layar ini: jalan ke pembeli berikutnya. -->
		<div class="border-b border-border px-2 py-1.5">
			<Button class="h-10 w-full text-[15px] font-semibold hover:bg-primary" onclick={onBaru}>
				Penjualan Baru
			</Button>
		</div>
	</div>
</section>
