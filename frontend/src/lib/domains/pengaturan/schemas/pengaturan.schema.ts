import { z } from 'zod';
import { positiveWholeNumber } from '$lib/utils';

/**
 * Contract of the pengaturan domain, mirroring the Go DTOs of
 * `adapter/httpapi` 1:1 (ADR-0006): the schema is the only thing allowed to
 * cross the HTTP boundary, and it is the single source of the types below it.
 *
 * The field names are the API's, so they are English; the domain term, the route
 * (`/pengaturan`) and the types here stay Indonesian — the boundary ADR-0012
 * records.
 */

/** The two thermal roll widths, in mm (ADR-0017, keputusan 4). */
export const PAPER_WIDTH_OPTIONS = [
	{ value: '58', label: '58 mm' },
	{ value: '80', label: '80 mm' }
] as const;

/**
 * The Struk paper width as a form submits it: the Select's string choice, parsed
 * into the integer the API stores. It is a backstop behind a Select that only
 * offers the two widths — a direct caller sending anything else is refused here
 * with the same rule the Go side enforces.
 */
const paperWidthInput = z.preprocess(
	(value) => (typeof value === 'string' ? Number(value) : value),
	z
		.number({ message: 'Lebar kertas harus 58 atau 80 mm.' })
		.int()
		.refine((value) => value === 58 || value === 80, {
			message: 'Lebar kertas harus 58 atau 80 mm.'
		})
);

/** The store's Pengaturan as the API ever answers it. */
export const PengaturanSchema = z.object({
	id: z.number().int(),
	/** The Struk template block printed above the sale lines (CONTEXT.md, Struk). */
	header: z.string(),
	/** The Struk template block printed below the sale lines. */
	footer: z.string(),
	/** The thermal roll width in mm — 58 or 80. */
	paper_width: z.number().int(),
	/** The ambang below which an Active Produk counts as Stok menipis. */
	low_stock_threshold: z.number().int()
});
export type Pengaturan = z.infer<typeof PengaturanSchema>;

/**
 * What the form fills in to change the Pengaturan: the Struk template blocks,
 * the paper width, and the ambang Stok menipis.
 *
 * Header and footer are free-text blocks — newline-separated lines printed
 * verbatim — so they are not trimmed into a single line the way a Produk Nama
 * would be (ADR-0017, keputusan 4).
 */
export const UpdatePengaturanInputSchema = z.object({
	header: z.string(),
	footer: z.string(),
	paper_width: paperWidthInput,
	low_stock_threshold: positiveWholeNumber(
		'Ambang Stok menipis harus bilangan bulat.',
		'Ambang Stok menipis harus lebih dari nol.'
	)
});
export type UpdatePengaturanInput = z.infer<typeof UpdatePengaturanInputSchema>;

/** Every answer that carries the Pengaturan under `data.settings`. */
export const PengaturanEnvelopeSchema = z.object({
	data: z.object({ settings: PengaturanSchema })
});
