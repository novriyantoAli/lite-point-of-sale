import { describe, expect, it } from 'vitest';
import { tanggalHariIni } from '../schemas/penjualan.schema';
import { laporanState } from './laporan.state.svelte';

describe('LaporanState', () => {
	it('starts on today', () => {
		laporanState.reset();

		expect(laporanState.tanggal).toBe(tanggalHariIni());
	});

	it('keeps the day the Admin picked', () => {
		laporanState.tanggal = '2026-09-01';

		expect(laporanState.tanggal).toBe('2026-09-01');
	});

	it('resets back to today', () => {
		laporanState.tanggal = '2026-09-01';

		laporanState.reset();

		expect(laporanState.tanggal).toBe(tanggalHariIni());
	});
});
