import { describe, expect, it } from 'vitest';
import { BackupEnvelopeSchema, BackupListSchema, BackupSchema } from './backup.schema';

const backup = {
	name: 'pos-20260115-120000-000000000.db',
	size: 8192,
	created_at: '2026-01-15T12:00:00Z'
};

describe('BackupSchema', () => {
	it('reads a snapshot as the API answers it', () => {
		expect(BackupSchema.parse(backup)).toEqual(backup);
	});

	it('rejects a size that is not a whole number of bytes', () => {
		expect(() => BackupSchema.parse({ ...backup, size: 8.5 })).toThrow();
	});

	it('rejects a negative size', () => {
		expect(() => BackupSchema.parse({ ...backup, size: -1 })).toThrow();
	});

	it('rejects a name that is not text', () => {
		expect(() => BackupSchema.parse({ ...backup, name: 42 })).toThrow();
	});
});

describe('BackupEnvelopeSchema', () => {
	it('reads the envelope of a manual export', () => {
		expect(BackupEnvelopeSchema.parse({ data: { backup } }).data.backup).toEqual(backup);
	});

	it('rejects an answer that carries no backup', () => {
		expect(() => BackupEnvelopeSchema.parse({ data: {} })).toThrow();
	});
});

describe('BackupListSchema', () => {
	it('reads the list the API answers', () => {
		expect(BackupListSchema.parse({ data: [backup] }).data).toEqual([backup]);
	});

	it('reads an empty list as no backups', () => {
		expect(BackupListSchema.parse({ data: [] }).data).toEqual([]);
	});

	it('rejects a list whose entry is not a snapshot', () => {
		expect(() => BackupListSchema.parse({ data: [{ name: 'x.db' }] })).toThrow();
	});
});
