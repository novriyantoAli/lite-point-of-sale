<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import type { Role } from '../schemas/auth.schema';

	/** The Peran as it is written for a person, in one place. */
	const labels: Record<Role, string> = { kasir: 'Kasir', admin: 'Admin' };

	let { role }: { role: Role } = $props();
</script>

<!--
  Tag tenang dunia baru: tinggi 15px, huruf 12px/600, isian Wash Grey dengan
  garis rambut. Peran adalah keadaan yang tenang — merah hanya milik tab dan
  harga, jadi ia tidak pernah menandai Peran (DESIGN.md).
-->
<Badge variant="secondary" class="h-[15px] border-border px-[5px] py-0 text-xs font-semibold">
	{labels[role]}
</Badge>
