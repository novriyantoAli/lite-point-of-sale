<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
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
		} finally {
			// Even a failed logout leaves the terminal on the login page: the
			// cookie is cleared there, and a Kasir handing the terminal over must
			// not be left looking at an open till.
			await goto(resolve('/login'), { invalidateAll: true });
		}
	}
</script>

<div class="flex items-center gap-2">
	<span class="text-sm text-muted-foreground">{user.username}</span>
	<Badge variant="secondary">{user.role === 'admin' ? 'Admin' : 'Kasir'}</Badge>
	<Button variant="outline" size="sm" disabled={logout.isPending} onclick={() => void signOut()}>
		Keluar
	</Button>
</div>
