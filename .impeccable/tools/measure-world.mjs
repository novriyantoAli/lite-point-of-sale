// Mengukur QUALITY BAR board + hero dunia terpilih, karena model di sesi ini
// tidak bisa melihat gambar. Dijalankan dari frontend/ supaya @playwright/test
// bisa diresolusi, lalu dihapus.
import { chromium } from '@playwright/test';
import { readFileSync } from 'node:fs';

const FILES = [
	['board', '/home/real/Projects/lite-point-of-sale/.impeccable/reference/japanese-high-density-web.webp'],
	['hero', '/home/real/Projects/lite-point-of-sale/.impeccable/reference/japanese-high-density-web-hero.webp']
];

const browser = await chromium.launch();
const page = await browser.newPage();

for (const [label, path] of FILES) {
	const dataUrl = 'data:image/webp;base64,' + readFileSync(path).toString('base64');
	const out = await page.evaluate(async (src) => {
		const img = new Image();
		img.src = src;
		await img.decode();
		const w = img.naturalWidth,
			h = img.naturalHeight;
		const c = document.createElement('canvas');
		c.width = w;
		c.height = h;
		const ctx = c.getContext('2d', { willReadFrequently: true });
		ctx.drawImage(img, 0, 0);
		const d = ctx.getImageData(0, 0, w, h).data;

		const lum = (i) => 0.2126 * d[i] + 0.7152 * d[i + 1] + 0.0722 * d[i + 2];
		const hist = new Map();
		let white = 0,
			ink = 0,
			red = 0,
			n = 0;
		const rowDark = new Float64Array(h);
		const colDark = new Float64Array(w);

		for (let y = 0; y < h; y++) {
			for (let x = 0; x < w; x++) {
				const i = (y * w + x) * 4;
				const L = lum(i);
				n++;
				if (d[i] >= 245 && d[i + 1] >= 245 && d[i + 2] >= 245) white++;
				if (L <= 60) ink++;
				if (d[i] > 140 && d[i + 1] < 100 && d[i + 2] < 100) red++;
				if (L < 200) {
					rowDark[y]++;
					colDark[x]++;
				}
				const k = `${d[i] >> 4},${d[i + 1] >> 4},${d[i + 2] >> 4}`;
				hist.set(k, (hist.get(k) ?? 0) + 1);
			}
		}

		// Garis rambut: baris/kolom yang mayoritas gelap sepanjang dimensinya.
		const lines = (arr, len) => {
			const hits = [];
			for (let i = 0; i < arr.length; i++) if (arr[i] / len > 0.6) hits.push(i);
			const groups = [];
			for (const i of hits) {
				const last = groups[groups.length - 1];
				if (last && i - last[last.length - 1] <= 2) last.push(i);
				else groups.push([i]);
			}
			return groups.map((g) => Math.round(g.reduce((a, b) => a + b, 0) / g.length));
		};
		const hLines = lines(rowDark, w);
		const vLines = lines(colDark, h);
		const gaps = (a) => {
			const g = [];
			for (let i = 1; i < a.length; i++) g.push(a[i] - a[i - 1]);
			g.sort((x, y) => x - y);
			return g.length ? { min: g[0], median: g[Math.floor(g.length / 2)], max: g[g.length - 1] } : null;
		};

		const top = [...hist.entries()]
			.sort((a, b) => b[1] - a[1])
			.slice(0, 10)
			.map(([k, v]) => {
				const [r, g, b] = k.split(',').map(Number);
				return {
					hex: '#' + [r, g, b].map((x) => ((x << 4) | 8).toString(16).padStart(2, '0')).join(''),
					share: +(100 * v / n).toFixed(1)
				};
			});

		return {
			size: `${w}×${h}`,
			whitePct: +(100 * white / n).toFixed(1),
			inkPct: +(100 * ink / n).toFixed(1),
			redPct: +(100 * red / n).toFixed(2),
			top,
			hLines: hLines.length,
			vLines: vLines.length,
			hPitch: gaps(hLines),
			vPitch: gaps(vLines),
			moduleEstimate: (hLines.length - 1) * (vLines.length - 1)
		};
	}, dataUrl);

	console.log(`\n=== ${label}: ${path.split('/').pop()} ===`);
	console.log(`  ukuran ${out.size} · putih ${out.whitePct}% · tinta ${out.inkPct}% · merah ${out.redPct}%`);
	console.log(`  garis rambut: ${out.hLines} mendatar, ${out.vLines} tegak`);
	console.log(`  jarak garis mendatar: ${JSON.stringify(out.hPitch)}`);
	console.log(`  jarak garis tegak: ${JSON.stringify(out.vPitch)}`);
	console.log(`  perkiraan jumlah modul: ${out.moduleEstimate}`);
	console.log('  warna teratas:');
	for (const t of out.top) console.log(`    ${t.hex}  ${String(t.share).padStart(5)}%`);
}

await browser.close();
