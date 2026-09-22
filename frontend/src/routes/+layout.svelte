<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';

	let { children } = $props();

	// Created per instantiation (per request on the server), so state never
	// leaks between requests or between tests.
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: { staleTime: 30_000, retry: 1 }
		}
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<QueryClientProvider client={queryClient}>
	{@render children()}
</QueryClientProvider>
