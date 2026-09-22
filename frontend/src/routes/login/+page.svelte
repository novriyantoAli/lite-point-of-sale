<script lang="ts">
	import { LoginForm } from '$lib/domains/auth';
</script>

<main class="flex min-h-svh w-full flex-col items-center justify-center gap-6 p-6">
	<div class="space-y-1 text-center">
		<h1 class="text-2xl font-semibold">Lite Point of Sale</h1>
		<p class="text-sm text-muted-foreground">Kasir untuk satu toko, satu terminal.</p>
	</div>

	<LoginForm />
</main>
