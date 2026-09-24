<script lang="ts">
	import { LaporanHarian } from '$lib/domains/penjualan';

	// The screen owns its heading, its date field and its two queries; the route is
	// only the adapter that puts the domain component on a URL. It is Admin-only,
	// enforced in `hooks.server.ts` and again by the API.
</script>

<LaporanHarian />
