<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import RoleBadge from './RoleBadge.svelte';
	import { createLogoutMutation } from '../queries/auth.queries';
	import type { Pengguna } from '../schemas/auth.schema';

	/**
	 * The Pengguna comes from the layout's server load, not from a query of its
	 * own: who is logged in is answered on the server on every navigation, and a
	 * second copy in the client cache would be duplicated server state.
	 */
	let { user }: { user: Pengguna } = $props();

	const logout = createLogoutMutation();

	async function signOut() {
		try {
			await logout.mutateAsync();
		} catch {
			// A failed logout means the BFF never saw the request, so the cookie is
			// still there and the guard sends this navigation straight back to the
			// dashboard — which is the truth: the session is still open. What must
			// not happen is an unhandled rejection.
		}

		await goto(resolve('/login'), { invalidateAll: true });
	}
</script>

<!--
  Siapa yang sedang memakai terminal, di dalam sel rel navigasi yang disediakan
  layout. Namanya memakai tinta hitam seperti seluruh teks lain di dunia ini —
  abu-abu hanya boleh jadi garis dan isian (DESIGN.md, Zero-Grey Rule).
-->
<div class="flex items-center gap-2">
	<span class="text-[13px]">{user.username}</span>
	<RoleBadge role={user.role} />

	<!--
	  Tombol petak berbingkai rambut. Saat menunggu ia tidak dipudarkan:
	  opasitas 0,5 milik dunia lama dan terukur di bawah AA, jadi tintanya yang
	  dipertahankan dan `cursor: not-allowed` yang menandai keadaannya.
	-->
	<Button
		variant="ghost"
		size="sm"
		class="h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:opacity-100"
		disabled={logout.isPending}
		onclick={() => void signOut()}
	>
		Keluar
	</Button>
</div>
