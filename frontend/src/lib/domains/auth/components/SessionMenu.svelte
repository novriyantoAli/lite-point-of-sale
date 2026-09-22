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

<div class="flex items-center gap-2">
	<span class="text-sm text-muted-foreground">{user.username}</span>
	<RoleBadge role={user.role} />
	<Button variant="outline" size="sm" disabled={logout.isPending} onclick={() => void signOut()}>
		Keluar
	</Button>
</div>
