import { describe, expect, it } from 'vitest';
import { HealthSchema } from './health.schema';

describe('HealthSchema', () => {
	it('reads the status and database fields of a healthy API response', () => {
		const health = HealthSchema.parse(JSON.parse('{"status":"ok","database":"ok"}'));

		expect(health.status).toBe('ok');
		expect(health.database).toBe('ok');
	});

	it('accepts a degraded service with an unavailable database', () => {
		const health = HealthSchema.parse({ status: 'degraded', database: 'unavailable' });

		expect(health.status).toBe('degraded');
		expect(health.database).toBe('unavailable');
	});

	it('rejects a status the API never returns', () => {
		expect(() => HealthSchema.parse({ status: 'fine', database: 'ok' })).toThrow();
	});

	it('rejects a response that is missing the database field', () => {
		expect(() => HealthSchema.parse({ status: 'ok' })).toThrow();
	});
});
