<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { collectFieldErrors } from '$lib/utils';
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

	const paperWidthOption = $derived(
		PAPER_WIDTH_OPTIONS.find((option) => option.value === paperWidth) ?? PAPER_WIDTH_OPTIONS[1]
	);

	function setPaperWidth(value: string) {
		paperWidth = value === '58' ? '58' : '80';
	}

	const pending = $derived(update.isPending);
	const error = $derived(update.error);

	/**
	 * The shared style of the two template textareas. The block fields read alike
	 * because they are alike — header and footer are two textareas of the same
	 * kind, so the class lives here once rather than drifting between copies.
	 */
	const textareaClass =
		'w-full rounded-md border border-input bg-transparent px-2.5 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm';

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
</script>

<div class="space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Pengaturan</h1>
		<p class="text-sm text-muted-foreground">
			Setelan toko: template Struk dan ambang Stok menipis.
		</p>
	</div>

	{#if pengaturan.isPending}
		<p class="text-sm text-muted-foreground">Memuat Pengaturan…</p>
	{:else if pengaturan.error}
		<div class="space-y-3">
			<p class="text-sm text-destructive">{pengaturan.error.message}</p>
			<Button variant="outline" size="sm" onclick={() => void pengaturan.refetch()}>
				Coba lagi
			</Button>
		</div>
	{:else}
		<form
			class="space-y-6 rounded-lg border p-4"
			aria-label="Formulir Pengaturan"
			onsubmit={submit}
			novalidate
		>
			<div class="space-y-2">
				<Label for="pengaturan-header">Header Struk</Label>
				<textarea
					id="pengaturan-header"
					name="header"
					rows={4}
					bind:value={header}
					class={textareaClass}
				></textarea>
				<p class="text-sm text-muted-foreground">
					Blok teks di atas baris Item pada Struk — misalnya nama dan alamat toko.
				</p>
			</div>

			<div class="space-y-2">
				<Label for="pengaturan-footer">Footer Struk</Label>
				<textarea
					id="pengaturan-footer"
					name="footer"
					rows={4}
					bind:value={footer}
					class={textareaClass}
				></textarea>
				<p class="text-sm text-muted-foreground">
					Blok teks di bawah total pada Struk — misalnya ucapan terima kasih.
				</p>
			</div>

			<div class="grid gap-4 sm:grid-cols-2">
				<div class="space-y-2">
					<Label for="pengaturan-lebar">Lebar kertas</Label>
					<Select.Root type="single" value={paperWidth} onValueChange={setPaperWidth}>
						<Select.Trigger id="pengaturan-lebar" class="w-full">
							<span data-slot="select-value">{paperWidthOption.label}</span>
						</Select.Trigger>
						<Select.Content>
							{#each PAPER_WIDTH_OPTIONS as option (option.value)}
								<Select.Item value={option.value} label={option.label}>
									{option.label}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
					{#if fieldErrors.paper_width}
						<p class="text-sm text-destructive">{fieldErrors.paper_width}</p>
					{/if}
					<p class="text-sm text-muted-foreground">Lebar kertas thermal: 58 atau 80 mm.</p>
				</div>

				<div class="space-y-2">
					<Label for="pengaturan-ambang">Ambang Stok menipis</Label>
					<Input
						id="pengaturan-ambang"
						name="low_stock_threshold"
						inputmode="numeric"
						autocomplete="off"
						bind:value={threshold}
						aria-invalid={fieldErrors.low_stock_threshold ? true : undefined}
					/>
					{#if fieldErrors.low_stock_threshold}
						<p class="text-sm text-destructive">{fieldErrors.low_stock_threshold}</p>
					{/if}
					<p class="text-sm text-muted-foreground">
						Produk Aktif dengan Stok di bawah angka ini masuk daftar Stok menipis.
					</p>
				</div>
			</div>

			{#if error}
				<p class="text-sm text-destructive" role="alert">{error.message}</p>
			{/if}

			{#if notice}
				<p class="text-sm text-muted-foreground" role="status">{notice}</p>
			{/if}

			<div class="flex gap-2">
				<Button type="submit" disabled={pending}>
					{pending ? 'Menyimpan…' : 'Simpan Pengaturan'}
				</Button>
			</div>
		</form>
	{/if}
</div>
