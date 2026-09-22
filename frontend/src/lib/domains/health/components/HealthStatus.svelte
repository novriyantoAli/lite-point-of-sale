<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { createHealthQuery } from '../queries/health.queries';

	const health = createHealthQuery();

	const isHealthy = $derived(health.data?.status === 'ok');
</script>

<Card class="max-w-md">
	<CardHeader>
		<CardTitle>Status layanan</CardTitle>
		<CardDescription>Diperiksa dari API Go lewat BFF SvelteKit.</CardDescription>
	</CardHeader>
	<CardContent>
		{#if health.isPending}
			<p class="text-sm text-muted-foreground">Memeriksa…</p>
		{:else if health.error}
			<div class="space-y-3">
				<p class="text-sm text-destructive">{health.error.message}</p>
				<Button variant="outline" size="sm" onclick={() => void health.refetch()}>Coba lagi</Button>
			</div>
		{:else if health.data}
			<div class="flex items-center gap-3">
				<Badge variant={isHealthy ? 'default' : 'destructive'}>
					{isHealthy ? 'OK' : 'DEGRADED'}
				</Badge>
				<span class="text-sm text-muted-foreground">
					Basis data: {health.data.database}
				</span>
			</div>
		{/if}
	</CardContent>
</Card>
