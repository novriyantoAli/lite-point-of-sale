<script lang="ts">
	import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';

	let { children } = $props();

	// Fresh QueryClient per render so cached server state never leaks between
	// component tests (ADR-0007). Test-only: nothing in the app imports it.
	//
	// `retryDelay: 0` keeps a query that carries its own retry policy from sitting
	// on the production backoff; the default `retry: false` still applies to every
	// other query.
	const queryClient = new QueryClient({
		defaultOptions: { queries: { retry: false, retryDelay: 0 } }
	});
</script>

<QueryClientProvider client={queryClient}>
	{@render children()}
</QueryClientProvider>
