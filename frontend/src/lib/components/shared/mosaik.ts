/**
 * Kosakata kelas dunia mosaik — satu tempat untuk fakta visual yang dipakai
 * setiap layar.
 *
 * Sebelas port pertama menyalin konstanta ini ke tiap komponen, dan setelah
 * layar terakhir mendarat salinannya ada di empat belas komponen. Kelas yang sama
 * ditulis di banyak tempat akan berbeda sendiri dalam dua port berikutnya; di
 * sini ia ditulis sekali, jadi aturan tingkat dunia punya satu wujud yang bisa
 * dibaca dan diubah.
 *
 * Yang tinggal di sini adalah **fakta satu baris**, bukan markah. Ia bukan modul
 * komponen: shape yang bervariasi (strip dengan atau tanpa aksi, kepala modul
 * dengan catatan atau tombol) tetap disusun pemanggilnya, karena membungkusnya
 * jadi komponen berarti satu prop per variasi — antarmuka yang lebih besar
 * daripada yang digantikannya (DESIGN.md; solution doc, "Yang belum diputuskan").
 *
 * Pemindai Tailwind membaca teks berkas, jadi kelas di `.ts` ini tetap
 * ter-generate seperti kalau ia ditulis di komponen (solution doc §5).
 */

/** Papan mosaik: satu kolom selebar papan. Padding luarnya milik rel. */
export const PAPAN = 'grid grid-cols-1 gap-0';

/**
 * Strip judul (DESIGN.md, Typography): judul 24px/700 dengan `-mb-px` plus
 * `z-[2]` supaya garis tintanya yang menutup modul di bawahnya (Shared-Hairline).
 */
export const STRIP =
	'z-[2] col-span-full -mb-px flex flex-wrap items-baseline gap-x-3 gap-y-1 border border-border border-b-foreground bg-card px-2 py-2';
/** Satu-satunya tempat ukuran 24px muncul di sebuah layar. */
export const JUDUL = 'text-2xl leading-none font-bold tracking-[-0.015em]';
/** Lantai huruf dunia ini: tidak ada teks di bawah 12px. */
export const CATATAN = 'text-xs';
/** Catatan yang memuat angka — cacah, ambang — selalu tabular-nums. */
export const CATATAN_ANGKA = `${CATATAN} tabular-nums`;
/** Aksi utama layar yang berdiri di baris strip, mengambil sisa baris ke kanan. */
export const AKSI_STRIP = 'ml-auto flex items-center gap-1.5';

/**
 * Modul (DESIGN.md, Modules): latar petak dengan garis rambut di atas, kiri, dan
 * kanan; garis bawahnya ditiadakan (`border-b-0`) supaya baris terakhir modul yang
 * menutupnya — satu garis dipakai bersama, bukan digambar dua kali. Modul yang
 * tidak bersebelahan menambahkan `mt-2` lewat `cn()`.
 */
export const MODUL = 'border border-border border-b-0 bg-card';
/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
export const MODUL_KEPALA =
	'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
export const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
/** Satu baris di dalam modul; garis bawahnya membentuk garis rambut bersama. */
export const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
/** Baris modul yang isinya teks biasa: keadaan, catatan, atau kalimat kosong. */
export const MODUL_TEKS = `${MODUL_BARIS} text-xs`;
/** Daftar baris: satu garis rambut antar baris, ditutup garis bawah daftar. */
export const DAFTAR = 'divide-y divide-border border-b border-border';

/**
 * Fields & Inputs (DESIGN.md): field 26px dengan label 12px/600 di atasnya,
 * jaraknya 3px. Ukurannya literal, bukan token baru.
 */
export const FIELD = 'flex flex-col gap-[3px]';
export const LABEL = 'text-xs leading-[1.2] font-semibold';
export const INPUT =
	'h-[26px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]';
/** Field yang isinya angka: uang, jumlah, atau Stok — "tanpa kecuali". */
export const INPUT_ANGKA = `${INPUT} tabular-nums`;
/** Field yang tidak valid: pesannya 12px/600, warnanya tinta, bukan merah. */
export const ERROR = 'text-xs font-semibold';
/**
 * Figures (DESIGN.md): angka yang jadi jawaban layar — 20px/700, selalu
 * `tabular-nums`, selalu didahului label 12px/600 yang menyebut apa angka itu.
 */
export const FIGURE = 'text-[20px] leading-[1.1] font-bold tabular-nums';
/**
 * Sel metode (DESIGN.md, Methods): radio asli yang disembunyikan, sel persegi
 * yang terlihat. Terpilih = bidang tinta penuh, bukan tint dan bukan centang.
 * Karena tintanya datang dari `has-[:checked]`, cincin fokusnya ditanggung di
 * sini (tinta di atas tinta terukur 1,00:1 — Browser surfaces).
 */
export const METODE =
	'flex h-[26px] cursor-pointer items-center justify-center border border-border bg-card text-[13px] font-medium transition-colors hover:bg-muted focus-within:outline-2 focus-within:outline-solid focus-within:-outline-offset-2 focus-within:outline-foreground has-[:checked]:focus-within:outline-primary-foreground has-[:checked]:border-foreground has-[:checked]:bg-foreground has-[:checked]:font-semibold has-[:checked]:text-primary-foreground';

/**
 * Quiet tag (DESIGN.md, Tags): tinggi 15px, huruf 12px/600, isian Wash Grey.
 * `TAG` menambahkan `whitespace-nowrap` untuk sel tabel; `TAG_TENANG` adalah
 * bentuk yang sama tanpa itu, untuk tempat yang boleh membungkus.
 */
export const TAG_TENANG =
	'inline-flex h-[15px] items-center border border-border bg-muted px-[5px] text-xs font-semibold';
export const TAG = `${TAG_TENANG} whitespace-nowrap`;
/** Tag bergaris tinta untuk keadaan tidak sehat: bentuk dan kata, bukan merah. */
export const TAG_TINTA =
	'inline-flex h-[15px] items-center border border-foreground bg-card px-[5px] text-xs font-semibold';
/** Red tag: Kode Produk, tab aktif, dan `tersegel` — merah yang diizinkan. */
export const TAG_MERAH =
	'inline-flex h-[15px] items-center bg-destructive px-[5px] text-xs font-semibold text-primary-foreground';
/** Cara dunia ini menyampaikan keadaan mati: garis coret, bukan `opacity`. */
export const CORET = 'line-through decoration-1';
/** Tag mati: `TAG` bergaris coret — kata "Nonaktif" di banyak layar. */
export const TAG_CORET = `${TAG} ${CORET}`;

/**
 * Empat petak tombol dunia ini; yang membedakan mereka ukurannya, bukan warnanya.
 *
 * `AKSI_MODUL` adalah `.btn` 26px/13px di kepala modul dan baris galat;
 * `AKSI_BARIS` adalah 22px/12px di dalam sel tabel. Keadaan mati keduanya
 * kehilangan kursor, bukan tintanya.
 *
 * `COMMIT_BARIS` (26px) dan `COMMIT_KOLOM` (40px) adalah Commit Button: satu-
 * satunya bidang bertinta penuh. Saat mati ia kehilangan tintanya lalu jadi putih
 * bergaris putus-putus, bukan pudar (State-Is-Not-Faded). Selalu dipakai dengan
 * `variant="default"`, yang memang membawa `bg-primary`.
 *
 * `AKSI_MODUL` dan `AKSI_BARIS` selalu dipakai dengan `variant="ghost"`: varian itu
 * sudah membawa `hover:bg-muted` — persis latar hover dunia ini — dan tidak membawa
 * apa pun yang harus dilawan. Varian `default` membawa `text-primary-foreground`,
 * jadi tombol berbingkai tanpa varian berakhir putih di atas putih (solution doc §3).
 */
export const AKSI_MODUL =
	'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
export const AKSI_BARIS =
	'h-[22px] border-border bg-card px-1.5 text-xs font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
export const COMMIT_BARIS =
	'h-[26px] border-foreground bg-primary px-2 text-[13px] font-semibold text-primary-foreground hover:border-foreground hover:bg-primary hover:text-primary-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:border-dashed disabled:border-border disabled:bg-card disabled:text-foreground disabled:opacity-100';
export const COMMIT_KOLOM =
	'h-10 w-full text-[15px] font-semibold hover:bg-primary disabled:pointer-events-auto disabled:cursor-not-allowed disabled:border-dashed disabled:border-border disabled:bg-card disabled:text-foreground disabled:opacity-100';

/**
 * Tabel mosaik (DESIGN.md, Typography: kepala tabel 12px/600; "setiap angka …
 * tabular-nums, tanpa kecuali"). Satu garis rambut dipakai bersama antar baris,
 * kepala Wash Grey ditutup garis tinta, dan sel angka selalu tabular.
 *
 * Kosakata ini dulu disalin ke tiap layar yang punya tabel — Produk, Stok,
 * Laporan — dan salinannya sempat berbeda sendiri: `TD_ANGKA` di satu layar
 * kehilangan `tabular-nums` yang di layar lain masih ada. Sekarang ia ditulis
 * sekali, jadi aturan "angka tabular" punya satu wujud yang bisa dijaga tes.
 */
export const TABEL = 'w-full border-collapse text-[13px] leading-[1.25]';
/** Kotak tabel: gulir mendatar di dalam modul kalau layarnya sempit. */
export const TABEL_BUNGKUS = 'overflow-x-auto border-b border-border';
/** Kepala kolom: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
export const TH =
	'border-b border-foreground bg-muted px-2 py-1.5 text-left text-xs font-semibold whitespace-nowrap';
/** Kepala kolom angka: rata kanan dan tabular. */
export const TH_ANGKA = `${TH} text-right tabular-nums`;
/** Satu baris tabel; garis bawahnya membentuk garis rambut bersama. */
export const BARIS = 'border-b border-border transition-colors last:border-b-0 hover:bg-muted';
/** Sel tabel. */
export const TD = 'px-2 py-1.5 align-top';
/** Sel angka: rata kanan dan tabular — "tanpa kecuali" (DESIGN.md, Typography). */
export const TD_ANGKA = `${TD} text-right tabular-nums`;
/** Sel keterangan: teks 12px, seperti catatan di luar tabel. */
export const TD_SUB = `${TD} text-xs`;
/** Sel aksi: sederet tombol yang tidak boleh membungkus. */
export const TD_AKSI = `${TD} whitespace-nowrap`;
/** Sel nama: satu-satunya yang ditebalkan di barisnya. */
export const TD_NAMA = `${TD} font-medium`;
/** Sel Harga: merah utilitas dan tabular, tidak pernah monospace (DESIGN.md, Typography). */
export const TD_HARGA = `${TD_ANGKA} font-semibold text-destructive`;
