<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { SessionMenu } from '$lib/domains/auth';

	let { data, children } = $props();

	// The till is the Kasir's screen and the Admin's too: at a one-terminal store the
	// owner is behind the counter as often as the Kasir is, so "Kasir" is the one
	// entry both Peran see. Only an Admin sees the catalogue, the Stok and the
	// Pengguna entries, and the route guard backs that up.
	//
	// `resolve` takes SvelteKit's route ids, which keep the group prefix — it is
	// what turns `/(app)/pengguna` into the `/pengguna` a link needs.
	const navigation = $derived([
		{ href: resolve('/'), label: 'Beranda' },
		{ href: resolve('/(app)/kasir'), label: 'Kasir' },
		...(data.user?.role === 'admin'
			? [
					{ href: resolve('/(app)/produk'), label: 'Produk' },
					{ href: resolve('/(app)/stok'), label: 'Stok' },
					{ href: resolve('/(app)/pengguna'), label: 'Pengguna' },
					{ href: resolve('/(app)/pengaturan'), label: 'Pengaturan' }
				]
			: [])
	]);
</script>

<div class="min-h-svh">
	<header class="border-b">
		<div class="mx-auto flex w-full max-w-4xl flex-wrap items-center justify-between gap-4 p-4">
			<div class="flex items-center gap-6">
				<span class="font-semibold">Lite Point of Sale</span>
				<nav class="flex gap-4 text-sm">
					{#each navigation as item (item.href)}
						<a
							href={item.href}
							class={page.url.pathname === item.href
								? 'font-medium'
								: 'text-muted-foreground hover:text-foreground'}
						>
							{item.label}
						</a>
					{/each}
				</nav>
			</div>

			{#if data.user}
				<SessionMenu user={data.user} />
			{/if}
		</div>
	</header>

	<main class="mx-auto w-full max-w-4xl p-4">
		{@render children()}
	</main>
</div>
