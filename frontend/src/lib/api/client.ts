import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { resolve } from '$app/paths';
import axios from 'axios';
import { normalizeApiError } from './errors';

/**
 * The single axios instance of the app. `baseURL: '/api'` is same-origin, so
 * the browser only ever talks to the SvelteKit BFF — never to Go directly
 * (ADR-0001, ADR-0006). Domain `api/` layers are the only callers.
 */
export const apiClient = axios.create({
	baseURL: '/api',
	timeout: 15_000
});

/**
 * Every failure is normalized in this one place into `AppError` (§9, §11), so
 * no domain reimplements error parsing and no component sees a raw axios error.
 */
apiClient.interceptors.response.use(
	(response) => response,
	(error) => {
		const normalized = normalizeApiError(error);

		// A session that ended is not something a component can recover from, so
		// the terminal goes back to the login page from here. The code decides:
		// `invalid_token` is a dead session, while `invalid_credentials` is a
		// wrong password and has to stay on the form that asked for it.
		if (browser && normalized.status === 401 && normalized.code === 'invalid_token') {
			void goto(resolve('/login'));
		}

		return Promise.reject(normalized);
	}
);
