import { describe, expect, it } from 'vitest';
import {
	PengaturanEnvelopeSchema,
	PengaturanSchema,
	UpdatePengaturanInputSchema
} from './pengaturan.schema';

const pengaturan = {
	id: 1,
	header: 'Toko Kopi',
	footer: 'Terima kasih',
	paper_width: 80,
	low_stock_threshold: 5
};

describe('PengaturanSchema', () => {
	it('reads the Pengaturan as the API answers it', () => {
		expect(PengaturanSchema.parse(pengaturan)).toEqual(pengaturan);
	});

	it('reads an empty template as the seeded defaults', () => {
		const parsed = PengaturanSchema.parse({ ...pengaturan, header: '', footer: '' });

		expect(parsed.header).toBe('');
		expect(parsed.footer).toBe('');
	});

	it('rejects a paper width that is not a whole millimetre count', () => {
		expect(() => PengaturanSchema.parse({ ...pengaturan, paper_width: 80.5 })).toThrow();
	});
});

describe('UpdatePengaturanInputSchema', () => {
	it('turns the form strings into the integers the API stores', () => {
		const parsed = UpdatePengaturanInputSchema.parse({
			header: 'Toko Kopi\nJl. Melati 1',
			footer: 'Terima kasih',
			paper_width: '58',
			low_stock_threshold: '3'
		});

		expect(parsed).toEqual({
			header: 'Toko Kopi\nJl. Melati 1',
			footer: 'Terima kasih',
			paper_width: 58,
			low_stock_threshold: 3
		});
	});

	it('keeps a multi-line header block as typed, without flattening it', () => {
		const parsed = UpdatePengaturanInputSchema.parse({
			header: '  Toko Kopi\n\nJl. Melati 1  ',
			footer: '',
			paper_width: '80',
			low_stock_threshold: '5'
		});

		expect(parsed.header).toBe('  Toko Kopi\n\nJl. Melati 1  ');
	});

	it('rejects a paper width other than 58 or 80', () => {
		expect(() =>
			UpdatePengaturanInputSchema.parse({
				header: '',
				footer: '',
				paper_width: '60',
				low_stock_threshold: '5'
			})
		).toThrow();
	});

	it('rejects an ambang of zero or less', () => {
		for (const threshold of ['0', '-5']) {
			expect(() =>
				UpdatePengaturanInputSchema.parse({
					header: '',
					footer: '',
					paper_width: '80',
					low_stock_threshold: threshold
				})
			).toThrow();
		}
	});

	it('rejects an ambang that is not a whole number', () => {
		expect(() =>
			UpdatePengaturanInputSchema.parse({
				header: '',
				footer: '',
				paper_width: '80',
				low_stock_threshold: '18.000'
			})
		).toThrow();
	});
});

describe('PengaturanEnvelopeSchema', () => {
	it('reads the envelope the API wraps the Pengaturan in', () => {
		expect(
			PengaturanEnvelopeSchema.parse({ data: { settings: pengaturan } }).data.settings
		).toEqual(pengaturan);
	});

	it('rejects an answer that carries no Pengaturan', () => {
		expect(() => PengaturanEnvelopeSchema.parse({ data: {} })).toThrow();
	});
});
