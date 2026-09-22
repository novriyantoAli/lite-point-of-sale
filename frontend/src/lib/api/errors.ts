import { isAxiosError } from 'axios';

/**
 * The one error shape of the app (ADR-0006 conventions): `api/` layers and
 * components read `message`, and branch on `status`/`code` when they must.
 * Raw axios/Error values never leave the HTTP boundary.
 */
export interface AppError {
	message: string;
	status?: number;
	code?: string;
}

const UNREACHABLE_MESSAGE = 'Tidak dapat menghubungi server.';
const UNEXPECTED_MESSAGE = 'Terjadi kesalahan tak terduga.';

/** The Go API answers failures with `{ message, error }`. */
interface ErrorBody {
	message?: unknown;
	error?: unknown;
}

export function normalizeApiError(error: unknown): AppError {
	if (isAxiosError(error)) {
		const status = error.response?.status;
		const body = error.response?.data as ErrorBody | undefined;

		return {
			message: pickMessage(body, status),
			status,
			code: pickCode(body, error.code)
		};
	}

	if (error instanceof Error) {
		return { message: UNEXPECTED_MESSAGE, code: error.name };
	}

	return { message: UNEXPECTED_MESSAGE };
}

function pickMessage(body: ErrorBody | undefined, status: number | undefined): string {
	if (typeof body?.message === 'string' && body.message !== '') {
		return body.message;
	}
	if (status !== undefined) {
		return `Permintaan gagal (HTTP ${status}).`;
	}
	return UNREACHABLE_MESSAGE;
}

function pickCode(body: ErrorBody | undefined, axiosCode: string | undefined): string | undefined {
	if (typeof body?.error === 'string' && body.error !== '') {
		return body.error;
	}
	return axiosCode;
}
