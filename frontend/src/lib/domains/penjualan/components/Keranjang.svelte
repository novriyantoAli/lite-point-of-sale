<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { formatRupiah } from '$lib/utils';
	import { keranjangState } from '../state/keranjang.state.svelte';

	/**
	 * Keranjang: apa yang sedang dijual sekarang, dengan total berjalan
	 * (CONTEXT.md, Item). Ia membaca keadaan runes milik domainnya sendiri, jadi
	 * mosaik katalog di kolom tengah dan form Pembayaran di bawahnya bekerja atas
	 * rancangan yang sama tanpa props yang diuntai di antaranya.
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

	/**
	 * Satu tombol operator. Angka dan tombolnya berbagi satu bingkai rambut, jadi
	 * tidak ada dua garis yang menempel — dan petaknya yang mati kehilangan
	 * tintanya, bukan opasitasnya (DESIGN.md, State-Is-Not-Faded).
	 */
	const OPERATOR =
		'grid h-[22px] w-[22px] place-items-center border-border text-[13px] leading-none enabled:hover:bg-muted disabled:cursor-not-allowed disabled:text-border';
	/**
	 * Field jumlah tinggal di dalam bingkai petaknya: tanpa garis sendiri, tanpa
	 * cincin cahaya, dan angka putar bawaan browser disembunyikan — ia bagian dari
	 * peramban yang tidak digambar dunia ini.
	/**
	 * Field jumlah tinggal di dalam bingkai petaknya: tanpa garis sendiri dan tanpa
	 * cincin cahaya — fokusnya digambar oleh aturan dunia di `app.css`, sekali untuk
	 * seluruh aplikasi. Angka putar bawaan browser disembunyikan: ia bagian dari
	 * peramban yang tidak digambar dunia ini.
	 */
	const JUMLAH =
		'h-[22px] w-10 border-0 bg-card px-0.5 text-center text-[13px] tabular-nums shadow-none aria-invalid:ring-0 md:text-[13px] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none';
</script>

<section aria-labelledby="kasir-keranjang-judul">
	<div
		class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
	>
		<h2 id="kasir-keranjang-judul" class="text-[13px] font-bold tracking-[0.01em]">Keranjang</h2>
		<span class="text-xs">
			{keranjangState.units === 0
				? 'Belum ada Item.'
				: `${keranjangState.units} unit dalam ${keranjangState.items.length} Item.`}
		</span>
	</div>

	{#if keranjangState.items.length === 0}
		<p class="border-b border-border px-2 py-1.5 text-xs">
			Keranjang kosong. Tekan satu petak di katalog untuk menambah Produk.
		</p>
	{:else}
		<ul>
			{#each keranjangState.items as item (item.produk.id)}
				{@const terlaluBanyak = item.qty > item.produk.stock}
				<li class="grid grid-cols-[1fr_auto] gap-x-2 gap-y-1 border-b border-border px-2 py-1.5">
					<span class="text-[13px] font-medium">{item.produk.name}</span>
					<span class="text-right text-[13px] font-semibold tabular-nums">
						{formatRupiah(item.produk.price * item.qty)}
					</span>

					<span class="col-span-2 text-xs tabular-nums">
						{formatRupiah(item.produk.price)} · Stok {item.produk.stock}
					</span>

					<span class="col-span-2 flex items-center gap-1">
						<span class="flex items-center border border-border">
							<button
								type="button"
								class={`${OPERATOR} border-r`}
								disabled={item.qty <= 1}
								aria-label={`Kurangi jumlah ${item.produk.name}`}
								onclick={() => keranjangState.setQuantity(item.produk.id, item.qty - 1)}
							>
								−
							</button>
							<Input
								type="number"
								min="1"
								step="1"
								inputmode="numeric"
								class={JUMLAH}
								value={item.qty}
								aria-label={`Jumlah ${item.produk.name}`}
								aria-invalid={terlaluBanyak ? true : undefined}
								oninput={(event) => setQuantity(event, item.produk.id)}
							/>
							<button
								type="button"
								class={`${OPERATOR} border-l`}
								aria-label={`Tambah jumlah ${item.produk.name}`}
								onclick={() => keranjangState.setQuantity(item.produk.id, item.qty + 1)}
							>
								+
							</button>
						</span>

						<Button
							variant="ghost"
							size="sm"
							class="h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground"
							aria-label={`Hapus ${item.produk.name} dari keranjang`}
							onclick={() => keranjangState.remove(item.produk.id)}
						>
							Hapus
						</Button>
					</span>

					{#if terlaluBanyak}
						<!--
							Barisnya sendiri yang menyebut keadaannya; tanda merahnya dibawa
							oleh field jumlah di atasnya (DESIGN.md, Fields & Inputs), bukan
							oleh tint huruf yang harus tetap bisa dibaca.
						-->
						<p class="col-span-2 text-xs font-semibold">
							Melebihi Stok: tersisa {item.produk.stock}. Kurangi jumlahnya sebelum checkout.
						</p>
					{/if}
				</li>
			{/each}
		</ul>

		<!--
			Pita total: satu-satunya angka besar di kolom ini, dan satu-satunya tempat
			aksi keranjang sendiri berdiri. Garis tintanya yang memisahkannya dari
			Pembayaran di bawah — bukan bayangan, tidak pernah bayangan.
		-->
		<div
			class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
		>
			<span
				class="flex items-baseline gap-2"
				role="status"
				aria-label={`Total keranjang ${formatRupiah(keranjangState.total)}`}
			>
				<span class="text-[13px] font-bold">Total</span>
				<span class="text-[15px] font-bold tabular-nums">
					{formatRupiah(keranjangState.total)}
				</span>
			</span>

			<Button
				variant="ghost"
				size="sm"
				class="h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground"
				onclick={() => keranjangState.clear()}
			>
				Kosongkan
			</Button>
		</div>
	{/if}
</section>
