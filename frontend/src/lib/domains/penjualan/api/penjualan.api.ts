import { apiClient } from '$lib/api/client';
import {
	CheckoutInputSchema,
	PenjualanEnvelopeSchema,
	type CheckoutInput,
	type Penjualan
} from '../schemas/penjualan.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 *
 * Every call goes to the SvelteKit BFF on this origin — never to Go. The session
 * cookie rides along by itself, which is the whole point of Pattern A (ADR-0001,
 * ADR-0006).
 */
export interface PenjualanApi {
	/**
	 * Checks a cart out into a finished Penjualan. The API prices the Items from
	 * the catalogue, refuses a cart the Stok cannot cover, and works out the
	 * Kembalian — none of which this side sends or decides.
	 */
	checkout(input: CheckoutInput): Promise<Penjualan>;
}

export const penjualanApi: PenjualanApi = {
	async checkout(input: CheckoutInput): Promise<Penjualan> {
		const { data } = await apiClient.post('/penjualan', CheckoutInputSchema.parse(input));

		return PenjualanEnvelopeSchema.parse(data).data.sale;
	}
};
