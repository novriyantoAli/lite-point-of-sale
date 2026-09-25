<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { createBackupListQuery, createBackupMutation } from '../queries/backup.queries';

	/**
	 * The Backup screen: the Admin's safety net (issue #10). It lists the
	 * snapshots the store has kept and offers a manual export now. The daily
	 * automatic backup runs on the server — this screen does not turn it on or
	 * off, it only takes the same snapshot the scheduler would.
	 */
	const backups = createBackupListQuery();
	const exportBackup = createBackupMutation();

	let notice = $state('');

	async function backupNow() {
		notice = '';
		try {
			const created = await exportBackup.mutateAsync();
			notice = `Backup dibuat: ${created.name}`;
		} catch {
			// exportBackup.error carries the normalized message, rendered below.
		}
	}

	/**
	 * The shared column style of a row: name, size, and time read alike because
	 * they are the three facts every snapshot has.
	 */
	const metaClass = 'text-sm text-muted-foreground';

	/** A snapshot length as a person reads it, not a raw byte count. */
	function formatBytes(bytes: number): string {
		if (bytes < 1024) {
			return `${bytes} B`;
		}
		if (bytes < 1024 * 1024) {
			return `${(bytes / 1024).toFixed(1)} KB`;
		}
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	/** The API's RFC 3339 timestamp as the store's local clock writes it. */
	function formatTime(createdAt: string): string {
		const date = new Date(createdAt);
		if (Number.isNaN(date.getTime())) {
			return createdAt;
		}

		return date.toLocaleString('id-ID');
	}
</script>

<div class="space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Backup</h1>
		<p class="text-sm text-muted-foreground">
			Salinan database toko. Backup harian berjalan otomatis; tombol di bawah mengambil salinan
			manual saat ini juga.
		</p>
	</div>

	<div class="flex gap-2">
		<Button onclick={backupNow} disabled={exportBackup.isPending}>
			{exportBackup.isPending ? 'Membuat backup…' : 'Backup sekarang'}
		</Button>
	</div>

	{#if exportBackup.error}
		<p class="text-sm text-destructive" role="alert">{exportBackup.error.message}</p>
	{/if}
	{#if notice}
		<p class="text-sm text-muted-foreground" role="status">{notice}</p>
	{/if}

	{#if backups.isPending}
		<p class="text-sm text-muted-foreground">Memuat daftar backup…</p>
	{:else if backups.error}
		<div class="space-y-3">
			<p class="text-sm text-destructive">{backups.error.message}</p>
			<Button variant="outline" size="sm" onclick={() => void backups.refetch()}>
				Coba lagi
			</Button>
		</div>
	{:else if backups.data?.length === 0}
		<p class="text-sm text-muted-foreground">
			Belum ada backup. Tekan “Backup sekarang” untuk membuat yang pertama.
		</p>
	{:else}
		<ul class="divide-y rounded-lg border">
			{#each backups.data ?? [] as backup (backup.name)}
				<li class="flex items-center justify-between gap-4 p-4">
					<div class="min-w-0">
						<span class="block truncate font-medium">{backup.name}</span>
						<span class={metaClass}>
							{formatTime(backup.created_at)} · {formatBytes(backup.size)}
						</span>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>
