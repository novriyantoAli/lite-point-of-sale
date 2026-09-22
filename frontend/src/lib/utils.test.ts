import { describe, expect, it } from 'vitest';
import { formatRupiah } from './utils';

describe('formatRupiah', () => {
	it('groups whole rupiah in thousands', () => {
		expect(formatRupiah(18000)).toBe('Rp 18.000');
		expect(formatRupiah(1234567)).toBe('Rp 1.234.567');
	});

	it('writes a price of 0 rather than leaving the amount blank', () => {
		expect(formatRupiah(0)).toBe('Rp 0');
	});

	it('drops the decimals money in this app does not have', () => {
		expect(formatRupiah(3000)).not.toContain(',');
	});
});
