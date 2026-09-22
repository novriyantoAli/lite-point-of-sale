<script lang="ts">
	import { HealthStatus } from '$lib/domains/health';
</script>

<main class="mx-auto flex min-h-svh w-full max-w-2xl flex-col justify-center gap-6 p-6">
	<header class="space-y-1">
		<h1 class="text-2xl font-semibold">Lite Point of Sale</h1>
		<p class="text-sm text-muted-foreground">Kasir untuk satu toko, satu terminal.</p>
	</header>

	<HealthStatus />
</main>
