import type { ZodError } from 'zod';

export { cn } from 'cn';

/**
 * Harga, total, and every other amount as it is written for a person: whole
 * rupiah, grouped in thousands. The grouping is `Intl`'s, but the `Rp` prefix
 * is added by hand — the currency style would join them with a non-breaking
 * space, which reads as a different string to anything comparing or trimming it.
 *
 * Money has no decimals in this app (CONTEXT.md), so the fraction digits are
 * dropped rather than formatted as `,00`.
 */
const RUPIAH_GROUPING = new Intl.NumberFormat('id-ID');

export function formatRupiah(amount: number): string {
	return `Rp ${RUPIAH_GROUPING.format(amount)}`;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };

/**
 * Collects the first zod issue of each field, so a form can show one message per
 * input instead of dumping every issue at once (§11). Issues that belong to no
 * field in `fields` are ignored — those are the form-level ones.
 */
export function collectFieldErrors<T extends string>(
	error: ZodError,
	fields: readonly T[]
): Partial<Record<T, string>> {
	const errors: Partial<Record<T, string>> = {};

	for (const issue of error.issues) {
		const field = issue.path[0];
		if (typeof field === 'string' && (fields as readonly string[]).includes(field)) {
			const known = field as T;
			errors[known] ??= issue.message;
		}
	}

	return errors;
}
