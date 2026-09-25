import { chromium } from '@playwright/test';
import { readFileSync } from 'node:fs';
const FILES = [
  ['board', '/home/real/Projects/lite-point-of-sale/.impeccable/reference/japanese-high-density-web.webp'],
  ['hero',  '/home/real/Projects/lite-point-of-sale/.impeccable/reference/japanese-high-density-web-hero.webp']
];
const browser = await chromium.launch();
const page = await browser.newPage();
for (const [label, path] of FILES) {
  const out = await page.evaluate(async (src) => {
    const img = new Image(); img.src = src; await img.decode();
    const c = document.createElement('canvas'); c.width = img.naturalWidth; c.height = img.naturalHeight;
    const ctx = c.getContext('2d', { willReadFrequently: true }); ctx.drawImage(img, 0, 0);
    const d = ctx.getImageData(0, 0, c.width, c.height).data;
    const hist = new Map(); const greys = new Map();
    let n = 0, redSum = [0,0,0], redN = 0;
    for (let i = 0; i < d.length; i += 4) {
      const r = d[i], g = d[i+1], b = d[i+2]; n++;
      const k = (r << 16) | (g << 8) | b;
      hist.set(k, (hist.get(k) ?? 0) + 1);
      if (r === g && g === b) greys.set(r, (greys.get(r) ?? 0) + 1);
      if (r > 140 && g < 100 && b < 100) { redSum[0]+=r; redSum[1]+=g; redSum[2]+=b; redN++; }
    }
    const hex = (k) => '#' + k.toString(16).padStart(6, '0');
    return {
      n,
      top: [...hist.entries()].sort((a,b)=>b[1]-a[1]).slice(0,14).map(([k,v]) => ({ hex: hex(k), pct: +(100*v/n).toFixed(2) })),
      greyRamp: [...greys.entries()].sort((a,b)=>b[1]-a[1]).slice(0,10).map(([v,c]) => ({ hex: '#'+v.toString(16).padStart(2,'0').repeat(3), pct: +(100*c/n).toFixed(2) })),
      redMean: redN ? '#' + redSum.map(v => Math.round(v/redN).toString(16).padStart(2,'0')).join('') : null,
      redShare: +(100*redN/n).toFixed(2)
    };
  }, 'data:image/webp;base64,' + readFileSync(path).toString('base64'));
  console.log(`\n=== ${label} ===`);
  console.log('  warna persis teratas:');
  for (const t of out.top) console.log(`    ${t.hex}  ${String(t.pct).padStart(5)}%`);
  console.log('  ramp abu-abu:', out.greyRamp.map(g => `${g.hex}(${g.pct}%)`).join(' '));
  console.log(`  merah: rata-rata ${out.redMean}, ${out.redShare}% dari gambar`);
}
await browser.close();
