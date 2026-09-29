<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { SessionMenu } from '$lib/domains/auth';
	import { createStoreNameQuery } from '$lib/domains/pengaturan';
	import { cn } from '$lib/utils';

	let { data, children } = $props();

	// The till and the Penjualan lookup are the Kasir's screens and the Admin's
	// too: at a one-terminal store the owner is behind the counter as often as the
	// Kasir is, and CONTEXT.md gives the Kasir "cetak Struk" — so both entries are
	// for both Peran. Only an Admin sees the catalogue, the Stok and the Pengguna
	// entries, and the route guard backs that up.
	//
	// `resolve` takes SvelteKit's route ids, which keep the group prefix — it is
	// what turns `/(app)/pengguna` into the `/pengguna` a link needs.
	const utama = [
		{ href: resolve('/'), label: 'Beranda' },
		{ href: resolve('/(app)/kasir'), label: 'Kasir' },
		{ href: resolve('/(app)/penjualan'), label: 'Penjualan' }
	];

	const admin = [
		{ href: resolve('/(app)/produk'), label: 'Produk' },
		{ href: resolve('/(app)/stok'), label: 'Stok' },
		{ href: resolve('/(app)/laporan'), label: 'Laporan' },
		{ href: resolve('/(app)/pengguna'), label: 'Pengguna' },
		{ href: resolve('/(app)/pengaturan'), label: 'Pengaturan' },
		{ href: resolve('/(app)/backup'), label: 'Backup' }
	];

	/**
	 * Two rows of tabs, never one long strip: DESIGN.md fixes the rail's shape as
	 * the Kasir's three screens above the Admin's six, so a store that grows adds a
	 * row instead of pushing a tab off the edge. A Kasir sees only the first row,
	 * which is the same guard the API enforces — the rail never offers a screen the
	 * Pengguna cannot open.
	 */
	const rows = $derived(data.user?.role === 'admin' ? [utama, admin] : [utama]);

	/**
	 * The store's own name is the rail's identity (PRODUCT.md, dikonfirmasi
	 * pemilik): it leads on every screen, so it is read once here instead of per
	 * screen. A store that has not been named yet falls back to the product's name
	 * rather than a blank line, and while it loads nothing is shown — a flash of
	 * the wrong name is what this line exists to remove (ADR-0019).
	 */
	const storeName = createStoreNameQuery();

	/**
	 * One tab, styled in one place: the same classes copied onto nine links would
	 * drift apart. The active tab is filled with the utility red — one of the only
	 * two jobs red has in this world, so it is never spent on anything else here.
	 */
	const TAB =
		'inline-flex items-center gap-1.5 border-r border-border px-2.5 py-1.5 text-[13px] whitespace-nowrap transition-colors hover:bg-muted aria-[current=page]:bg-destructive aria-[current=page]:font-semibold aria-[current=page]:text-primary-foreground aria-[current=page]:hover:bg-destructive';
</script>

<!--
  Rel navigasi (DESIGN.md): satu baris identitas di atas, lalu baris-baris tab,
  menempel ke tepi atas. Garis tinta menutup rel dari papan; garis rambut yang
  memisahkan baris-baris di dalamnya. Tanpa bayangan, tanpa radius.
-->
<header class="sticky top-0 z-50 border-b border-foreground bg-card">
	<div class="flex items-stretch border-b border-border">
		<span
			class="flex items-center border-r border-border px-2.5 py-1.5 text-[15px] font-bold whitespace-nowrap"
		>
			{#if storeName.isPending}
				<span class="sr-only">Memuat nama toko…</span>
			{:else}
				{storeName.data || 'Lite Point of Sale'}
			{/if}
		</span>

		{#if data.user}
			<div class="ml-auto flex items-center border-l border-border px-2.5 py-1.5">
				<SessionMenu user={data.user} />
			</div>
		{/if}
	</div>

	<nav aria-label="Navigasi utama">
		{#each rows as row, index (index)}
			<div class={cn('flex flex-wrap items-stretch', index > 0 && 'border-t border-border')}>
				{#each row as item (item.href)}
					<a
						href={item.href}
						aria-current={page.url.pathname === item.href ? 'page' : undefined}
						class={TAB}
					>
						{item.label}
					</a>
				{/each}
			</div>
		{/each}
	</nav>
</header>

<!--
  Papan selebar viewport dengan padding luar 8px, bukan kolom tengah yang
  mengambang: tiap layar menggambar modulnya sendiri selebar papan ini.
-->
<main class="w-full p-2">
	{@render children()}
</main>
