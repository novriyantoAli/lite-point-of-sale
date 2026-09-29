<script lang="ts">
	import {
		AKSI_MODUL,
		CATATAN,
		COMMIT_KOLOM,
		ERROR,
		FIELD,
		INPUT_ANGKA,
		JUDUL,
		LABEL,
		METODE,
		MODUL,
		MODUL_BARIS,
		MODUL_JUDUL,
		MODUL_KEPALA,
		MODUL_TEKS,
		PAPAN,
		STRIP
	} from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { cn, collectFieldErrors } from '$lib/utils';
	import {
		createPengaturanQuery,
		createUpdatePengaturanMutation
	} from '../queries/pengaturan.queries';
	import { PAPER_WIDTH_OPTIONS, UpdatePengaturanInputSchema } from '../schemas/pengaturan.schema';

	/**
	 * The Pengaturan screen: the store's one row of settings, edited by an Admin.
	 * The three things it changes are the Struk template (header and footer as
	 * free-text blocks), the paper width (58/80 mm), and the ambang Stok menipis
	 * (ADR-0017, keputusan 3 and 4).
	 */
	const pengaturan = createPengaturanQuery();
	const update = createUpdatePengaturanMutation();

	/**
	 * The fields stay strings: a form submits text, and the schema is the one place
	 * that turns that text into the integers the API takes — the same reason the
	 * Produk form keeps Harga and Stok as strings (§11).
	 *
	 * They are seeded once, when the query first answers, and left alone after:
	 * re-seeding on a later refetch would fight the person typing into them.
	 */
	let header = $state('');
	let footer = $state('');
	let paperWidth = $state<'58' | '80'>('80');
	let threshold = $state('');
	let fieldErrors = $state<Partial<Record<'paper_width' | 'low_stock_threshold', string>>>({});
	let notice = $state('');

	let seeded = $state(false);
	$effect(() => {
		if (seeded || !pengaturan.data) {
			return;
		}
		header = pengaturan.data.header;
		footer = pengaturan.data.footer;
		paperWidth = pengaturan.data.paper_width === 58 ? '58' : '80';
		threshold = String(pengaturan.data.low_stock_threshold);
		seeded = true;
	});

	const pending = $derived(update.isPending);
	const error = $derived(update.error);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};
		notice = '';

		const parsed = UpdatePengaturanInputSchema.safeParse({
			header,
			footer,
			paper_width: paperWidth,
			low_stock_threshold: threshold
		});
		if (!parsed.success) {
			fieldErrors = collectFieldErrors(parsed.error, ['paper_width', 'low_stock_threshold']);
			return;
		}

		try {
			await update.mutateAsync(parsed.data);
			// Keep the threshold field in step with what was stored (a number), so
			// the text the person reads is the number the next load will show.
			threshold = String(parsed.data.low_stock_threshold);
			notice = 'Pengaturan disimpan.';
		} catch {
			// `error` carries the normalized message, rendered below.
		}
	}

	/**
	 * `.textarea` prototipe: min-height 88px, padding 6px, huruf 13px/1.4, garis
	 * rambut, radius nol, `resize: vertical`. Tidak ada primitif shadcn untuknya,
	 * jadi pakaiannya ditulis di sini.
	 */
	const TEXTAREA =
		'min-h-[88px] w-full resize-y rounded-none border border-border bg-card p-1.5 text-[13px] leading-[1.4] text-foreground shadow-none outline-none focus-visible:border-foreground';
</script>

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya
		(DESIGN.md, Typography). Layar ini tidak punya aksi di strip: satu-satunya
		aksi adalah Simpan, dan ia menutup modulnya.
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Pengaturan</h1>
		<p class={CATATAN}>Setelan toko: template Struk dan ambang Stok menipis.</p>
	</div>

	<section class={MODUL} aria-labelledby="setelan-judul" aria-busy={pengaturan.isPending}>
		<div class={MODUL_KEPALA}>
			<h2 id="setelan-judul" class={MODUL_JUDUL}>Setelan toko</h2>
			<span class={CATATAN}>satu baris, tersimpan utuh</span>
		</div>

		{#if pengaturan.isPending}
			<p class={MODUL_TEKS}>Memuat Pengaturan…</p>
		{:else if pengaturan.error}
			<div class={MODUL_BARIS}>
				<p class="text-xs font-semibold" role="alert">{pengaturan.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void pengaturan.refetch()}>Coba lagi</Button
				>
			</div>
		{:else}
			<form aria-label="Formulir Pengaturan" onsubmit={submit} novalidate>
				<div class={MODUL_BARIS}>
					<div class={FIELD}>
						<Label for="pengaturan-header" class={LABEL}>Header Struk</Label>
						<textarea
							id="pengaturan-header"
							name="header"
							rows={4}
							bind:value={header}
							class={TEXTAREA}></textarea>
						<p class="text-xs">
							Blok teks di atas baris Item pada Struk — misalnya nama dan alamat toko. Baris pertama
							yang tidak kosong di sini menjadi nama toko di rel dan layar Masuk.
						</p>
					</div>
				</div>

				<div class={MODUL_BARIS}>
					<div class={FIELD}>
						<Label for="pengaturan-footer" class={LABEL}>Footer Struk</Label>
						<textarea
							id="pengaturan-footer"
							name="footer"
							rows={4}
							bind:value={footer}
							class={TEXTAREA}></textarea>
						<p class="text-xs">
							Blok teks di bawah total pada Struk — misalnya ucapan terima kasih.
						</p>
					</div>
				</div>

				<!--
					Kisi field menumpuk di 900px, bukan di titik papan 1080px
					(ADR-0020): kisi field bukan papan.
				-->
				<div
					class={cn(MODUL_BARIS, 'grid grid-cols-1 gap-x-3 gap-y-[10px] min-[901px]:grid-cols-2')}
				>
					<div class={FIELD}>
						<p class={LABEL} id="lebar-kertas-label">Lebar kertas</p>
						<!--
							Native radios, so arrow keys move between the two widths and the
							group is announced as one choice. The input is only hidden from
							sight: the cell is what shows the width and what carries the focus
							ring (DESIGN.md, Methods).
						-->
						<div
							class="grid grid-cols-2 gap-1"
							role="radiogroup"
							aria-labelledby="lebar-kertas-label"
						>
							{#each PAPER_WIDTH_OPTIONS as option (option.value)}
								<label class={METODE}>
									<input
										type="radio"
										name="paper_width"
										value={option.value}
										bind:group={paperWidth}
										class="sr-only"
									/>
									{option.label}
								</label>
							{/each}
						</div>
						{#if fieldErrors.paper_width}
							<p class={ERROR}>{fieldErrors.paper_width}</p>
						{/if}
						<p class="text-xs">Lebar kertas thermal: 58 atau 80 mm.</p>
					</div>

					<div class={FIELD}>
						<Label for="pengaturan-ambang" class={LABEL}>Ambang Stok menipis</Label>
						<Input
							id="pengaturan-ambang"
							name="low_stock_threshold"
							inputmode="numeric"
							autocomplete="off"
							class={INPUT_ANGKA}
							bind:value={threshold}
							aria-invalid={fieldErrors.low_stock_threshold ? true : undefined}
						/>
						{#if fieldErrors.low_stock_threshold}
							<p class={ERROR}>{fieldErrors.low_stock_threshold}</p>
						{/if}
						<p class="text-xs">
							Produk Aktif dengan Stok di bawah angka ini masuk daftar Stok menipis.
						</p>
					</div>
				</div>

				{#if error}
					<p class={cn(MODUL_BARIS, ERROR)} role="alert">{error.message}</p>
				{/if}

				{#if notice}
					<p class={cn(MODUL_BARIS, ERROR)} role="status">{notice}</p>
				{/if}

				<div class="border-b border-border px-2 py-2">
					<Button type="submit" class={COMMIT_KOLOM} disabled={pending}>
						{pending ? 'Menyimpan…' : 'Simpan Pengaturan'}
					</Button>
				</div>
			</form>
		{/if}
	</section>
</div>
