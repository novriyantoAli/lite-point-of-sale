import { describe, expect, expectTypeOf, it } from 'vitest';
import {
	CetakEnvelopeSchema,
	CheckoutEnvelopeSchema,
	CheckoutInputSchema,
	CheckoutItemSchema,
	HasilCetakSchema,
	JumlahBayarSchema,
	METODE_LABEL,
	NomorStrukSchema,
	OmzetHarianEnvelopeSchema,
	PenjualanEnvelopeSchema,
	PenjualanListSchema,
	PenjualanSchema,
	TanggalLaporanSchema,
	punyaKembalian,
	tanggalHariIni,
	type NomorStruk
} from './penjualan.schema';

// Compile-time: the schema's own `unknown` plumbing must not leak out — the
// lookup keys a query and builds a URL on a plain number.
expectTypeOf<NomorStruk>().toEqualTypeOf<number>();

const sale = {
	receipt_number: 1,
	created_at: '2026-09-23 10:00:00',
	cashier_id: 1,
	cashier_name: 'kasir1',
	total: 36000,
	items: [{ product_id: 1, name: 'Kopi Susu', price: 18000, quantity: 2, subtotal: 36000 }],
	payment: { method: 'cash', amount: 50000, change: 14000 }
};

describe('PenjualanSchema', () => {
	it('reads a Penjualan as the API answers it', () => {
		expect(PenjualanSchema.parse(sale)).toEqual(sale);
	});

	it('reads a non-tunai Pembayaran, so a stored Penjualan stays readable', () => {
		const parsed = PenjualanSchema.parse({
			...sale,
			payment: { method: 'qris', amount: 36000, change: 0 }
		});

		expect(parsed.payment.method).toBe('qris');
	});

	it('rejects a method the API does not know', () => {
		expect(() =>
			PenjualanSchema.parse({ ...sale, payment: { ...sale.payment, method: 'bitcoin' } })
		).toThrow();
	});

	it('rejects a Penjualan missing its Nomor Struk', () => {
		// The Nomor Struk is what identifies the sale for a reprint, so an answer
		// without it has to fail loudly rather than be read as 0.
		const withoutReceipt: Partial<typeof sale> = { ...sale };
		delete withoutReceipt.receipt_number;

		expect(() => PenjualanSchema.parse(withoutReceipt)).toThrow();
	});

	it('rejects a fractional total: money is whole rupiah in this app', () => {
		expect(() => PenjualanSchema.parse({ ...sale, total: 36000.5 })).toThrow();
	});
});

describe('PenjualanEnvelopeSchema', () => {
	it('reads the envelope the API answers with', () => {
		expect(PenjualanEnvelopeSchema.parse({ data: { sale } }).data.sale).toEqual(sale);
	});

	it('reads the stored sale a lookup answers with, non-tunai Pembayaran and all', () => {
		const qris = { ...sale, payment: { method: 'qris', amount: 36000, change: 0 } };

		const parsed = PenjualanEnvelopeSchema.parse({ data: { sale: qris } }).data.sale;

		expect(parsed.receipt_number).toBe(sale.receipt_number);
		expect(parsed.items).toEqual(sale.items);
		expect(parsed.payment.method).toBe('qris');
	});
});

describe('HasilCetakSchema', () => {
	it('reads a Struk that printed, with no message', () => {
		expect(HasilCetakSchema.parse({ printed: true })).toEqual({ printed: true });
	});

	it('reads a Struk that did not print, with the reason a Kasir can read', () => {
		expect(HasilCetakSchema.parse({ printed: false, message: 'Printer belum diatur.' })).toEqual({
			printed: false,
			message: 'Printer belum diatur.'
		});
	});

	it('refuses an answer that does not say whether the Struk printed', () => {
		expect(() => HasilCetakSchema.parse({ message: 'entah' })).toThrow();
	});
});

describe('CheckoutEnvelopeSchema', () => {
	it('reads the sale a checkout stored together with the result of its print', () => {
		const parsed = CheckoutEnvelopeSchema.parse({
			data: { sale, print: { printed: false, message: 'Printer belum diatur.' } }
		});

		expect(parsed.data.sale).toEqual(sale);
		expect(parsed.data.print).toEqual({ printed: false, message: 'Printer belum diatur.' });
	});

	it('refuses a checkout answer with no print result', () => {
		// A checkout always prints, so an answer that does not say how that went
		// would hide a failed print from the Kasir (ADR-0017, keputusan 1).
		expect(() => CheckoutEnvelopeSchema.parse({ data: { sale } })).toThrow();
	});
});

describe('CetakEnvelopeSchema', () => {
	it('reads the outcome of a reprint', () => {
		expect(CetakEnvelopeSchema.parse({ data: { print: { printed: true } } }).data.print).toEqual({
			printed: true
		});
	});
});

describe('NomorStrukSchema', () => {
	it('reads the number a person types into the lookup field', () => {
		expect(NomorStrukSchema.parse('7')).toBe(7);
		expect(NomorStrukSchema.parse(7)).toBe(7);
	});

	it('refuses zero and a negative number: a Nomor Struk starts at one', () => {
		expect(() => NomorStrukSchema.parse('0')).toThrow();
		expect(() => NomorStrukSchema.parse('-1')).toThrow();
	});

	it('refuses a blank field rather than reading it as 0', () => {
		expect(() => NomorStrukSchema.parse('')).toThrow();
	});

	it('says a blank field is required, not that it is a malformed number', () => {
		const parsed = NomorStrukSchema.safeParse('  ');

		expect(parsed.success).toBe(false);
		expect(parsed.error?.issues[0]?.message).toBe('Nomor Struk wajib diisi.');
	});

	it('refuses what is not a number at all, so the field need not ask the API', () => {
		expect(() => NomorStrukSchema.parse('abc')).toThrow();
	});

	it('refuses a thousand separator instead of reading 1.000 as 1', () => {
		expect(() => NomorStrukSchema.parse('1.000')).toThrow();
	});
});

describe('CheckoutItemSchema', () => {
	it('reads a Produk id and a whole quantity', () => {
		expect(CheckoutItemSchema.parse({ product_id: 1, quantity: 3 })).toEqual({
			product_id: 1,
			quantity: 3
		});
	});

	it('parses the quantity a form submits as text', () => {
		expect(CheckoutItemSchema.parse({ product_id: 1, quantity: '3' }).quantity).toBe(3);
	});

	it('refuses a quantity of zero or less: a line of zero is not an Item', () => {
		expect(() => CheckoutItemSchema.parse({ product_id: 1, quantity: 0 })).toThrow();
		expect(() => CheckoutItemSchema.parse({ product_id: 1, quantity: -2 })).toThrow();
	});

	it('refuses a thousand separator instead of reading 2.000 as 2', () => {
		// The same rule the Harga field carries: guessing wrong by 1000× is worse
		// than asking again.
		expect(() => CheckoutItemSchema.parse({ product_id: 1, quantity: '2.000' })).toThrow();
	});
});

describe('JumlahBayarSchema', () => {
	it('parses the text a form submits into whole rupiah', () => {
		expect(JumlahBayarSchema.parse('50000')).toBe(50000);
	});

	it('accepts a payment of exactly zero, which the API is the one to refuse against the total', () => {
		expect(JumlahBayarSchema.parse(0)).toBe(0);
	});

	it('refuses a negative payment', () => {
		expect(() => JumlahBayarSchema.parse('-1')).toThrow();
	});

	it('refuses a blank field rather than reading it as 0', () => {
		expect(() => JumlahBayarSchema.parse('')).toThrow();
	});

	it('refuses a thousand separator instead of reading 50.000 as 50', () => {
		expect(() => JumlahBayarSchema.parse('50.000')).toThrow();
	});
});

describe('CheckoutInputSchema', () => {
	it('reads the body of a checkout, parsing the amount the form submitted', () => {
		const parsed = CheckoutInputSchema.parse({
			items: [{ product_id: 1, quantity: 2 }],
			payment: { method: 'cash', amount: '50000' }
		});

		expect(parsed).toEqual({
			items: [{ product_id: 1, quantity: 2 }],
			payment: { method: 'cash', amount: 50000 }
		});
	});

	it('refuses an empty keranjang', () => {
		expect(() =>
			CheckoutInputSchema.parse({ items: [], payment: { method: 'cash', amount: 10000 } })
		).toThrow();
	});

	it('accepts the three recorded methods, so the till can offer them', () => {
		// No gateway is involved: the method and the nominal are all the till sends
		// for a non-tunai Pembayaran (CONTEXT.md, Pembayaran).
		for (const method of ['qris', 'debit', 'transfer'] as const) {
			const parsed = CheckoutInputSchema.parse({
				items: [{ product_id: 1, quantity: 1 }],
				payment: { method, amount: 10000 }
			});

			expect(parsed.payment.method).toBe(method);
		}
	});

	it('refuses a method the API does not know', () => {
		expect(() =>
			CheckoutInputSchema.parse({
				items: [{ product_id: 1, quantity: 1 }],
				payment: { method: 'bitcoin', amount: 10000 }
			})
		).toThrow();
	});
});

describe('METODE_LABEL', () => {
	it('writes Tunai for the cash method rather than the English word', () => {
		expect(METODE_LABEL.cash).toBe('Tunai');
		expect(METODE_LABEL.qris).toBe('QRIS');
	});
});

describe('punyaKembalian', () => {
	it('is Tunai alone, so the form and the Struk agree on what shows a Kembalian', () => {
		expect(punyaKembalian('cash')).toBe(true);
		for (const method of ['qris', 'debit', 'transfer'] as const) {
			expect(punyaKembalian(method)).toBe(false);
		}
	});
});

describe('PenjualanListSchema', () => {
	const row = {
		receipt_number: 7,
		created_at: '2026-09-23 10:00:00',
		cashier_id: 1,
		cashier_name: 'kasir1',
		total: 36000,
		method: 'cash'
	};

	it('reads a sales list as the API answers it', () => {
		expect(PenjualanListSchema.parse({ data: [row] })).toEqual({ data: [row] });
	});

	it('rejects a row whose method the API does not know', () => {
		expect(() => PenjualanListSchema.parse({ data: [{ ...row, method: 'bitcoin' }] })).toThrow();
	});

	it('rejects a row missing its Nomor Struk', () => {
		const withoutReceipt: Partial<typeof row> = { ...row };
		delete withoutReceipt.receipt_number;

		expect(() => PenjualanListSchema.parse({ data: [withoutReceipt] })).toThrow();
	});
});

describe('OmzetHarianEnvelopeSchema', () => {
	const omzet = {
		date: '2026-09-23',
		total: 72000,
		transactions: 3,
		by_method: [
			{ method: 'cash', total: 36000, transactions: 2 },
			{ method: 'qris', total: 36000, transactions: 1 },
			{ method: 'debit', total: 0, transactions: 0 },
			{ method: 'transfer', total: 0, transactions: 0 }
		],
		by_cashier: [{ cashier_id: 1, cashier_name: 'kasir1', total: 72000, transactions: 3 }]
	};

	it('reads the omzet of a day, zero rows and all', () => {
		expect(OmzetHarianEnvelopeSchema.parse({ data: omzet }).data).toEqual(omzet);
	});

	it('rejects a fractional amount: money is whole rupiah in this app', () => {
		expect(() => OmzetHarianEnvelopeSchema.parse({ data: { ...omzet, total: 72000.5 } })).toThrow();
	});

	it('rejects a breakdown that names a method the API does not know', () => {
		expect(() =>
			OmzetHarianEnvelopeSchema.parse({
				data: { ...omzet, by_method: [{ method: 'bitcoin', total: 1, transactions: 1 }] }
			})
		).toThrow();
	});
});

describe('TanggalLaporanSchema', () => {
	it('accepts a store-local calendar date', () => {
		expect(TanggalLaporanSchema.parse(' 2026-09-23 ')).toBe('2026-09-23');
	});

	it('rejects a date in another order', () => {
		expect(() => TanggalLaporanSchema.parse('23-09-2026')).toThrow();
	});

	it('rejects a day that does not exist', () => {
		// "2026-02-30" would roll over to March if it were trusted as a Date.
		expect(() => TanggalLaporanSchema.parse('2026-02-30')).toThrow();
	});
});

describe('tanggalHariIni', () => {
	it('writes today in the shape the report API reads', () => {
		expect(tanggalHariIni(new Date(2026, 8, 23))).toBe('2026-09-23');
	});

	it('pads a single-digit month and day', () => {
		expect(tanggalHariIni(new Date(2026, 0, 5))).toBe('2026-01-05');
	});
});
