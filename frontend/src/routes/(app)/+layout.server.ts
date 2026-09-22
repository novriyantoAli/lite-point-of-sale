import type { LayoutServerLoad } from './$types';

/**
 * The session `hooks.server.ts` has already resolved, handed to the layout as
 * load data. The client never asks for it again: the Pengguna is rendered from
 * this, not kept in a second copy in the query cache (ADR-0006).
 */
export const load: LayoutServerLoad = ({ locals }) => {
	return { user: locals.user };
};
