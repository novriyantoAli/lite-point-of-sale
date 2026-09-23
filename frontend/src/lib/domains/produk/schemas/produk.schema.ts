import { z } from 'zod';
import { positiveWholeNumber, wholeNumber } from '$lib/utils';

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
 * the schema parses — then enforces exactly the rules `usecase/produk` on the Go
 * side enforces, so a field error and a request error cannot disagree.
 *
 * A blank field fails instead of quietly becoming 0: an empty Harga is almost
 * always one the Admin forgot, and a Produk priced 0 by accident is worse than a
 * form that asks again.
 *
 * The parsing itself lives in `$lib/utils` alongside `collectFieldErrors`: the
 * penjualan domain parses a payment amount the same way, and a second copy of
 * "what does 18.000 mean" is exactly the kind of rule that drifts.
 */

/**
 * What the form fills in to add a Produk.
 *
 * `stock` is the Stok awal the Produk starts life with (CONTEXT.md, Stok): it is
 * set here, once. From then on only the Tambah Stok form raises it and only a
 * Penjualan lowers it — which is why the update schema below has no such field.
 */
export const CreateProdukInputSchema = z.object({
	name: z.string().trim().min(1, 'Nama Produk wajib diisi.'),
	code: optionalText,
	price: wholeNumber('Harga harus bilangan bulat.', 'Harga tidak boleh negatif.'),
	category: optionalText,
	stock: wholeNumber('Stok harus bilangan bulat.', 'Stok tidak boleh negatif.'),
	/**
	 * The Status a new Produk starts with. It is optional because only create reads
	 * it: an edit cannot decide a Produk's Status — the Aktifkan/Nonaktifkan button
	 * is the one action that does — so an absent value means Aktif, a default that
	 * belongs to the Go side.
	 */
	active: z.boolean().optional()
});
export type CreateProdukInput = z.infer<typeof CreateProdukInputSchema>;

/**
 * What the form fills in to change a Produk: the editable record, which is nama,
 * Kode, harga and Kategori.
 *
 * It is derived from the create schema so the two can never disagree about what a
 * Nama or a Harga is, and it omits `stock` and `active` rather than accepting them
 * and letting the API drop them (ADR-0014). Stok moves through the Stok awal of a
 * create, the Tambah Stok form, and a Penjualan — never through an edit. A form
 * that kept a Stok field would send back the number it read when it opened and
 * overwrite a delivery that arrived in between; with no field to send, that cannot
 * happen at any layer.
 */
export const UpdateProdukInputSchema = CreateProdukInputSchema.omit({
	stock: true,
	active: true
});
export type UpdateProdukInput = z.infer<typeof UpdateProdukInputSchema>;

/**
 * What the Aktifkan/Nonaktifkan button posts. It is a body of its own rather than
 * a field on `CreateProdukInputSchema`, because Status is not an edit-form
 * decision and because it crosses the HTTP boundary like every other body, so it
 * is parsed like every other body instead of being handed to axios as a bare
 * object.
 */
export const SetActiveInputSchema = z.object({
	active: z.boolean()
});
export type SetActiveInput = z.infer<typeof SetActiveInputSchema>;

/**
 * What the restock form posts: how many units arrived for one Produk. It carries
 * the quantity, never the new total — the Stok already on the Produk is the
 * API's to know, and sending a total would let two restocks overwrite each other
 * (`usecase/produk.AddStock`).
 *
 * A quantity of zero is refused for the same reason it is on the Go side: it
 * records a delivery that did not happen. A negative one is refused harder —
 * Stok leaves the catalogue through a Penjualan, which is the transaction that
 * keeps it non-negative (#6), not through an Admin's form.
 */
export const TambahStokInputSchema = z.object({
	quantity: positiveWholeNumber(
		'Jumlah Stok harus bilangan bulat.',
		'Jumlah Stok harus lebih dari nol.'
	)
});
export type TambahStokInput = z.infer<typeof TambahStokInputSchema>;

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

/**
 * The restock list: the Produk whose Stok is menipis or habis, and the threshold
 * that decided which ones those are. The threshold is required rather than
 * defaulted — it is the rule the list was selected by, and a UI that invented
 * its own would show the Admin a badge that disagrees with the list underneath
 * it.
 */
export const StokMenipisSchema = z.object({
	data: z.object({
		threshold: z.number().int(),
		products: z.array(ProdukSchema)
	})
});
export type StokMenipis = z.infer<typeof StokMenipisSchema>['data'];
