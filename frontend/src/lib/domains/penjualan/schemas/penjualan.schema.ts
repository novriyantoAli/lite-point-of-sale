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
 * The four Pembayaran methods of CONTEXT.md. All four are named because a stored
 * Penjualan has to stay readable when #7 lands the non-tunai ones; the checkout
 * input below accepts Tunai only.
 */
export const MetodePembayaranSchema = z.enum(['cash', 'qris', 'debit', 'transfer']);
export type MetodePembayaran = z.infer<typeof MetodePembayaranSchema>;

/** How each method is written for a person. `cash` is Tunai (CONTEXT.md). */
export const METODE_LABEL: Record<MetodePembayaran, string> = {
	cash: 'Tunai',
	qris: 'QRIS',
	debit: 'Debit',
	transfer: 'Transfer'
};

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
 * The nominal the buyer handed over. It is its own schema so the payment form can
 * validate the field on its own with the same rule the request body uses — a
 * second copy of "what does 50.000 mean" is exactly the kind of rule that drifts.
 */
export const JumlahBayarSchema = wholeNumber(
	'Jumlah bayar harus bilangan bulat.',
	'Jumlah bayar tidak boleh negatif.'
);

/**
 * The body of a checkout.
 *
 * `method` is the literal `cash`, not the whole MetodePembayaranSchema: the till
 * offers Tunai and only Tunai until #7 lands the other three, and a schema that
 * accepted QRIS here would only move the refusal to the API.
 *
 * Whether the amount covers the total is a form-level check, not one here: the
 * schema does not know the total, which is the keranjang's to work out. The API
 * checks it again, and that check is the one that decides.
 */
export const CheckoutInputSchema = z.object({
	items: z.array(CheckoutItemSchema).min(1, 'Keranjang masih kosong.'),
	payment: z.object({
		method: z.literal('cash'),
		amount: JumlahBayarSchema
	})
});
export type CheckoutInput = z.infer<typeof CheckoutInputSchema>;

/** Every answer that carries one Penjualan under `data.sale`. */
export const PenjualanEnvelopeSchema = z.object({
	data: z.object({ sale: PenjualanSchema })
});
