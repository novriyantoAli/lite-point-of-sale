<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import type { Role } from '../schemas/auth.schema';

	/** The Peran as it is written for a person, in one place. */
	const labels: Record<Role, string> = { kasir: 'Kasir', admin: 'Admin' };

	let { role }: { role: Role } = $props();
</script>

<Badge variant="secondary">{labels[role]}</Badge>
