// Audit terukur untuk prototipe Kasir dunia "Japanese High-Density Grid".
// Dijalankan dari frontend/ supaya @playwright/test bisa diresolusi, lalu dihapus.
import { chromium } from '@playwright/test';

const URL = 'file:///home/real/Projects/lite-point-of-sale/.impeccable/build/kasir/index.html';
const OUT = '/tmp/kasir';
const results = [];
const check = (n, got, want) => results.push(`${got === want ? 'PASS' : 'FAIL'}  ${n}: ${JSON.stringify(got)}${got === want ? '' : ` (want ${JSON.stringify(want)})`}`);

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
const problems = [];
page.on('pageerror', (e) => problems.push(`pageerror: ${e.message}`));
page.on('console', (m) => m.type() === 'error' && problems.push(`console: ${m.text()}`));
await page.goto(URL);
await page.evaluate(() => document.fonts.ready);

// ---- kontras seluruh DOM ----
const contrast = await page.evaluate(() => {
	const ctx = document.createElement('canvas').getContext('2d', { willReadFrequently: true });
	const toRgba = (css) => { ctx.clearRect(0, 0, 1, 1); ctx.fillStyle = '#000'; ctx.fillStyle = css; ctx.fillRect(0, 0, 1, 1); const d = ctx.getImageData(0, 0, 1, 1).data; return [d[0], d[1], d[2], d[3] / 255]; };
	const lum = ([r, g, b]) => { const f = (v) => { v /= 255; return v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4); }; return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b); };
	const ratio = (a, b) => { const [x, y] = [lum(a), lum(b)].sort((m, n) => n - m); return (x + 0.05) / (y + 0.05); };
	const bgOf = (el) => { const layers = []; for (let n = el; n; n = n.parentElement) { const c = toRgba(getComputedStyle(n).backgroundColor); if (c[3] > 0) layers.unshift(c); } let out = [255, 255, 255]; for (const l of layers) out = out.map((b, i) => l[i] * l[3] + b * (1 - l[3])); return out; };
	const hex = (c) => '#' + c.map((v) => Math.round(v).toString(16).padStart(2, '0')).join('');
	const seen = new Map();
	for (const el of document.querySelectorAll('body *')) {
		const own = [...el.childNodes].some((n) => n.nodeType === 3 && n.textContent.trim() !== '');
		if (!own) continue;
		const cs = getComputedStyle(el);
		if (cs.visibility === 'hidden' || cs.display === 'none' || Number(cs.opacity) === 0) continue;
		const fg = toRgba(cs.color), bg = bgOf(el);
		const c = fg.slice(0, 3).map((v, i) => v * fg[3] + bg[i] * (1 - fg[3]));
		const px = parseFloat(cs.fontSize);
		const need = px >= 24 || (px >= 18.66 && Number(cs.fontWeight) >= 700) ? 3 : 4.5;
		const key = `${hex(c)} on ${hex(bg)} @${px}px/${cs.fontWeight}`;
		if (!seen.has(key)) seen.set(key, { got: ratio(c, bg), need, sample: el.textContent.trim().slice(0, 34) });
	}
	return [...seen.entries()].map(([k, v]) => ({ key: k, ...v }));
});

// ---- kepadatan: garis rambut dan petak ----
const density = await page.evaluate(() => {
	const all = [...document.querySelectorAll('body *')];
	const bordered = all.filter((el) => { const cs = getComputedStyle(el); return ['Top', 'Right', 'Bottom', 'Left'].some((s) => parseFloat(cs['border' + s + 'Width']) > 0); });
	const radii = new Set(all.map((el) => getComputedStyle(el).borderRadius).filter((r) => r && r !== '0px'));
	return {
		elems: all.length,
		bordered: bordered.length,
		tiles: document.querySelectorAll('.tile').length,
		radii: [...radii],
		shadows: all.filter((el) => getComputedStyle(el).boxShadow !== 'none').length,
		bodyWidth: document.body.getBoundingClientRect().width
	};
});

// ---- aturan merah ≤ 3% dari viewport, diukur dari piksel tangkapan layar ----
const shot = await page.screenshot();
const red = await page.evaluate(async (src) => {
	const img = new Image(); img.src = src; await img.decode();
	const c = document.createElement('canvas'); c.width = img.naturalWidth; c.height = img.naturalHeight;
	const ctx = c.getContext('2d', { willReadFrequently: true }); ctx.drawImage(img, 0, 0);
	const d = ctx.getImageData(0, 0, c.width, c.height).data;
	let n = 0, merah = 0, putih = 0, tinta = 0;
	for (let i = 0; i < d.length; i += 4) {
		const r = d[i], g = d[i + 1], b = d[i + 2]; n++;
		if (r > 150 && r < 230 && g < 70 && b < 70) merah++;
		if (r > 245 && g > 245 && b > 245) putih++;
		if (r < 40 && g < 40 && b < 40) tinta++;
	}
	return { merah: +(100 * merah / n).toFixed(2), putih: +(100 * putih / n).toFixed(2), tinta: +(100 * tinta / n).toFixed(2) };
}, 'data:image/png;base64,' + shot.toString('base64'));

// ---- perilaku ----
const txt = (s) => page.locator(s).innerText();
const val = (s) => page.locator(s).inputValue();
check('petak katalog', density.tiles, 24);
check('total awal', await txt('#keranjang-total'), 'Rp 136.500');
check('Kembalian awal', await txt('#kembalian'), 'Rp 13.500');
check('ringkasan awal', await txt('#keranjang-ringkas'), '7 unit · 2 Item');

await page.locator('.tile[data-add="14"]').click(); // Trafo 5A, stok 4
await page.waitForTimeout(50);
check('tambah petak → total', await txt('#keranjang-total'), 'Rp 281.500');
check('tambah petak → baris', await page.locator('#keranjang-baris > .row').count(), 3);

const q = page.locator('#keranjang-baris [data-id="14"] [data-act="qty"]');
await q.fill('9');
await page.waitForTimeout(50);
check('melebihi stok → field invalid', await q.getAttribute('aria-invalid'), 'true');
check('melebihi stok → blokir', await txt('#terblokir'), 'Ada Item yang melebihi Stok. Kurangi jumlahnya lebih dulu.');
check('melebihi stok → tombol mati', await page.locator('#tombol-bayar').isDisabled(), true);
await q.fill('1');
await page.waitForTimeout(50);
check('kembali → tombol hidup', await page.locator('#tombol-bayar').isEnabled(), true);

await page.locator('#bayar').fill('300.000');
await page.waitForTimeout(50);
check('mengetik: tidak ada galat dulu', await page.locator('#bayar-error').isVisible(), false);
check('mengetik: Kembalian berhenti di —', await txt('#kembalian'), '—');
await page.locator('#tombol-bayar').click();
await page.waitForTimeout(50);
check('submit: "300.000" ditolak', await txt('#bayar-error'), 'Jumlah bayar harus bilangan bulat.');
await page.locator('#bayar').fill('100000');
await page.locator('#tombol-bayar').click();
await page.waitForTimeout(50);
check('submit: kurang dari total', await txt('#bayar-error'), 'Jumlah bayar kurang dari total.');

await page.locator('#bayar').fill('300000');
await page.waitForTimeout(50);
check('Kembalian valid', await txt('#kembalian'), 'Rp 18.500');
await page.locator('#tombol-bayar').click();
await page.waitForTimeout(200);
check('tersimpan: Nomor Struk', await txt('#keranjang-baris b.tnum'), '7');
check('tersimpan: total', await txt('#keranjang-baris .total__value'), 'Rp 281.500');
check('tersimpan: tag', await txt('#keranjang-baris .tag'), 'tersegel');
check('tersimpan: kolom terisi', await page.locator('#penjualan-baru').isVisible(), true);

// ---- keadaan kosong ----
await page.locator('#penjualan-baru').click();
await page.waitForTimeout(80);
for (const b of await page.locator('#keranjang-baris [data-act="hapus"]').all()) await b.click();
await page.waitForTimeout(80);
check('kosong: pesan tampil', await page.locator('#keranjang-kosong').isVisible(), true);
check('kosong: total nol', await txt('#keranjang-total'), 'Rp 0');
check('kosong: tombol mati', await page.locator('#tombol-bayar').isDisabled(), true);
check('kosong: blokir', await txt('#terblokir'), 'Keranjang masih kosong.');

// ---- tangkapan layar ----
// Kembalikan ke keadaan tengah transaksi supaya tangkapannya realistis.
await page.locator('.tile[data-add="0"]').click();
await page.locator('.tile[data-add="1"]').click();
await page.locator('#bayar').fill('150000');
await page.waitForTimeout(100);
await page.screenshot({ path: `${OUT}-1-desktop.png` });
await page.locator('main.board').screenshot({ path: `${OUT}-2-board.png` });
const mobile = await browser.newPage({ viewport: { width: 390, height: 844 } });
await mobile.goto(URL);
await mobile.evaluate(() => document.fonts.ready);
await mobile.waitForTimeout(150);
await mobile.screenshot({ path: `${OUT}-3-mobile.png`, fullPage: true });
const overflow = await mobile.evaluate(() => {
	const w = document.documentElement.clientWidth;
	return { scrollWidth: document.documentElement.scrollWidth, clientWidth: w, cols: getComputedStyle(document.querySelector('.board')).gridTemplateColumns };
});
const deskOverflow = await page.evaluate(() => ({ scrollWidth: document.documentElement.scrollWidth, clientWidth: document.documentElement.clientWidth }));

const fits = await page.evaluate(() => ({ scrollHeight: document.documentElement.scrollHeight, innerHeight: window.innerHeight, tiles: document.querySelectorAll('.tile').length }));
console.log('\n=== KEPADATAN LAYAR 1440x900 ===');
console.log(`  tinggi isi ${fits.scrollHeight}px vs viewport ${fits.innerHeight}px → ${fits.scrollHeight <= fits.innerHeight ? `seluruh katalog (${fits.tiles} Produk) + keranjang + pembayaran muat TANPA menggulir` : 'MASIH PERLU MENGGULIR'}`);

await browser.close();

const fails = contrast.filter((c) => c.got < c.need);
console.log('=== KONTRAS ===');
console.log(fails.length ? fails.map((f) => `  GAGAL ${f.got.toFixed(2)}:1 (butuh ${f.need}) ${f.key} ← "${f.sample}"`).join('\n') : '  semua teks lolos AA');
console.log(`  (${contrast.length - fails.length}/${contrast.length} pasangan unik lolos)`);
console.log('\n=== KEPADATAN ===');
console.log(`  ${density.elems} elemen · ${density.bordered} berbingkai · ${density.tiles} petak`);
console.log(`  radius selain 0: ${density.radii.length ? density.radii.join(', ') : 'tidak ada (nol di mana-mana)'}`);
console.log(`  elemen berbayangan: ${density.shadows}`);
console.log('\n=== ATURAN MERAH (kontrak: ≤ 3% viewport) ===');
console.log(`  merah ${red.merah}% · putih ${red.putih}% · tinta ${red.tinta}%`);
console.log(`  → ${red.merah <= 3 ? 'LOLOS' : 'GAGAL'} kontrak`);
console.log('\n=== 1440px ===');
console.log(`  scrollWidth ${deskOverflow.scrollWidth} vs clientWidth ${deskOverflow.clientWidth} → ${deskOverflow.scrollWidth <= deskOverflow.clientWidth ? 'tanpa geser mendatar' : 'GESER MENDATAR'}`);
console.log('\n=== 390px ===');
console.log(`  scrollWidth ${overflow.scrollWidth} vs clientWidth ${overflow.clientWidth} → ${overflow.scrollWidth <= overflow.clientWidth ? 'tanpa geser mendatar' : 'GESER MENDATAR'}`);
console.log(`  kolom papan: ${overflow.cols}`);
console.log('\n=== PERILAKU ===');
console.log(results.join('\n'));
console.log('\nproblems:', problems.length ? problems.join('\n') : 'none');
console.log('FAILURES:', results.filter((r) => r.startsWith('FAIL')).length);
