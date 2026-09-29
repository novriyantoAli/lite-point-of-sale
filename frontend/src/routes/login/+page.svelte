<script lang="ts">
	import { LoginForm } from '$lib/domains/auth';
	import { createStoreNameQuery } from '$lib/domains/pengaturan';

	// The store's name is the one value the login screen may read before a
	// session exists (ADR-0019). While it loads nothing is shown, and an unnamed
	// store falls back to the product name rather than a blank line.
	const storeName = createStoreNameQuery();
</script>

<!--
  Layar Masuk adalah satu modul sempit di tengah, tanpa rel navigasi: sebelum
  login tidak ada sesi yang perlu diarahkan.
-->
<main class="flex min-h-svh w-full items-start justify-center bg-background p-6 pt-12">
	<section class="w-full max-w-[380px] border border-border bg-card">
		<div class="border-b border-border px-2 py-2">
			<h1 class="min-h-8 text-2xl font-bold tracking-[-0.015em]">
				{#if storeName.isPending}
					<span class="sr-only">Memuat nama toko…</span>
				{:else}
					{storeName.data || 'Lite Point of Sale'}
				{/if}
			</h1>
			<p class="mt-1 text-xs">Kasir untuk satu toko, satu terminal.</p>
		</div>

		<LoginForm />
	</section>
</main>
