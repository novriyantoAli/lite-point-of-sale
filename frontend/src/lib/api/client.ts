import axios from 'axios';

/**
 * The single axios instance of the app. `baseURL: '/api'` is same-origin, so
 * the browser only ever talks to the SvelteKit BFF — never to Go directly
 * (ADR-0001, ADR-0006). Domain `api/` layers are the only callers.
 */
export const apiClient = axios.create({
	baseURL: '/api',
	timeout: 15_000
});
