// Public surface of the auth domain. `api/` and `queries/` stay internal —
// other domains and routes import from here only (ADR-0006).
export { default as LoginForm } from './components/LoginForm.svelte';
export { default as PenggunaList } from './components/PenggunaList.svelte';
export { default as SessionMenu } from './components/SessionMenu.svelte';

export {
	authKeys,
	createLoginMutation,
	createLogoutMutation,
	createPenggunaListQuery,
	createPenggunaMutation,
	createSetPenggunaActiveMutation
} from './queries/auth.queries';

export {
	BackendSessionSchema,
	CreatePenggunaInputSchema,
	LoginInputSchema,
	PenggunaListSchema,
	PenggunaSchema,
	RoleSchema,
	SessionSchema,
	MAX_PASSWORD_LENGTH,
	MIN_PASSWORD_LENGTH,
	type CreatePenggunaInput,
	type LoginInput,
	type Pengguna,
	type Role
} from './schemas/auth.schema';
