<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { formatRupiah } from '$lib/utils';
	import { keranjangState } from '../state/keranjang.state.svelte';

	/**
	 * The keranjang: what is being sold right now, with a running total (CONTEXT.md,
	 * Item). It reads the domain's own runes state, so the lookup above and the
	 * payment form below both work on the same draft without props threaded between
	 * them.
	 */

	/**
	 * Sets a line's quantity from the field. A whole number of at least one is
	 * stored as typed; anything else is not — a line of zero is not an Item — and
	 * the field is put back to the quantity the keranjang holds, so it never shows a
	 * number the subtotal beside it disagrees with.
	 *
	 * An emptied field is left empty on purpose: the Kasir clearing it to type a new
	 * quantity must not have the old one typed back in under their fingers. The
	 * running total keeps counting the quantity that is still stored until the new
	 * one is typed.
	 */
	function setQuantity(event: Event, productId: number) {
		const field = event.currentTarget as HTMLInputElement;
		const typed = field.valueAsNumber;

		if (Number.isInteger(typed) && typed >= 1) {
			keranjangState.setQuantity(productId, typed);
			return;
		}

		if (field.value !== '') {
			field.value = String(keranjangState.setQuantity(productId, typed));
		}
	}
</script>

<section class="space-y-4" aria-labelledby="kasir-keranjang-judul">
	<div class="flex flex-wrap items-end justify-between gap-4">
		<div class="space-y-1">
			<h2 id="kasir-keranjang-judul" class="font-medium">Keranjang</h2>
			<p class="text-sm text-muted-foreground">
				{keranjangState.units === 0
					? 'Belum ada Item.'
					: `${keranjangState.units} unit dalam ${keranjangState.items.length} Item.`}
			</p>
		</div>

		<Button
			variant="outline"
			size="sm"
			disabled={keranjangState.items.length === 0}
			onclick={() => keranjangState.clear()}
		>
			Kosongkan
		</Button>
	</div>

	{#if keranjangState.items.length === 0}
		<p class="text-sm text-muted-foreground">
			Keranjang kosong. Cari Produk di atas, lalu tekan Tambah.
		</p>
	{:else}
		<ul class="space-y-2">
			{#each keranjangState.items as item (item.produk.id)}
				{@const terlaluBanyak = item.qty > item.produk.stock}
				<li class="space-y-2 rounded-lg border p-3">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="space-y-1">
							<p class="font-medium">
								{item.produk.name}
								{#if item.produk.code}
									<span class="text-muted-foreground">· {item.produk.code}</span>
								{/if}
							</p>
							<p class="text-sm text-muted-foreground">
								{formatRupiah(item.produk.price)} · Stok {item.produk.stock}
							</p>
						</div>

						<div class="flex items-center gap-2">
							<Button
								variant="outline"
								size="sm"
								disabled={item.qty <= 1}
								aria-label={`Kurangi jumlah ${item.produk.name}`}
								onclick={() => keranjangState.setQuantity(item.produk.id, item.qty - 1)}
							>
								−
							</Button>
							<Input
								class="w-20 text-center"
								type="number"
								min="1"
								step="1"
								inputmode="numeric"
								value={item.qty}
								aria-label={`Jumlah ${item.produk.name}`}
								aria-invalid={terlaluBanyak ? true : undefined}
								oninput={(event) => setQuantity(event, item.produk.id)}
							/>
							<Button
								variant="outline"
								size="sm"
								aria-label={`Tambah jumlah ${item.produk.name}`}
								onclick={() => keranjangState.setQuantity(item.produk.id, item.qty + 1)}
							>
								+
							</Button>
						</div>

						<div class="flex items-center gap-3">
							<span class="tabular-nums">{formatRupiah(item.produk.price * item.qty)}</span>
							<Button
								variant="outline"
								size="sm"
								aria-label={`Hapus ${item.produk.name} dari keranjang`}
								onclick={() => keranjangState.remove(item.produk.id)}
							>
								Hapus
							</Button>
						</div>
					</div>

					{#if terlaluBanyak}
						<p class="text-sm text-destructive">
							Melebihi Stok: tersisa {item.produk.stock}. Kurangi jumlahnya sebelum checkout.
						</p>
					{/if}
				</li>
			{/each}
		</ul>

		<p
			class="text-right text-lg font-semibold"
			role="status"
			aria-label={`Total keranjang ${formatRupiah(keranjangState.total)}`}
		>
			Total <span class="tabular-nums">{formatRupiah(keranjangState.total)}</span>
		</p>
	{/if}
</section>
