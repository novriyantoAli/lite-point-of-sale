<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';
	import TambahStokForm from './TambahStokForm.svelte';
	import { createProdukListQuery, createStokMenipisQuery } from '../queries/produk.queries';
	import type { Produk } from '../schemas/produk.schema';

	/**
	 * The Stok screen. Two lists answer two different questions — "what do I have
	 * to restock now" and "where does every Produk stand" — and both offer the same
	 * restock form, because an Admin who came for one of them should not have to
	 * walk over to the other to act on it.
	 *
	 * Every Produk, unfiltered: this is not the catalogue and does not own the
	 * catalogue's filters, so it asks for the whole list rather than reading
	 * `produkFilterState`. An Admin who narrowed the catalogue to one Kategori must
	 * not find half their Stok missing here.
	 */
	const catalogue = createProdukListQuery(() => ({ name: '', code: '', category: '' }));
	/** What to restock, with the threshold that selected it (the domain's rule). */
	const menipis = createStokMenipisQuery();

	/**
	 * The row whose restock form is open. The Produk alone is not enough to name it:
	 * the same Produk appears in both lists, and a form that opened in both at once
	 * would be two inputs for one delivery — the Admin would type into one and the
	 * other would keep a stale amount.
	 */
	type RestockPlace = 'menipis' | 'katalog';
	let restocking = $state<{ place: RestockPlace; id: number } | null>(null);
	let notice = $state('');

	function isRestocking(produk: Produk, place: RestockPlace): boolean {
		return restocking?.place === place && restocking.id === produk.id;
	}

	function startRestocking(produk: Produk, place: RestockPlace) {
		restocking = { place, id: produk.id };
		notice = '';
	}

	function restocked(produk: Produk) {
		notice = `Stok ${produk.name} sekarang ${produk.stock}.`;
		restocking = null;
	}

	/**
	 * Papan mosaik (DESIGN.md, Layout): satu kolom selebar papan — layar Stok bukan
	 * papan tiga kolom milik Kasir. Padding luarnya milik rel (`(app)/+layout`), jadi
	 * di sini tidak ada padding lagi.
	 */
	const PAPAN = 'grid grid-cols-1 gap-0';
	/**
	 * Strip judul: satu-satunya tempat ukuran 24px muncul di layar ini. `-mb-px`
	 * plus `z-[2]` menariknya turun satu piksel, jadi garis tintanya yang menutup
	 * modul di bawahnya alih-alih bertumpuk dengan garis rambut (DESIGN.md,
	 * Shared-Hairline).
	 */
	const STRIP =
		'z-[2] col-span-full -mb-px flex flex-wrap items-baseline gap-x-3 gap-y-1 border border-border border-b-foreground bg-card px-2 py-2';
	const JUDUL = 'text-2xl leading-none font-bold tracking-[-0.015em]';
	/** Seluruh teks 12px; lantai huruf dunia ini (DESIGN.md, Legibility Floor). */
	const CATATAN = 'text-xs';
	/**
	 * Cacah dan ambang di kepala modul. `tabular-nums` bukan hiasan di sini:
	 * DESIGN.md (Typography) menuntutnya untuk "setiap angka uang, jumlah, dan
	 * Stok", tanpa kecuali, dan prototipe menulis kelas `tnum` di catatan kepala.
	 */
	const CATATAN_ANGKA = `${CATATAN} tabular-nums`;
	/**
	 * Modul: latar petak dengan garis rambut, dan hanya garis bawahnya yang
	 * digambar — dua modul bersebelahan berbagi satu garis, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const MODUL = 'border border-border border-b-0 bg-card';
	/**
	 * Jarak 8px sebelum modul Stok per Produk bukan kelalaian: prototipe layar ini
	 * sendiri yang memisahkan dua modul dengan `margin-top: 8px`, jadi keduanya
	 * memang tidak bersebelahan.
	 */
	const MODUL_KATALOG = `${MODUL} mt-2`;
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/** Satu baris catatan atau keadaan di dalam modul, selebar modulnya. */
	const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
	const MODUL_CATATAN = `${MODUL_BARIS} text-xs`;
	/** Daftar Stok menipis: baris nama + keadaan + aksi, satu garis rambut bersama. */
	const DAFTAR = 'divide-y divide-border border-b border-border';
	const LIROW = 'flex flex-wrap items-center justify-between gap-x-3 gap-y-1.5 px-2 py-1.5';
	const LIROW_UTAMA = 'min-w-0';
	const LIROW_NAMA = 'text-[13px] font-medium';
	const LIROW_KODE = 'text-xs';
	const LIROW_META = 'flex flex-wrap items-center gap-x-2 gap-y-1 text-xs';
	const LIROW_AKSI = 'flex items-center gap-1';
	/** Tabel: satu garis rambut bersama antar baris, kepala Wash Grey + garis tinta. */
	const TABEL = 'w-full border-collapse text-[13px] leading-[1.25]';
	const TH =
		'border-b border-foreground bg-muted px-2 py-1.5 text-left text-xs font-semibold whitespace-nowrap';
	const TH_NUM = `${TH} text-right tabular-nums`;
	const BARIS = 'border-b border-border transition-colors last:border-b-0 hover:bg-muted';
	const TD = 'px-2 py-1.5 align-top';
	const TD_NUM = `${TD} text-right`;
	const TD_SUB = `${TD} text-xs`;
	const TD_AKSI = `${TD} whitespace-nowrap`;
	const TD_NAMA = `${TD} font-medium`;
	/**
	 * Coretan garis: cara dunia ini menyampaikan keadaan mati, bukan `opacity`
	 * (DESIGN.md, State-Is-Not-Faded; Do's: "let a disabled control lose its ink or
	 * take a strike, never its opacity"). Produk yang habis kehilangan tintanya —
	 * namanya di tabel dan di daftar dicoret, bukan dipudarkan.
	 */
	const CORET = 'line-through decoration-1';
	/** Quiet tag: Wash Grey bergaris rambut, tinggi 15px (DESIGN.md, Tags). */
	const TAG =
		'inline-flex h-[15px] items-center border border-border bg-muted px-[5px] text-xs font-semibold whitespace-nowrap';
	const TAG_CORET = `${TAG} ${CORET}`;
	/** Kata keadaan yang tenang, 12px/600, tanpa kotak: Aman dan angka telanjang. */
	const KEADAAN = 'text-xs font-semibold';
	/**
	 * Tiga petak tombol dunia ini, dan yang membedakan mereka ukurannya, bukan
	 * warnanya. `.btn` di prototipe tinggi 26px dengan huruf 13px; ukuran 22px/12px
	 * hanya milik `.tbl__acts .btn`, yaitu aksi di dalam sel tabel.
	 *
	 * Selalu dipakai dengan `variant="ghost"`: varian itu sudah membawa
	 * `hover:bg-muted` — persis latar hover dunia ini — dan tidak membawa apa pun
	 * yang harus dilawan (solution doc §3).
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
	const AKSI_BARIS =
		'h-[22px] border-border bg-card px-1.5 text-xs font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';

	/**
	 * Whether a Produk has run out and is still for sale — the one case the world
	 * marks with a strike rather than an opacity (DESIGN.md, Tiles: "Produk yang
	 * habis kehilangan tintanya: nama dan harganya dicoret garis"). It strikes the
	 * Nama in both the restock list and the table; the word that names the state is
	 * written beside the Stok number.
	 */
	function habis(produk: Produk): boolean {
		return produk.active && produk.stock === 0;
	}
</script>

{#snippet restock(produk: Produk, place: RestockPlace)}
	{#if isRestocking(produk, place)}
		<TambahStokForm {produk} onAdded={restocked} onCancel={() => (restocking = null)} />
	{:else}
		<Button
			variant="ghost"
			class={place === 'menipis' ? AKSI_MODUL : AKSI_BARIS}
			onclick={() => startRestocking(produk, place)}
		>
			Tambah Stok
		</Button>
	{/if}
{/snippet}

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya
		(DESIGN.md, Typography). Layar Stok tidak punya aksi utama di strip ini —
		aksi milik tiap baris, karena tiap baris adalah satu Produk yang berbeda.
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Stok</h1>
		<p class={CATATAN}>
			Catat barang masuk dan lihat Produk mana yang perlu ditambah. Stok berkurang sendiri saat
			Produk terjual.
		</p>
	</div>

	<!--
		Pemberitahuan hasil restock berdiri sebagai barisnya sendiri di bawah
		strip, bukan di dalam salah satu modul: dua modul itu adalah daftar Produk,
		dan nama Produk yang baru di-restock di sana akan berbohong tentang apa yang
		masih menipis. Baris ini berbagi garis rambut dengan strip lewat `-mb-px`
		milik strip, tanpa margin negatif kedua di sisinya. Angkanya Stok, jadi ia
		`tabular-nums` (DESIGN.md, Typography).
	-->
	{#if notice}
		<p
			class="border border-b-0 border-border bg-card px-2 py-1.5 text-xs font-semibold tabular-nums"
			role="status"
		>
			{notice}
		</p>
	{/if}

	<section class={MODUL} aria-labelledby="stok-menipis-judul" aria-busy={menipis.isPending}>
		<div class={MODUL_KEPALA}>
			<h2 id="stok-menipis-judul" class={MODUL_JUDUL}>Stok menipis</h2>
			<span class={CATATAN_ANGKA}>
				{#if menipis.data}
					ambang {menipis.data.threshold} · paling sedikit di atas
				{:else}
					paling sedikit di atas
				{/if}
			</span>
		</div>

		<!--
			Ambangnya adalah angka Stok, dan prototipe menandai digitnya `tnum`:
			`tabular-nums` di paragraf ini menahan lebar digitnya saat Admin mengubah
			ambang di Pengaturan (DESIGN.md, Typography: "setiap angka ... Stok",
			tanpa kecuali).
		-->
		<div class={`${MODUL_CATATAN} tabular-nums`}>
			{#if menipis.data}
				Produk Aktif dengan Stok di bawah {menipis.data.threshold}, yang paling sedikit di atas.
				Produk Nonaktif tidak dihitung — ia tidak sedang dijual.
			{:else}
				Produk Aktif dengan Stok paling sedikit, yang perlu ditambah.
			{/if}
		</div>

		{#if menipis.isPending}
			<p class={MODUL_CATATAN}>Memuat Stok menipis…</p>
		{:else if menipis.error}
			<div class={MODUL_BARIS}>
				<p class="text-xs font-semibold" role="alert">{menipis.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void menipis.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if menipis.data?.products.length === 0}
			<p class={MODUL_CATATAN}>
				Tidak ada Produk dengan Stok menipis. Semua Produk Aktif masih punya Stok di ambang atau
				lebih.
			</p>
		{:else}
			<ul class={DAFTAR}>
				{#each menipis.data?.products ?? [] as produk (produk.id)}
					<li class={LIROW}>
						<div class={LIROW_UTAMA}>
							<p class={habis(produk) ? cn(LIROW_NAMA, CORET) : LIROW_NAMA}>
								{produk.name}
								{#if produk.code}
									<span class={LIROW_KODE}>· {produk.code}</span>
								{/if}
							</p>
							<!--
								Kata menemani tandanya (DESIGN.md, Do's): `menipis` sebagai
								quiet tag dengan "Sisa 2", `habis` dengan "Stok habis".
							-->
							<p class={LIROW_META}>
								{#if produk.stock === 0}
									<span class={TAG}>habis</span> Stok habis
								{:else}
									<span class={TAG}>menipis</span> Sisa
									<span class="tabular-nums">{produk.stock}</span>
								{/if}
							</p>
						</div>
						<div class={LIROW_AKSI}>
							{@render restock(produk, 'menipis')}
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section
		class={MODUL_KATALOG}
		aria-labelledby="stok-produk-judul"
		aria-busy={catalogue.isPending}
	>
		<div class={MODUL_KEPALA}>
			<h2 id="stok-produk-judul" class={MODUL_JUDUL}>Stok per Produk</h2>
			<span class={CATATAN_ANGKA}>
				{catalogue.isPending
					? 'memuat…'
					: catalogue.error
						? '—'
						: `${catalogue.data?.length ?? 0} Produk · tanpa saringan`}
			</span>
		</div>

		{#if catalogue.isPending}
			<p class={MODUL_CATATAN}>Memuat Produk…</p>
		{:else if catalogue.error}
			<div class={MODUL_BARIS}>
				<p class="text-xs font-semibold" role="alert">{catalogue.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void catalogue.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if catalogue.data?.length === 0}
			<p class={MODUL_CATATAN}>Belum ada Produk. Tambahkan yang pertama di layar Produk.</p>
		{:else}
			<div class="overflow-x-auto border-b border-border">
				<table class={TABEL}>
					<thead>
						<tr>
							<th scope="col" class={TH}>Nama</th>
							<th scope="col" class={TH}>Kode</th>
							<th scope="col" class={TH_NUM}>Stok</th>
							<th scope="col" class={TH}>Aksi</th>
						</tr>
					</thead>
					<tbody>
						{#each catalogue.data ?? [] as produk (produk.id)}
							<tr class={BARIS}>
								<td class={habis(produk) ? cn(TD_NAMA, CORET) : TD_NAMA}>{produk.name}</td>

								<!--
									Kode yang kosong tidak boleh terlihat seperti data yang hilang:
									ia menuliskan keadaannya (DESIGN.md, Do's).
								-->
								<td class={TD_SUB}>
									{#if produk.code}
										<span class="tabular-nums">{produk.code}</span>
									{:else}
										tanpa Kode
									{/if}
								</td>

								<!--
									Angka dulu, lalu kata keadaannya: "4 · menipis" dan "Stok habis"
									dari prototipe, dengan cacahnya tetap terbaca supaya restock
									diukur dari angka, bukan dari kata. `tabular-nums` menahan
									lebarnya saat Stok berubah (DESIGN.md, Typography).
								-->
								<td class={TD_NUM}>
									<span class="inline-flex items-center justify-end gap-1.5">
										<span class="tabular-nums">{produk.stock}</span>
										{#if !produk.active}
											<!--
												Produk yang tidak dijual bukan Produk yang habis, jadi
												ia tidak disebut habis di sini (CONTEXT.md, Nonaktif).
											-->
											<span class={TAG_CORET}>Nonaktif</span>
										{:else if produk.stock === 0}
											<span class={cn(KEADAAN, CORET)}>Habis</span>
										{:else if menipis.data === undefined}
											<!--
												Ambang adalah arti kata "menipis", jadi tanpa ambangnya
												layar ini tidak boleh menyebut apa pun aman — itu tebakan,
												dan modul di atas sudah mengatakan daftarnya gagal dibaca.
											-->
											<span class={KEADAAN}>—</span>
										{:else if produk.stock < menipis.data.threshold}
											<span class={TAG}>Menipis</span>
										{:else}
											<span class={KEADAAN}>Aman</span>
										{/if}
									</span>
								</td>

								<td class={TD_AKSI}>
									{@render restock(produk, 'katalog')}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>
</div>
