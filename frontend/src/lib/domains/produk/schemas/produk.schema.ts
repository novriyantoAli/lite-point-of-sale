import { z } from 'zod';

/**
 * Contract of the produk domain, mirroring the Go DTOs of `adapter/httpapi` 1:1
 * (ADR-0006): the schema is the only thing allowed to cross the HTTP boundary,
 * and it is the single source of the types below it.
 *
 * The field names are the API's, so they are English; the domain term, the route
 * (`/produk`) and the types here stay Indonesian — the boundary ADR-0012 records.
 */

/** A Produk as the API ever answers it. */
export const ProdukSchema = z.object({
	id: z.number().int(),
	name: z.string(),
	/** The Kode: an optional barcode or internal code, unique when present. */
	code: z.string().nullable(),
	/** Harga in whole rupiah — never a fraction, which is why it is an int. */
	price: z.number().int(),
	/** The optional one-level Kategori, for grouping and filtering only. */
	category: z.string().nullable(),
	stock: z.number().int(),
	/** Active or Nonaktif: a Nonaktif Produk is gone from the kasir lookup. */
	active: z.boolean(),
	/**
	 * Whether this Produk was part of a finished Penjualan. A Sold Produk can be
	 * deactivated but never deleted, so the UI needs it to offer the right action.
	 */
	sold: z.boolean()
});
export type Produk = z.infer<typeof ProdukSchema>;

/**
 * Blank text is how the form says "none", and `null` is what the API stores for
 * a Produk without a Kode or Kategori — so the two are normalized here, once,
 * instead of every caller guessing what an empty string means.
 *
 * A field the caller left out entirely means the same thing, hence `.optional()`
 * alongside `.nullable()`: the two are not the same, and without it an absent
 * Kode fails validation and its message buries the real error on Harga or Stok.
 */
const optionalText = z
	.string()
	.trim()
	.nullable()
	.optional()
	.transform((value) => (value === '' || value === undefined ? null : value));

/**
 * Money and Stok arrive from a form as text while the API stores integers, so
 * the schema coerces — then enforces exactly the rules `usecase/produk` on the
 * Go side enforces, so a field error and a request error cannot disagree.
 *
 * A blank field fails instead of quietly becoming 0: an empty Harga is almost
 * always one the Admin forgot, and a Produk priced 0 by accident is worse than a
 * form that asks again.
 */
function wholeNumber(integerMessage: string, negativeMessage: string) {
	return z.preprocess(
		(value) => (typeof value === 'string' && value.trim() === '' ? Number.NaN : value),
		// The type check carries the same message as `.int()`: a value that is not a
		// number at all ("seribu") and a blank one both arrive as NaN, and zod
		// reports those from the type check — before `.int()` ever runs. Without
		// this, the form shows zod's English default instead of the message the
		// Admin needs to read.
		z.coerce.number({ message: integerMessage }).int(integerMessage).nonnegative(negativeMessage)
	);
}

/** What the form fills in to add or change a Produk. */
export const ProdukInputSchema = z.object({
	name: z.string().trim().min(1, 'Nama Produk wajib diisi.'),
	code: optionalText,
	price: wholeNumber('Harga harus bilangan bulat.', 'Harga tidak boleh negatif.'),
	category: optionalText,
	stock: wholeNumber('Stok harus bilangan bulat.', 'Stok tidak boleh negatif.')
});
export type ProdukInput = z.infer<typeof ProdukInputSchema>;

/**
 * The catalogue filters. Every field is optional: an empty one means "do not
 * filter on this", which is why they are trimmed here rather than at the URL.
 */
export const ProdukFilterSchema = z.object({
	name: z.string().trim().default(''),
	code: z.string().trim().default(''),
	category: z.string().trim().default(''),
	/** `undefined` asks for everything, which is what an Admin's screen wants. */
	active: z.boolean().optional()
});
export type ProdukFilter = z.infer<typeof ProdukFilterSchema>;

/** Every answer that carries one Produk under `data.product`. */
export const ProdukEnvelopeSchema = z.object({
	data: z.object({ product: ProdukSchema })
});

/** What Go answers to a catalogue listing. */
export const ProdukListSchema = z.object({
	data: z.array(ProdukSchema)
});

/** The Kategori in use, for the filter's dropdown. */
export const KategoriListSchema = z.object({
	data: z.array(z.string())
});
