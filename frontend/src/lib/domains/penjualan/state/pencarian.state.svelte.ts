/**
 * Saringan katalog yang sedang diketik di layar Kasir: dua field di kolom kiri
 * yang menyaring mosaik di kolom tengah (CONTEXT.md, Kode).
 *
 * Ini keadaan UI, bukan data server — yang disimpan hanya apa yang diketik, dan
 * hasilnya tetap milik query. Dua kolom bersebelahan membacanya, dan itulah
 * sebabnya ia tinggal di `state/` alih-alih di dalam salah satu komponen:
 * meneruskan dua field lewat props antar saudara kandung akan membuat Kasir tahu
 * urusan dalam pencarian (skill §6.4).
 */
class PencarianState {
	/** Kode yang diketik atau di-scan barcode. */
	code = $state('');
	/** Nama Produk yang dicari. */
	name = $state('');

	/**
	 * Mengosongkan kedua saringan. Dipanggil saat satu Penjualan tersimpan: layar
	 * kembali ke kasir yang bersih untuk pembeli berikutnya, bukan ke saringan
	 * pembeli sebelumnya.
	 */
	reset() {
		this.code = '';
		this.name = '';
	}
}

export const pencarianState = new PencarianState();
