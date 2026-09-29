<script lang="ts" module>
	import type { Role } from '../schemas/auth.schema';

	/**
	 * The Peran as it is written for a person, in one place. This tag draws it, and
	 * the Peran picker of the Pengguna screen writes it into its trigger. Two things
	 * make that a rule rather than a preference: the raw value (`kasir`, `admin`) is
	 * what the API sends and stores, so it must never reach a screen (ADR-0012), and
	 * a word that names a state should be written where the state is shown, not
	 * remembered somewhere else (DESIGN.md, Do's: "write every state's word next to
	 * its mark").
	 */
	export const PERAN_LABEL: Record<Role, string> = { kasir: 'Kasir', admin: 'Admin' };
</script>

<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';

	let { role }: { role: Role } = $props();
</script>

<!--
  Tag tenang dunia baru: tinggi 15px, huruf 12px/600, isian Wash Grey dengan
  garis rambut. Peran adalah keadaan yang tenang — merah hanya milik tab dan
  harga, jadi ia tidak pernah menandai Peran (DESIGN.md).
-->
<Badge variant="secondary" class="h-[15px] border-border px-[5px] py-0 text-xs font-semibold">
	{PERAN_LABEL[role]}
</Badge>
