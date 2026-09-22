import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { authApi } from '../api/auth.api';
import type { CreatePenggunaInput, LoginInput, Pengguna } from '../schemas/auth.schema';

export const authKeys = {
	all: ['auth'] as const,
	pengguna: () => [...authKeys.all, 'pengguna'] as const
};

/**
 * There is deliberately no "current user" query. Who is logged in is resolved
 * on the server on every navigation (`hooks.server.ts` → Go) and handed to the
 * layout as load data, so a second copy in the client cache would be exactly
 * the duplicated server state the layering forbids (ADR-0006).
 */

export function createLoginMutation() {
	return createMutation<Pengguna, AppError, LoginInput>(() => ({
		mutationFn: (input: LoginInput) => authApi.login(input)
	}));
}

/**
 * Logging out is the BFF dropping its httpOnly cookie: the Go token carries no
 * server-side state, so there is nothing to invalidate in a query cache. The
 * caller navigates, and the guard re-reads the session on the way.
 */
export function createLogoutMutation() {
	return createMutation<void, AppError, void>(() => ({
		mutationFn: () => authApi.logout()
	}));
}

/**
 * The staff list of the store. `enabled: browser` for the same reason as the
 * health query: it is a same-origin call to the BFF, so it can only be made
 * where the BFF is reachable, and the page does not need SSR.
 */
export function createPenggunaListQuery() {
	return createQuery<Pengguna[], AppError>(() => ({
		queryKey: authKeys.pengguna(),
		queryFn: () => authApi.listPengguna(),
		enabled: browser
	}));
}

export function createPenggunaMutation() {
	const queryClient = useQueryClient();

	return createMutation<Pengguna, AppError, CreatePenggunaInput>(() => ({
		mutationFn: (input: CreatePenggunaInput) => authApi.createPengguna(input),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: authKeys.pengguna() })
	}));
}

export function createSetPenggunaActiveMutation() {
	const queryClient = useQueryClient();

	return createMutation<Pengguna, AppError, { id: number; active: boolean }>(() => ({
		mutationFn: ({ id, active }) => authApi.setPenggunaActive(id, active),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: authKeys.pengguna() })
	}));
}
