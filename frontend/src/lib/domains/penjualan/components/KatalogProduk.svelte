<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { createProdukListQuery, type Produk } from '$lib/domains/produk';
	import { formatRupiah } from '$lib/utils';
	import { pencarianState } from '../state/pencarian.state.svelte';

	/**
	 * Mosaik katalog: setiap Produk yang boleh dijual, satu petak per Produk, dan
	 * petaknya sendiri yang menambahkannya ke keranjang (CONTEXT.md, Kode).
	 *
	 * Query-nya tinggal di kolom ini, bukan di kolom pencarian: kolom inilah yang
	 * menggambar jawabannya, dan keadaan memuat, galat, serta kosong semuanya milik
	 * mosaik yang tidak jadi tergambar (skill §11).
	 */
	let { onAdd }: { onAdd: (produk: Produk) => void } = $props();

	/**
	 * `active: true` is the whole of "the kasir lookup": a Produk that is Nonaktif
	 * is not for sale, so it is not offered (CONTEXT.md, Nonaktif).
	 */
	const list = createProdukListQuery(() => ({
		name: pencarianState.name,
		code: pencarianState.code,
		category: '',
		active: true
	}));

	const found = $derived(list.data ?? []);
	const narrowed = $derived(Boolean(pencarianState.code || pencarianState.name));

	/**
	 * Angka yang belum bisa dihitung ditulis "—", bukan 0 (DESIGN.md, Figures).
	 * Sebuah galat juga belum bisa dihitung: menulis 0 akan berbohong tentang
	 * katalog yang tidak pernah datang.
	 */
	const belumDiketahui = $derived(list.isPending || Boolean(list.error));

	/**
	 * Petak hantu selama memuat: sebentuk mosaik dengan ukuran petak yang sama,
	 * supaya layar tidak melompat saat Produknya datang — dan supaya yang terlihat
	 * bukan layar kosong yang tak bisa dibedakan dari katalog yang memang kosong.
	 */
	const hantu = Array.from({ length: 24 }, (_, i) => i);

	/**
	 * Satu petak. Seluruh petak adalah tombolnya: satu Produk satu aksi, dan
	 * aksinya tidak muncul-muncul saat kursor lewat (DESIGN.md, Named-Not-Hidden).
	 *
	 * Produk yang Stoknya habis tetap menampilkan namanya dengan garis coret, bukan
	 * dipudarkan: keadaannya ditulis, dan petaknya mati (DESIGN.md,
	 * State-Is-Not-Faded).
	 */
	const PETAK =
		'group flex h-full w-full flex-col gap-1 border-r border-b border-border bg-card px-2 pt-1.5 pb-2 text-left transition-colors enabled:hover:border-foreground enabled:hover:bg-muted disabled:cursor-not-allowed';
	const KODE = 'inline-flex h-[15px] items-center self-start px-[5px] text-xs font-semibold';
</script>

<section class="flex flex-col" aria-labelledby="kasir-katalog-judul">
	<div
		class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5"
	>
		<h2 id="kasir-katalog-judul" class="text-[13px] font-bold tracking-[0.01em]">Katalog</h2>
		<span class="text-xs">
			{#if belumDiketahui}—{:else}{found.length}{/if} Produk Aktif · tekan satu petak untuk menambah
		</span>
	</div>

	{#if list.isPending}
		<!--
			Kerangkanya bukan lingkaran berputar di tengah layar: petak-petaknya
			digambar pada ukuran dan jumlah yang sama seperti petak sungguhan, jadi
			papan di bawahnya tidak bergerak saat datanya tiba.
		-->
		<ul
			class="-mr-px grid grid-cols-[repeat(auto-fill,minmax(154px,1fr))]"
			aria-busy="true"
			aria-label="Memuat katalog Produk"
		>
			{#each hantu as petak (petak)}
				<li class="h-[73px] border-r border-b border-border bg-muted"></li>
			{/each}
		</ul>
		<p class="sr-only" role="status">Memuat Produk…</p>
	{:else if list.error}
		<div class="border-b border-border px-2 py-1.5">
			<p class="text-xs font-semibold">{list.error.message}</p>
			<Button
				variant="ghost"
				size="sm"
				class="mt-1.5 h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground"
				onclick={() => void list.refetch()}
			>
				Coba lagi
			</Button>
		</div>
	{:else if found.length === 0}
		<p class="border-b border-border px-2 py-1.5 text-xs">
			{narrowed
				? 'Tidak ada Produk Aktif yang cocok. Periksa Kode atau namanya.'
				: 'Belum ada Produk yang bisa dijual. Admin perlu menambahkannya lebih dulu.'}
		</p>
	{:else}
		<ul
			class="-mr-px grid grid-cols-[repeat(auto-fill,minmax(154px,1fr))]"
			aria-label="Hasil pencarian Produk"
		>
			{#each found as produk (produk.id)}
				<li>
					<button
						type="button"
						class={PETAK}
						disabled={produk.stock === 0}
						aria-label={`Tambah ${produk.name} ke keranjang`}
						onclick={() => onAdd(produk)}
					>
						{#if produk.code}
							<span class={`${KODE} bg-destructive text-primary-foreground tabular-nums`}>
								{produk.code}
							</span>
						{:else}
							<!-- Kolom Kode yang kosong tidak boleh terlihat seperti data yang hilang. -->
							<span class={`${KODE} border border-border bg-muted`}>tanpa Kode</span>
						{/if}

						<span
							class="text-[13px] leading-[1.2] font-medium group-disabled:line-through group-disabled:decoration-1"
						>
							{produk.name}
						</span>

						<span class="mt-auto flex items-baseline justify-between gap-2">
							<span class="text-xs tabular-nums">
								{produk.stock === 0 ? 'Stok habis' : `Stok ${produk.stock}`}
							</span>
							<span
								class="text-[13px] font-semibold text-destructive tabular-nums group-disabled:line-through group-disabled:decoration-1"
							>
								{formatRupiah(produk.price)}
							</span>
						</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</section>
