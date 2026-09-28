import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import { keranjangState } from '../state/keranjang.state.svelte';
import Pembayaran from './Pembayaran.svelte';

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/penjualan.api', () => ({
	penjualanApi: { checkout: vi.fn(), getByReceiptNumber: vi.fn(), cetakStruk: vi.fn() }
}));

function renderPembayaran() {
	return render(Pembayaran, { onCheckedOut: vi.fn() }, { wrapper: QueryClientHarness });
}

beforeEach(() => {
	keranjangState.clear();
});

describe('Pembayaran', () => {
	/**
	 * `data-caret="utility"` is the only thread between this field and the world's
	 * `[data-caret='utility'] { caret-color: var(--destructive) }` in `app.css`
	 * (DESIGN.md, Browser surfaces). Without the attribute the field's caret is ink
	 * like every other field — a defect no jsdom test can see, because a caret is not
	 * text, layout, or a name. The running-DOM guard for it lives in
	 * `tests/e2e/permukaan-peramban.spec.ts`; this line is only the unit half.
	 */
	it('menandai field jumlah bayar dengan data-caret="utility"', () => {
		renderPembayaran();

		expect(screen.getByLabelText('Jumlah bayar')).toHaveAttribute('data-caret', 'utility');
	});
});
