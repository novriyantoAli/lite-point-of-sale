import { z } from 'zod';
import { positiveWholeNumber, wholeNumber } from '$lib/utils';

/**
 * Contract of the penjualan domain, mirroring the Go DTOs of `adapter/httpapi`
 * 1:1 (ADR-0006): the schema is the only thing allowed to cross the HTTP
 * boundary, and it is the single source of the types below it.
 *
 * The field names are the API's, so they are English; the domain term, the route
 * (`/penjualan`) and the types here stay Indonesian — the boundary ADR-0012
 * records.
 */

/**
 * The four Pembayaran methods of CONTEXT.md, in the order the till offers them:
 * Tunai first, then the three that are only recorded.
 */
export const METODE_URUT = ['cash', 'qris', 'debit', 'transfer'] as const;

/**
 * The Pembayaran method of a Penjualan. A checkout takes all four: Tunai takes
 * money, and QRIS, Debit and Transfer are recorded without a gateway
 * (CONTEXT.md, Pembayaran).
 */
export const MetodePembayaranSchema = z.enum(METODE_URUT);
export type MetodePembayaran = z.infer<typeof MetodePembayaranSchema>;

/** How each method is written for a person. `cash` is Tunai (CONTEXT.md). */
export const METODE_LABEL: Record<MetodePembayaran, string> = {
	cash: 'Tunai',
	qris: 'QRIS',
	debit: 'Debit',
	transfer: 'Transfer'
};

/**
 * Whether a Pembayaran by this method can produce a Kembalian. Only Tunai can be
 * handed back; QRIS, Debit and Transfer pay the total exactly, so their Kembalian
 * is always zero (CONTEXT.md, Kembalian).
 *
 * The form and the Struk both ask this instead of comparing against the `cash`
 * literal, so the rule has one answer rather than two that can drift.
 */
export function punyaKembalian(method: MetodePembayaran): boolean {
	return method === 'cash';
}

/**
 * One Item of a Penjualan: the Produk it came from, plus the name and price as
 * they stood at checkout. The name and the price are copies — a Produk renamed
 * or repriced afterwards does not change a sale that already happened.
 */
export const ItemPenjualanSchema = z.object({
	product_id: z.number().int(),
	name: z.string(),
	price: z.number().int(),
	quantity: z.number().int(),
	/** What this line contributed to the total: `price × quantity`. */
	subtotal: z.number().int()
});
export type ItemPenjualan = z.infer<typeof ItemPenjualanSchema>;

/**
 * The one Pembayaran of a Penjualan. For Tunai, `amount` is what the buyer
 * handed over and `change` is the Kembalian; for a non-tunai method `change` is
 * zero.
 */
export const PembayaranSchema = z.object({
	method: MetodePembayaranSchema,
	amount: z.number().int(),
	change: z.number().int()
});
export type Pembayaran = z.infer<typeof PembayaranSchema>;

/** A Penjualan as the API ever answers it. It is final: nothing voids one. */
export const PenjualanSchema = z.object({
	/** The Nomor Struk: global, unique, never reused, never reset daily. */
	receipt_number: z.number().int(),
	/** Store-local time (`YYYY-MM-DD HH:MM:SS`), as the API stores it. */
	created_at: z.string(),
	cashier_id: z.number().int(),
	/** The username of whoever rang it up, copied at checkout. */
	cashier_name: z.string(),
	total: z.number().int(),
	items: z.array(ItemPenjualanSchema),
	payment: PembayaranSchema
});
export type Penjualan = z.infer<typeof PenjualanSchema>;

/**
 * One line of the cart as it is posted: which Produk, and how many units. The
 * name and the price are deliberately absent — the catalogue is the API's to
 * read, and a client that sent them could write its own history.
 *
 * The quantity goes through the same whole-number field as every other number a
 * form submits, so a typed "2.000" is refused here rather than read as 2.
 */
export const CheckoutItemSchema = z.object({
	product_id: z.number().int().positive(),
	quantity: positiveWholeNumber(
		'Jumlah Item harus bilangan bulat.',
		'Jumlah Item harus lebih dari nol.'
	)
});
export type CheckoutItem = z.infer<typeof CheckoutItemSchema>;

/**
 * The nominal of the Pembayaran: for Tunai what the buyer handed over, for a
 * recorded method the total of the Penjualan. It is its own schema so the payment
 * form can validate the field on its own with the same rule the request body
 * uses — a second copy of "what does 50.000 mean" is exactly the kind of rule that
 * drifts.
 */
export const JumlahBayarSchema = wholeNumber(
	'Jumlah bayar harus bilangan bulat.',
	'Jumlah bayar tidak boleh negatif.'
);

/**
 * The body of a checkout.
 *
 * `method` is any of the four methods of CONTEXT.md. For Tunai, `amount` is what
 * the buyer handed over and the API works the Kembalian out from it; for a
 * recorded method it is the total of the sale, and the API refuses anything else
 * — a non-tunai Pembayaran has no Kembalian to absorb a difference (CONTEXT.md,
 * Pembayaran, Kembalian).
 *
 * Whether a Tunai amount covers the total is a form-level check, not one here:
 * the schema does not know the total, which is the keranjang's to work out. The
 * API checks it again, and that check is the one that decides.
 */
export const CheckoutInputSchema = z.object({
	items: z.array(CheckoutItemSchema).min(1, 'Keranjang masih kosong.'),
	payment: z.object({
		method: MetodePembayaranSchema,
		amount: JumlahBayarSchema
	})
});
export type CheckoutInput = z.infer<typeof CheckoutInputSchema>;

/** Every answer that carries one Penjualan under `data.sale`. */
export const PenjualanEnvelopeSchema = z.object({
	data: z.object({ sale: PenjualanSchema })
});

/**
 * The outcome of printing a Struk: whether the paper came out, and — when it did
 * not — the reason a Kasir can read.
 *
 * Both print paths answer this same shape (ADR-0017, keputusan 5): the automatic
 * print a checkout reports, and the reprint of a sale that already happened.
 */
export const HasilCetakSchema = z.object({
	printed: z.boolean(),
	/**
	 * The reason a Struk did not come out, written for the person at the till. The
	 * API leaves it out on success, so it is optional rather than an empty string.
	 */
	message: z.string().optional()
});
export type HasilCetak = z.infer<typeof HasilCetakSchema>;

/**
 * The answer to a checkout: the Penjualan it stored, and the result of the Struk
 * print that followed it.
 *
 * The print is reported, never allowed to fail the sale — the money has already
 * moved, and a Struk is a document that can be issued again (ADR-0017,
 * keputusan 1).
 */
export const CheckoutEnvelopeSchema = z.object({
	data: z.object({
		sale: PenjualanSchema,
		print: HasilCetakSchema
	})
});
export type HasilCheckout = z.infer<typeof CheckoutEnvelopeSchema>['data'];

/** The answer to a reprint: the outcome of the print, and nothing else. */
export const CetakEnvelopeSchema = z.object({
	data: z.object({ print: HasilCetakSchema })
});

/**
 * The Nomor Struk a person types into the lookup screen. It is the rule Go
 * applies to the path itself (`receiptNumberParam`): a whole number more than
 * zero. Keeping it here lets the field refuse "abc" without a round trip, while
 * a number that names no Penjualan is still Go's 404 to answer — the schema
 * knows what a Nomor Struk looks like, never which ones exist.
 *
 * An untouched field gets its own message. `positiveWholeNumber` reads "" as a
 * malformed number and answers "harus bilangan bulat", which is the wrong
 * complaint about a box nobody has typed in yet.
 */
export const NomorStrukSchema = z
	.union([z.string(), z.number()])
	.transform((value, ctx): unknown => {
		if (typeof value !== 'string') {
			return value;
		}

		const trimmed = value.trim();
		if (trimmed === '') {
			ctx.addIssue({ code: 'custom', message: 'Nomor Struk wajib diisi.' });
			return z.NEVER;
		}

		return trimmed;
	})
	.pipe(
		positiveWholeNumber('Nomor Struk harus bilangan bulat.', 'Nomor Struk harus lebih dari nol.')
	);
export type NomorStruk = z.infer<typeof NomorStrukSchema>;

/**
 * One row of the sales list (#9): the Nomor Struk and the few fields the list
 * shows, without the Item lines a detail read carries. Opening a row reads the
 * full Penjualan by its Nomor Struk through `PenjualanSchema` — the list is a
 * view over the same record, not a second copy of it.
 */
export const PenjualanRingkasSchema = z.object({
	receipt_number: z.number().int(),
	created_at: z.string(),
	cashier_id: z.number().int(),
	cashier_name: z.string(),
	total: z.number().int(),
	method: MetodePembayaranSchema
});
export type PenjualanRingkas = z.infer<typeof PenjualanRingkasSchema>;

/** The answer to the sales list: one row per Penjualan of the day. */
export const PenjualanListSchema = z.object({
	data: z.array(PenjualanRingkasSchema)
});

/**
 * What one Pembayaran method contributed to a day: how many Penjualan and how
 * much. A method nobody used is a zero row, not a missing one — the API answers
 * all four (CONTEXT.md, Pembayaran), so the screen renders the same rows every
 * day.
 */
export const RingkasanMetodeSchema = z.object({
	method: MetodePembayaranSchema,
	total: z.number().int(),
	transactions: z.number().int()
});
export type RingkasanMetode = z.infer<typeof RingkasanMetodeSchema>;

/** What one Kasir rang up in a day. */
export const RingkasanKasirSchema = z.object({
	cashier_id: z.number().int(),
	cashier_name: z.string(),
	total: z.number().int(),
	transactions: z.number().int()
});
export type RingkasanKasir = z.infer<typeof RingkasanKasirSchema>;

/**
 * The omzet of one store-local day (#9): the total, the number of Penjualan, and
 * the breakdown by Pembayaran method and by Kasir.
 */
export const OmzetHarianSchema = z.object({
	date: z.string(),
	total: z.number().int(),
	transactions: z.number().int(),
	by_method: z.array(RingkasanMetodeSchema),
	by_cashier: z.array(RingkasanKasirSchema)
});
export type OmzetHarian = z.infer<typeof OmzetHarianSchema>;

/** Every answer that carries the omzet of a day under `data`. */
export const OmzetHarianEnvelopeSchema = z.object({
	data: OmzetHarianSchema
});

/**
 * The day a report is over: a store-local calendar date (`YYYY-MM-DD`), the shape
 * the API reads. It is its own schema so the date field can refuse a malformed
 * day before it becomes a query key — Go refuses one too, and it is the same
 * rule in both places (ADR-0015).
 */
export const TanggalLaporanSchema = z
	.string()
	.trim()
	.refine(isCalendarDate, 'Tanggal laporan tidak valid.');
export type TanggalLaporan = z.infer<typeof TanggalLaporanSchema>;

/**
 * Whether `value` is a real `YYYY-MM-DD` date. The round-trip through `Date`
 * rejects a day that does not exist ("2026-02-30" rolls over to March), and the
 * `Z` keeps the check in UTC so a store east of Greenwich cannot see yesterday.
 */
function isCalendarDate(value: string): boolean {
	if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
		return false;
	}

	const date = new Date(`${value}T00:00:00Z`);

	return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value;
}

/**
 * The store-local date of "today", in the shape the report API reads. One store
 * and one terminal (ADR-0002): the browser's calendar is the store's calendar,
 * and the API writes `created_at` with the same local clock (ADR-0015).
 */
export function tanggalHariIni(now: Date = new Date()): string {
	const year = now.getFullYear();
	const month = String(now.getMonth() + 1).padStart(2, '0');
	const day = String(now.getDate()).padStart(2, '0');

	return `${year}-${month}-${day}`;
}
