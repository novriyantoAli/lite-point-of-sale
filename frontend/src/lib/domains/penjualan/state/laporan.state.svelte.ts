import { tanggalHariIni } from '../schemas/penjualan.schema';

/**
 * The UI-only state of the Laporan screen: which day it is reading. The day is a
 * choice the Admin made, not server data, so it lives in runes and never in the
 * query cache — the omzet and the sales list it selects are the queries' to own
 * (Golden Rule 6, ADR-0006).
 *
 * A module singleton rather than component state, so the day survives navigating
 * away from the report and back: an Admin who looked at yesterday, opened a
 * Penjualan, and returned comes back to yesterday.
 */
class LaporanState {
	/**
	 * The chosen day, in the shape the API reads (`YYYY-MM-DD`). It starts on
	 * today, which is what an Admin opening a report almost always wants.
	 */
	tanggal = $state(tanggalHariIni());

	/** Back to today, so tests start from a known day. */
	reset() {
		this.tanggal = tanggalHariIni();
	}
}

export const laporanState = new LaporanState();
