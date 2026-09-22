import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from 'axios';
import { describe, expect, it } from 'vitest';
import { normalizeApiError } from './errors';

function responseError(status: number, data?: unknown): AxiosError {
	const config: InternalAxiosRequestConfig = { headers: new AxiosHeaders() };

	return new AxiosError('Request failed', AxiosError.ERR_BAD_RESPONSE, config, undefined, {
		data,
		status,
		statusText: '',
		headers: {},
		config
	});
}

function networkError(): AxiosError {
	const config: InternalAxiosRequestConfig = { headers: new AxiosHeaders() };

	return new AxiosError('Network Error', AxiosError.ERR_NETWORK, config);
}

describe('normalizeApiError', () => {
	it('keeps the message and code the API sent', () => {
		const error = responseError(503, {
			message: 'Basis data tidak tersedia.',
			error: 'database_unavailable'
		});

		expect(normalizeApiError(error)).toEqual({
			message: 'Basis data tidak tersedia.',
			status: 503,
			code: 'database_unavailable'
		});
	});

	it('falls back to the HTTP status when the body carries no message', () => {
		const error = responseError(500, '<html>Internal Server Error</html>');

		expect(normalizeApiError(error)).toMatchObject({
			message: 'Permintaan gagal (HTTP 500).',
			status: 500,
			code: AxiosError.ERR_BAD_RESPONSE
		});
	});

	it('reports an unreachable server when the request got no response', () => {
		const normalized = normalizeApiError(networkError());

		expect(normalized.message).toBe('Tidak dapat menghubungi server.');
		expect(normalized.status).toBeUndefined();
	});

	it('does not leak non-axios errors to the UI', () => {
		expect(normalizeApiError(new Error('boom'))).toEqual({
			message: 'Terjadi kesalahan tak terduga.',
			code: 'Error'
		});
		expect(normalizeApiError('boom')).toEqual({ message: 'Terjadi kesalahan tak terduga.' });
	});
});
