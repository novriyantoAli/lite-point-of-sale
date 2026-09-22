<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import {
		createPenggunaListQuery,
		createPenggunaMutation,
		createSetPenggunaActiveMutation
	} from '../queries/auth.queries';
	import { CreatePenggunaInputSchema, type Role } from '../schemas/auth.schema';

	/**
	 * `currentUserId` is the Pengguna using the page. It comes from the layout's
	 * server load, so the row of the Admin themselves can be shown as untouchable
	 * instead of offering a button the API would refuse.
	 */
	let { currentUserId }: { currentUserId: number } = $props();

	const pengguna = createPenggunaListQuery();
	const create = createPenggunaMutation();
	const setActive = createSetPenggunaActiveMutation();

	let username = $state('');
	let password = $state('');
	let role = $state<Role>('kasir');
	let fieldErrors = $state<{ username?: string; password?: string }>({});
	let notice = $state('');

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};
		notice = '';

		// The same schema the api layer parses with, so a rule is written once.
		const parsed = CreatePenggunaInputSchema.safeParse({ username, password, role });
		if (!parsed.success) {
			for (const issue of parsed.error.issues) {
				const field = issue.path[0];
				if (field === 'username' || field === 'password') {
					fieldErrors[field] ??= issue.message;
				}
			}
			return;
		}

		try {
			const created = await create.mutateAsync(parsed.data);
			notice = `Pengguna ${created.username} ditambahkan.`;
			username = '';
			password = '';
			role = 'kasir';
		} catch {
			// create.error carries the normalized message, rendered below.
		}
	}
</script>

<div class="space-y-6">
	<form class="space-y-4 rounded-lg border p-4" onsubmit={submit} novalidate>
		<h2 class="font-medium">Tambah Pengguna</h2>

		<div class="grid gap-4 sm:grid-cols-3">
			<div class="space-y-2">
				<Label for="pengguna-username">Username</Label>
				<Input
					id="pengguna-username"
					name="username"
					autocomplete="off"
					bind:value={username}
					aria-invalid={fieldErrors.username ? true : undefined}
				/>
				{#if fieldErrors.username}
					<p class="text-sm text-destructive">{fieldErrors.username}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="pengguna-password">Password</Label>
				<Input
					id="pengguna-password"
					name="password"
					type="password"
					autocomplete="new-password"
					bind:value={password}
					aria-invalid={fieldErrors.password ? true : undefined}
				/>
				{#if fieldErrors.password}
					<p class="text-sm text-destructive">{fieldErrors.password}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="pengguna-role">Peran</Label>
				<Select.Root type="single" bind:value={role}>
					<Select.Trigger id="pengguna-role" class="w-full">
						<Select.Value placeholder="Pilih peran" />
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="kasir" label="Kasir">Kasir</Select.Item>
						<Select.Item value="admin" label="Admin">Admin</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>
		</div>

		{#if create.error}
			<p class="text-sm text-destructive" role="alert">{create.error.message}</p>
		{/if}
		{#if notice}
			<p class="text-sm text-muted-foreground" role="status">{notice}</p>
		{/if}

		<Button type="submit" disabled={create.isPending}>
			{create.isPending ? 'Menyimpan…' : 'Tambah'}
		</Button>
	</form>

	{#if pengguna.isPending}
		<p class="text-sm text-muted-foreground">Memuat daftar Pengguna…</p>
	{:else if pengguna.error}
		<div class="space-y-3">
			<p class="text-sm text-destructive">{pengguna.error.message}</p>
			<Button variant="outline" size="sm" onclick={() => void pengguna.refetch()}>Coba lagi</Button>
		</div>
	{:else if pengguna.data?.length === 0}
		<p class="text-sm text-muted-foreground">
			Belum ada Pengguna lain. Tambahkan Kasir lewat formulir di atas.
		</p>
	{:else}
		<ul class="divide-y rounded-lg border">
			{#each pengguna.data ?? [] as user (user.id)}
				<li class="flex items-center justify-between gap-4 p-4">
					<div class="flex items-center gap-2">
						<span class="font-medium">{user.username}</span>
						<Badge variant="secondary">{user.role === 'admin' ? 'Admin' : 'Kasir'}</Badge>
						{#if !user.active}
							<Badge variant="outline">Nonaktif</Badge>
						{/if}
						{#if user.id === currentUserId}
							<span class="text-sm text-muted-foreground">(Anda)</span>
						{/if}
					</div>

					{#if user.id === currentUserId}
						<span class="text-sm text-muted-foreground">Tidak bisa menonaktifkan diri sendiri</span>
					{:else}
						<Button
							variant="outline"
							size="sm"
							disabled={setActive.isPending}
							onclick={() => setActive.mutate({ id: user.id, active: !user.active })}
						>
							{user.active ? 'Nonaktifkan' : 'Aktifkan'}
						</Button>
					{/if}
				</li>
			{/each}
		</ul>

		{#if setActive.error}
			<p class="text-sm text-destructive" role="alert">{setActive.error.message}</p>
		{/if}
	{/if}
</div>
