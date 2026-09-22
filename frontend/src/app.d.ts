// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { Pengguna } from '$lib/domains/auth';

declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			/**
			 * The Pengguna behind the session cookie, resolved on every page
			 * navigation in `hooks.server.ts`. Null when nobody is logged in.
			 */
			user: Pengguna | null;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
