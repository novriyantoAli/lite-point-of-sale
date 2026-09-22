import { describe, expect, it } from 'vitest';
import {
	BackendSessionSchema,
	CreatePenggunaInputSchema,
	LoginInputSchema,
	PenggunaListSchema,
	PenggunaSchema,
	SessionSchema
} from './auth.schema';

const pengguna = { id: 1, username: 'kasir1', role: 'kasir', active: true };

describe('PenggunaSchema', () => {
	it('reads a Pengguna as the API answers it', () => {
		expect(PenggunaSchema.parse(pengguna)).toEqual(pengguna);
	});

	it('rejects a Peran the API never sends', () => {
		expect(() => PenggunaSchema.parse({ ...pengguna, role: 'pemilik' })).toThrow();
	});

	it('rejects a Pengguna that is missing its active state', () => {
		expect(() => PenggunaSchema.parse({ id: 1, username: 'kasir1', role: 'kasir' })).toThrow();
	});
});

describe('LoginInputSchema', () => {
	it('trims the username before it becomes a request', () => {
		expect(LoginInputSchema.parse({ username: '  admin ', password: 'rahasia' })).toEqual({
			username: 'admin',
			password: 'rahasia'
		});
	});

	it('rejects an empty username with a message fit for the form', () => {
		const result = LoginInputSchema.safeParse({ username: '   ', password: 'rahasia' });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Username wajib diisi.');
	});

	it('rejects an empty password', () => {
		expect(LoginInputSchema.safeParse({ username: 'admin', password: '' }).success).toBe(false);
	});
});

describe('CreatePenggunaInputSchema', () => {
	it('accepts a Kasir with a password of the minimum length', () => {
		const input = { username: 'kasir1', password: '12345678', role: 'kasir' as const };

		expect(CreatePenggunaInputSchema.parse(input)).toEqual(input);
	});

	it('rejects a password below the minimum, saying what the minimum is', () => {
		const result = CreatePenggunaInputSchema.safeParse({
			username: 'kasir1',
			password: 'pendek',
			role: 'kasir'
		});

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Password minimal 8 karakter.');
	});

	it('rejects a password longer than bcrypt accepts', () => {
		const result = CreatePenggunaInputSchema.safeParse({
			username: 'kasir1',
			password: 'a'.repeat(73),
			role: 'kasir'
		});

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Password maksimal 72 karakter.');
	});
});

describe('SessionSchema', () => {
	it('reads the Pengguna the BFF answers with', () => {
		expect(SessionSchema.parse({ data: { user: pengguna } }).data.user).toEqual(pengguna);
	});

	it('has no field for a token: the browser never receives one', () => {
		const parsed = SessionSchema.parse({ data: { token: 'rahasia', user: pengguna } });

		expect(parsed.data).toEqual({ user: pengguna });
		expect(JSON.stringify(parsed)).not.toContain('rahasia');
	});
});

describe('BackendSessionSchema', () => {
	it('is the only schema that describes the token, and only the BFF parses it', () => {
		const parsed = BackendSessionSchema.parse({ data: { token: 'rahasia', user: pengguna } });

		expect(parsed.data.token).toBe('rahasia');
	});

	it('rejects a Go answer without a token', () => {
		expect(() => BackendSessionSchema.parse({ data: { user: pengguna } })).toThrow();
	});
});

describe('PenggunaListSchema', () => {
	it('reads a list of Pengguna', () => {
		expect(PenggunaListSchema.parse({ data: [pengguna] }).data).toHaveLength(1);
	});

	it('reads an empty store as an empty list, not a missing field', () => {
		expect(PenggunaListSchema.parse({ data: [] }).data).toEqual([]);
	});
});
