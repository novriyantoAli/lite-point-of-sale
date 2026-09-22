<script lang="ts">
	import { HealthStatus } from '$lib/domains/health';
</script>

<div class="space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Beranda</h1>
		<p class="text-sm text-muted-foreground">Kasir untuk satu toko, satu terminal.</p>
	</div>

	<HealthStatus />
</div>
