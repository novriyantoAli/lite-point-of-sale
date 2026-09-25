<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { collectFieldErrors } from '$lib/utils';
	import { createLoginMutation } from '../queries/auth.queries';
	import { LoginInputSchema } from '../schemas/auth.schema';

	let username = $state('');
	let password = $state('');
	let fieldErrors = $state<Partial<Record<'username' | 'password', string>>>({});

	const login = createLoginMutation();

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};

		// Validated with the same schema that builds the request (§11, §6.8), so
		// a field error and a request error can never disagree.
		const parsed = LoginInputSchema.safeParse({ username, password });
		if (!parsed.success) {
			fieldErrors = collectFieldErrors(parsed.error, ['username', 'password'] as const);
			return;
		}

		try {
			await login.mutateAsync(parsed.data);
			// invalidateAll re-runs the server load, so the guard and the layout
			// see the session the BFF has just set.
			await goto(resolve('/'), { invalidateAll: true });
		} catch {
			// The mutation keeps the normalized AppError; it is rendered below.
		}
	}
</script>

<!--
  Kepala modul: latar isian, ditutup garis tinta — satu-satunya penanda kepala
  di dunia ini, tanpa bayangan.
-->
<div class="flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5">
	<h2 class="text-[13px] font-bold">Masuk</h2>
	<span class="text-xs">username &amp; password Pengguna</span>
</div>

<form class="space-y-2 px-2 py-2" onsubmit={submit} novalidate>
	<div class="space-y-1">
		<Label for="username" class="text-xs font-semibold">Username</Label>
		<Input
			id="username"
			name="username"
			autocomplete="username"
			class="h-[26px] px-1.5 text-[13px] md:text-[13px] shadow-none"
			bind:value={username}
			aria-invalid={fieldErrors.username ? true : undefined}
		/>
		{#if fieldErrors.username}
			<p class="text-xs font-semibold text-destructive">{fieldErrors.username}</p>
		{/if}
	</div>

	<div class="space-y-1">
		<Label for="password" class="text-xs font-semibold">Password</Label>
		<Input
			id="password"
			name="password"
			type="password"
			autocomplete="current-password"
			class="h-[26px] px-1.5 text-[13px] md:text-[13px] shadow-none"
			bind:value={password}
			aria-invalid={fieldErrors.password ? true : undefined}
		/>
		{#if fieldErrors.password}
			<p class="text-xs font-semibold text-destructive">{fieldErrors.password}</p>
		{/if}
	</div>

	{#if login.error}
		<p class="text-xs font-semibold text-destructive" role="alert">{login.error.message}</p>
	{/if}

	<!--
	  Aksi utama: satu-satunya bidang bertinta penuh. Saat mati ia kehilangan
	  tintanya (bukan dipudarkan), jadi putih bergaris putus-putus.
	-->
	<Button
		type="submit"
		class="h-10 w-full text-[15px] font-semibold hover:bg-primary disabled:border-border disabled:border-dashed disabled:bg-card disabled:text-foreground disabled:opacity-100"
		disabled={login.isPending}
	>
		{login.isPending ? 'Memeriksa…' : 'Masuk'}
	</Button>
</form>
