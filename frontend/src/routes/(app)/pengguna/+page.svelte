<script lang="ts">
	import { PenggunaList } from '$lib/domains/auth';

	// `data` is the merged result of this page's loads and the (app) layout's:
	// the Pengguna comes from the layout, which got it from hooks.server.ts.
	// The screen owns its own heading and both of its modules; the route is only
	// the adapter that puts the domain component on a URL.
	let { data } = $props();
</script>

{#if data.user}
	<PenggunaList currentUserId={data.user.id} />
{/if}
