<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
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

<Card class="w-full max-w-sm">
	<CardHeader>
		<CardTitle>Masuk</CardTitle>
		<CardDescription>Gunakan username dan password Pengguna.</CardDescription>
	</CardHeader>
	<CardContent>
		<form class="space-y-4" onsubmit={submit} novalidate>
			<div class="space-y-2">
				<Label for="username">Username</Label>
				<Input
					id="username"
					name="username"
					autocomplete="username"
					bind:value={username}
					aria-invalid={fieldErrors.username ? true : undefined}
				/>
				{#if fieldErrors.username}
					<p class="text-sm text-destructive">{fieldErrors.username}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="password">Password</Label>
				<Input
					id="password"
					name="password"
					type="password"
					autocomplete="current-password"
					bind:value={password}
					aria-invalid={fieldErrors.password ? true : undefined}
				/>
				{#if fieldErrors.password}
					<p class="text-sm text-destructive">{fieldErrors.password}</p>
				{/if}
			</div>

			{#if login.error}
				<p class="text-sm text-destructive" role="alert">{login.error.message}</p>
			{/if}

			<Button type="submit" class="w-full" disabled={login.isPending}>
				{login.isPending ? 'Memeriksa…' : 'Masuk'}
			</Button>
		</form>
	</CardContent>
</Card>
