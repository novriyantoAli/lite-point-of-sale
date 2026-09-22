<script lang="ts">
	import { PenggunaList } from '$lib/domains/auth';

	// `data` is the merged result of this page's loads and the (app) layout's:
	// the Pengguna comes from the layout, which got it from hooks.server.ts.
	let { data } = $props();
</script>

<div class="space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Pengguna</h1>
		<p class="text-sm text-muted-foreground">
			Kasir dan Admin yang boleh memakai kasir ini. Akun yang Nonaktif tetap tersimpan riwayatnya,
			tetapi tidak bisa login lagi.
		</p>
	</div>

	{#if data.user}
		<PenggunaList currentUserId={data.user.id} />
	{/if}
</div>
