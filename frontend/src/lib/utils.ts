import { z, type ZodError } from 'zod';

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

/**
 * Only a plain integer literal counts as a number. `z.coerce.number()` would run
 * `Number()` first, and `Number('18.000')` is 18 — so somebody who typed the
 * Indonesian thousands separator would save a price at a thousandth of the
 * amount the list then showed back ("Rp 18.000"), with nothing to notice.
 *
 * Money in this app has no decimals (CONTEXT.md), so a `.` or `,` in the field is
 * a separator these forms do not take. Failing is the honest answer: the schema
 * cannot tell whether `18.000` meant 18000 or a mistyped 18, and guessing wrong
 * by 1000× is worse than asking again.
 */
const INTEGER_LITERAL = /^[+-]?\d+$/;

function parseIntegerLiteral(value: string): number {
	const trimmed = value.trim();

	return INTEGER_LITERAL.test(trimmed) ? Number(trimmed) : Number.NaN;
}

/**
 * A whole-number field of a form, accepting the text a form submits or a number
 * already parsed. Money and quantities arrive as text while the API stores
 * integers, so the field parses — and the rule about how large the number may be
 * is a separate argument, because the fields disagree about the rule (a price may
 * be 0, a quantity may not) and must not disagree about how "18.000" is read.
 *
 * Shared by the domains rather than copied into each schema: a second copy of
 * this parsing is a second answer to what `18.000` means.
 */
function wholeNumberField(integerMessage: string, rule: (schema: z.ZodNumber) => z.ZodNumber) {
	return z.preprocess(
		(value) => (typeof value === 'string' ? parseIntegerLiteral(value) : value),
		// The type check carries the same message as `.int()`: a value that is not a
		// number at all ("seribu") and a blank one both arrive as NaN, and zod
		// reports those from the type check — before `.int()` ever runs. Without
		// this, the form shows zod's English default instead of the message the
		// person needs to read.
		rule(z.number({ message: integerMessage }).int(integerMessage))
	);
}

/** A whole number that may be zero. */
export function wholeNumber(integerMessage: string, negativeMessage: string) {
	return wholeNumberField(integerMessage, (schema) => schema.nonnegative(negativeMessage));
}

/** A whole number that must be more than zero. */
export function positiveWholeNumber(integerMessage: string, positiveMessage: string) {
	return wholeNumberField(integerMessage, (schema) => schema.positive(positiveMessage));
}
