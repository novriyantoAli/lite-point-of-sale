import { z } from 'zod';

/**
 * Contract of the auth domain, mirroring the Go DTOs of `adapter/httpapi` 1:1
 * (ADR-0006): the schema is the only thing allowed to cross the HTTP boundary,
 * and it is the single source of the types below it.
 */

/** The two Peran of CONTEXT.md. The values are what the API sends and stores. */
export const RoleSchema = z.enum(['kasir', 'admin']);
export type Role = z.infer<typeof RoleSchema>;

/** A Pengguna as the API ever answers it: never with a password or its hash. */
export const PenggunaSchema = z.object({
	id: z.number(),
	username: z.string(),
	role: RoleSchema,
	active: z.boolean()
});
export type Pengguna = z.infer<typeof PenggunaSchema>;

/**
 * Password bounds, kept in step with `usecase/auth` on the Go side: the lower
 * one is the rule, the upper one is bcrypt's ceiling.
 */
export const MIN_PASSWORD_LENGTH = 8;
export const MAX_PASSWORD_LENGTH = 72;

/**
 * bcrypt hashes at most 72 **bytes**, and Go compares `len()` — bytes — before
 * hashing. zod's `.max()` counts UTF-16 code units instead, so the check is done
 * on the encoded length: without this, a password of 40 accented characters is
 * 80 bytes, passes here, and comes back as a 400 from the API.
 */
function withinBcryptLimit(password: string): boolean {
	return new TextEncoder().encode(password).length <= MAX_PASSWORD_LENGTH;
}

export const LoginInputSchema = z.object({
	username: z.string().trim().min(1, 'Username wajib diisi.'),
	password: z.string().min(1, 'Password wajib diisi.')
});
export type LoginInput = z.infer<typeof LoginInputSchema>;

export const CreatePenggunaInputSchema = z.object({
	username: z.string().trim().min(1, 'Username wajib diisi.'),
	password: z
		.string()
		.min(MIN_PASSWORD_LENGTH, `Password minimal ${MIN_PASSWORD_LENGTH} karakter.`)
		.refine(withinBcryptLimit, { message: `Password maksimal ${MAX_PASSWORD_LENGTH} byte.` }),
	role: RoleSchema
});
export type CreatePenggunaInput = z.infer<typeof CreatePenggunaInputSchema>;

/** The body of a deactivation or a reactivation. */
export const PenggunaActiveInputSchema = z.object({ active: z.boolean() });
export type PenggunaActiveInput = z.infer<typeof PenggunaActiveInputSchema>;

/**
 * Every answer that carries one Pengguna under `data.user`: login, `/auth/me`,
 * and creating or reactivating a Pengguna. Note what is *not* here — the token.
 * It lives in an httpOnly cookie the browser cannot read, so no schema on this
 * side ever has a reason to describe it (ADR-0001, ADR-0006).
 */
export const PenggunaEnvelopeSchema = z.object({
	data: z.object({ user: PenggunaSchema })
});

/** What Go answers to a login. Only the BFF ever parses this one. */
export const BackendSessionSchema = z.object({
	data: z.object({ token: z.string(), user: PenggunaSchema })
});

/** What Go answers to a Pengguna list. */
export const PenggunaListSchema = z.object({
	data: z.array(PenggunaSchema)
});
