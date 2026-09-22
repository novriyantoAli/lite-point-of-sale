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
 * no domain reimplements error parsing and no component sees a raw axios
 * error. The 401 → `/login` redirect belongs here once the auth slice
 * (`domains/auth`) and its route exist.
 */
apiClient.interceptors.response.use(
	(response) => response,
	(error) => Promise.reject(normalizeApiError(error))
);
