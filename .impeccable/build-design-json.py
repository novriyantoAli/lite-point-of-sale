#!/usr/bin/env python3
"""Generate .impeccable/design.json from DESIGN.md — dunia "The Dense Mosaic".

Narasinya diekstrak dari DESIGN.md supaya keduanya tidak bisa berbeda.
Token primitif ada di frontmatter DESIGN.md; sidecar ini hanya membawa yang
tidak muat di sana: tonal ramp, display name, shadow/motion/breakpoint, dan
komponen sebagai HTML/CSS yang bisa langsung dirender.
"""
import json
import re
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent  # akar repo
md = (ROOT / "DESIGN.md").read_text(encoding="utf-8")
lines = md.split("\n")

# Dasar netral berkroma nol; merah utilitas punya hue sendiri.
NEUTRAL_RAMP = [f"oklch({l}% 0 0)" for l in (15, 27, 39, 51, 63, 75, 87, 96)]
UTILITY_RAMP = [
    "oklch(15% 0.05 27)",
    "oklch(30% 0.12 27)",
    "oklch(43% 0.18 27)",
    "oklch(0.55 0.22 27)",
    "oklch(68% 0.19 27)",
    "oklch(79% 0.13 27)",
    "oklch(90% 0.06 27)",
    "oklch(96% 0.02 27)",
]

COLORS = [
    ("ground", "neutral", "Ground Grey", "#fafafa", NEUTRAL_RAMP),
    ("tile", "primary", "Tile White", "#ffffff", NEUTRAL_RAMP),
    ("hairline", "neutral", "Hairline Grey", "#e8e8e8", NEUTRAL_RAMP),
    ("wash", "neutral", "Wash Grey", "#f5f5f5", NEUTRAL_RAMP),
    ("ink", "primary", "Ink Black", "#000000", NEUTRAL_RAMP),
    ("utility", "secondary", "Utility Red", "#cc0d0d", UTILITY_RAMP),
]

TYPO_META = {
    "title": "Satu per layar. Satu-satunya tempat 24px muncul, dan lompatan skala yang membuat hierarki terbaca.",
    "figure": "Angka yang jadi jawaban layar: Kembalian, jumlah dibayar. Selalu tabular-nums.",
    "strong": "Total berjalan, nilai di kepala kolom, teks tombol utama.",
    "body": "Badan seluruh antarmuka: nama Produk, isi tabel, input, teks tombol. Ukuran kerja sistem ini.",
    "micro": "Tab, badge, kepala tabel, label field. Ini lantainya; tidak ada teks di bawah 12px.",
}

SHADOWS = []  # dunia ini tidak punya bayangan sama sekali

MOTION = [
    {"name": "duration-swap", "value": "100ms", "purpose": "Satu-satunya transisi: latar dan garis berubah saat hover dan fokus."},
    {"name": "ease-default", "value": "cubic-bezier(0.4, 0, 0.2, 1)", "purpose": "Satu-satunya easing."},
    {"name": "reduced-motion", "value": "0ms", "purpose": "prefers-reduced-motion mematikan transisi; tukar keadaan tetap seketika."},
]

BREAKPOINTS = [
    {"name": "stack", "value": "1080px", "purpose": "Di bawah ini papan menumpuk jadi satu kolom, tidak pernah menggeser mendatar."},
    {"name": "narrow", "value": "560px", "purpose": "Rel membungkus dan mosaik turun ke minimum 132px per petak."},
]

BTN = "display: inline-flex; align-items: center; justify-content: center; gap: 6px; font-family: 'Inter Variable', system-ui, sans-serif; border: 1px solid var(--hairline, #e8e8e8); border-radius: 0; cursor: pointer; white-space: nowrap; transition: background 100ms cubic-bezier(0.4, 0, 0.2, 1), border-color 100ms cubic-bezier(0.4, 0, 0.2, 1);"

COMPONENTS = [
    {
        "name": "Module",
        "kind": "card",
        "refersTo": "module",
        "description": "Blok berisi terkecil: petak putih dengan satu garis rambut. Kepalanya memakai isian wash dan ditutup garis tinta.",
        "html": '<div class="ds-mod"><div class="ds-mod-head"><span class="ds-mod-title">Keranjang</span><span class="ds-mod-note">7 unit · 2 Item</span></div><div class="ds-mod-body">Modul isi.</div></div>',
        "css": (
            ".ds-mod { background: var(--tile, #ffffff); border: 1px solid var(--hairline, #e8e8e8); border-bottom: 0; font-family: 'Inter Variable', system-ui, sans-serif; color: var(--ink, #000000); }\n"
            ".ds-mod-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 6px 8px; background: var(--wash, #f5f5f5); border-bottom: 1px solid var(--ink, #000000); }\n"
            ".ds-mod-title { font-size: 0.8125rem; font-weight: 700; }\n"
            ".ds-mod-note { font-size: 0.75rem; font-variant-numeric: tabular-nums; }\n"
            ".ds-mod-body { padding: 6px 8px; font-size: 0.8125rem; }\n"
            ".ds-mod > .ds-mod-body:last-child { border-bottom: 1px solid var(--hairline, #e8e8e8); }"
        ),
    },
    {
        "name": "Product Tile",
        "kind": "button",
        "refersTo": "tile",
        "description": "Unit terkecil yang bisa ditekan. Tab Kode di kiri atas, nama, lalu Stok di kiri dan harga merah di kanan. Hover menaikkan garisnya jadi tinta, tanpa bayangan.",
        "html": '<button class="ds-tile" type="button"><span class="ds-tag">NYM325</span><span class="ds-tile-name">Kabel NYM 3x2.5 mm</span><span class="ds-tile-foot"><span class="ds-tile-stock">Stok 120</span><span class="ds-tile-price">Rp 18.500</span></span></button>',
        "css": (
            ".ds-tile { position: relative; display: flex; flex-direction: column; gap: 4px; width: 154px; padding: 6px 8px 8px; text-align: left; background: var(--tile, #ffffff); border: 1px solid var(--hairline, #e8e8e8); border-radius: 0; cursor: pointer; font-family: 'Inter Variable', system-ui, sans-serif; color: var(--ink, #000000); transition: background 100ms cubic-bezier(0.4, 0, 0.2, 1), border-color 100ms cubic-bezier(0.4, 0, 0.2, 1); }\n"
            ".ds-tile:hover { background: var(--wash, #f5f5f5); border-color: var(--ink, #000000); }\n"
            ".ds-tile:focus-visible { outline: 2px solid var(--ink, #000000); outline-offset: -2px; }\n"
            ".ds-tag { align-self: flex-start; display: inline-flex; align-items: center; height: 15px; padding: 0 5px; font-size: 0.75rem; font-weight: 600; color: var(--tile, #ffffff); background: var(--utility, #cc0d0d); }\n"
            ".ds-tile-name { font-size: 0.8125rem; font-weight: 500; line-height: 1.2; }\n"
            ".ds-tile-foot { display: flex; align-items: baseline; justify-content: space-between; gap: 8px; margin-top: auto; }\n"
            ".ds-tile-stock { font-size: 0.75rem; font-variant-numeric: tabular-nums; }\n"
            ".ds-tile-price { font-size: 0.8125rem; font-weight: 600; color: var(--utility, #cc0d0d); font-variant-numeric: tabular-nums; }"
        ),
    },
    {
        "name": "Product Tile — Stok habis",
        "kind": "button",
        "refersTo": "tile-disabled",
        "description": "Produk yang habis kehilangan tintanya: nama dan harga dicoret garis, petaknya mati. Bukan dipudarkan.",
        "html": '<button class="ds-tile ds-tile--out" type="button" disabled><span class="ds-tag ds-tag--quiet">tanpa Kode</span><span class="ds-tile-name">Gergaji Besi</span><span class="ds-tile-foot"><span class="ds-tile-stock">Stok habis</span><span class="ds-tile-price">Rp 25.000</span></span></button>',
        "css": (
            ".ds-tile--out { cursor: not-allowed; }\n"
            ".ds-tile--out:hover { background: var(--tile, #ffffff); border-color: var(--hairline, #e8e8e8); }\n"
            ".ds-tile--out .ds-tile-name, .ds-tile--out .ds-tile-price { text-decoration: line-through; text-decoration-thickness: 1px; }\n"
            ".ds-tag--quiet { color: var(--ink, #000000); background: var(--wash, #f5f5f5); border: 1px solid var(--hairline, #e8e8e8); }"
        ),
    },
    {
        "name": "Nav Tab",
        "kind": "nav",
        "refersTo": "tab",
        "description": "Navigasi adalah tab, bukan daftar. Tab aktif memakai isian merah penuh dengan teks putih dan cacah tabular di sebelah namanya.",
        "html": '<nav class="ds-tabs"><a class="ds-tab" href="#">Beranda</a><a class="ds-tab ds-tab--on" href="#" aria-current="page">Kasir</a><a class="ds-tab" href="#">Produk <span class="ds-tab-n">24</span></a></nav>',
        "css": (
            ".ds-tabs { display: flex; align-items: stretch; font-family: 'Inter Variable', system-ui, sans-serif; }\n"
            ".ds-tab { display: inline-flex; align-items: center; gap: 6px; padding: 6px 10px; font-size: 0.8125rem; color: var(--ink, #000000); background: var(--tile, #ffffff); border-right: 1px solid var(--hairline, #e8e8e8); text-decoration: none; white-space: nowrap; transition: background 100ms cubic-bezier(0.4, 0, 0.2, 1); }\n"
            ".ds-tab:hover { background: var(--wash, #f5f5f5); }\n"
            ".ds-tab--on, .ds-tab--on:hover { background: var(--utility, #cc0d0d); color: var(--tile, #ffffff); font-weight: 600; }\n"
            ".ds-tab-n { font-size: 0.75rem; font-variant-numeric: tabular-nums; }"
        ),
    },
    {
        "name": "Input Field",
        "kind": "input",
        "refersTo": "input",
        "description": "Field 26px persegi, label 12px di atasnya. Fokus memakai garis tinta plus outline ke dalam, bukan glow. Field uang memakai caret merah.",
        "html": '<div class="ds-field"><label class="ds-label" for="ds-bayar">Jumlah bayar</label><input id="ds-bayar" class="ds-input" value="150000" /></div>',
        "css": (
            ".ds-field { display: flex; flex-direction: column; gap: 3px; font-family: 'Inter Variable', system-ui, sans-serif; }\n"
            ".ds-label { font-size: 0.75rem; font-weight: 600; color: var(--ink, #000000); }\n"
            ".ds-input { height: 26px; width: 100%; padding: 0 6px; font-family: inherit; font-size: 0.8125rem; font-variant-numeric: tabular-nums; color: var(--ink, #000000); background: var(--tile, #ffffff); border: 1px solid var(--hairline, #e8e8e8); border-radius: 0; outline: none; caret-color: var(--utility, #cc0d0d); transition: border-color 100ms cubic-bezier(0.4, 0, 0.2, 1); }\n"
            ".ds-input:hover { border-color: var(--ink, #000000); }\n"
            ".ds-input:focus-visible { border-color: var(--ink, #000000); outline: 2px solid var(--ink, #000000); outline-offset: -2px; }\n"
            ".ds-input[aria-invalid='true'] { border-color: var(--utility, #cc0d0d); outline: 2px solid var(--utility, #cc0d0d); outline-offset: -2px; }"
        ),
    },
    {
        "name": "Payment Method Cell",
        "kind": "input",
        "refersTo": "method",
        "description": "Empat metode sebagai radio asli yang disembunyikan, dengan label sebagai sel persegi. Terpilih berarti bidang tinta penuh, bukan tint dan bukan centang.",
        "html": '<div class="ds-methods"><label class="ds-method"><input type="radio" name="ds-m" checked />Tunai</label><label class="ds-method"><input type="radio" name="ds-m" />QRIS</label><label class="ds-method"><input type="radio" name="ds-m" />Debit</label><label class="ds-method"><input type="radio" name="ds-m" />Transfer</label></div>',
        "css": (
            ".ds-methods { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 4px; font-family: 'Inter Variable', system-ui, sans-serif; }\n"
            ".ds-method { display: flex; align-items: center; justify-content: center; height: 26px; font-size: 0.8125rem; font-weight: 500; color: var(--ink, #000000); background: var(--tile, #ffffff); border: 1px solid var(--hairline, #e8e8e8); border-radius: 0; cursor: pointer; transition: background 100ms cubic-bezier(0.4, 0, 0.2, 1), border-color 100ms cubic-bezier(0.4, 0, 0.2, 1); }\n"
            ".ds-method:hover { background: var(--wash, #f5f5f5); }\n"
            ".ds-method input { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0, 0, 0, 0); }\n"
            ".ds-method:focus-within { outline: 2px solid var(--ink, #000000); outline-offset: -2px; }\n"
            ".ds-method:has(input:checked) { background: var(--ink, #000000); color: var(--tile, #ffffff); border-color: var(--ink, #000000); font-weight: 600; }"
        ),
    },
    {
        "name": "Commit Button",
        "kind": "button",
        "refersTo": "button-commit",
        "description": "Satu-satunya bidang bertinta penuh di layar. Ia tidak diangkat dengan bayangan, ia diangkat dengan tinta.",
        "html": '<button class="ds-commit" type="button">Bayar &amp; Simpan Penjualan</button>',
        "css": (
            ".ds-commit { " + BTN + " width: 100%; height: 40px; font-size: 0.9375rem; font-weight: 600; color: var(--tile, #ffffff); background: var(--ink, #000000); border-color: var(--ink, #000000); }\n"
            ".ds-commit:focus-visible { outline: 2px solid var(--ink, #000000); outline-offset: 2px; }"
        ),
    },
    {
        "name": "Commit Button — disabled",
        "kind": "button",
        "refersTo": "button-commit-disabled",
        "description": "Keadaan mati bukan warna pudar: tombolnya kehilangan tinta dan berubah jadi putih bergaris putus-putus.",
        "html": '<button class="ds-commit ds-commit--off" type="button" disabled>Bayar &amp; Simpan Penjualan</button>',
        "css": (
            ".ds-commit--off { color: var(--ink, #000000); background: var(--tile, #ffffff); border-color: var(--ink, #000000); border-style: dashed; cursor: not-allowed; }\n"
            ".ds-commit--off:hover { background: var(--tile, #ffffff); }"
        ),
    },
    {
        "name": "Figure",
        "kind": "custom",
        "refersTo": "figure",
        "description": "Angka yang jadi jawaban layar, selalu didahului label yang menyebut apa angka itu. Angka yang belum bisa dihitung ditulis —.",
        "html": '<div class="ds-fig"><p class="ds-fig-label">Kembalian</p><p class="ds-fig-value">Rp 13.500</p></div>',
        "css": (
            ".ds-fig { font-family: 'Inter Variable', system-ui, sans-serif; }\n"
            ".ds-fig-label { margin: 0; font-size: 0.75rem; font-weight: 600; color: var(--ink, #000000); }\n"
            ".ds-fig-value { margin: 0; font-size: 1.25rem; font-weight: 700; line-height: 1.1; font-variant-numeric: tabular-nums; color: var(--ink, #000000); }"
        ),
    },
    {
        "name": "Shared Hairline Mosaic",
        "kind": "custom",
        "refersTo": "tile",
        "description": "Dua petak bersebelahan berbagi satu garis 1px, bukan dua garis yang menempel. Inilah yang membuat mosaiknya padat tanpa celah.",
        "html": '<div class="ds-mosaic"><span class="ds-cell">Sekring 2A</span><span class="ds-cell">Tespen</span><span class="ds-cell">Fitting Lampu</span><span class="ds-cell">Meteran 5 m</span></div>',
        "css": (
            ".ds-mosaic { display: grid; grid-template-columns: repeat(2, 154px); border-left: 1px solid var(--hairline, #e8e8e8); font-family: 'Inter Variable', system-ui, sans-serif; }\n"
            ".ds-cell { padding: 6px 8px; font-size: 0.8125rem; color: var(--ink, #000000); background: var(--tile, #ffffff); border: 1px solid var(--hairline, #e8e8e8); margin: -1px 0 0 -1px; }"
        ),
    },
]

# ---- narasi, diekstrak dari DESIGN.md ---------------------------------------
north_star = ""
overview_paras = []
chars = []
rules = []
dos = []
donts = []

SECTION_ALIASES = {"elevation & depth": "elevation"}
section = ""
mode = None
rule_body = None


def flush_rule():
    global rule_body
    if rule_body:
        rules.append({"name": rule_body[0], "body": " ".join(rule_body[1]).strip(), "section": rule_body[2]})
    rule_body = None


for line in lines:
    if line.startswith("## "):
        flush_rule()
        section = line[3:].strip().lower()
        mode = None
    if line.startswith("### "):
        flush_rule()
        head = line[4:].strip()
        if head.startswith("Don"):
            mode = "donts"
        elif head.startswith("Do"):
            mode = "dos"
        else:
            mode = None

    m = re.match(r'^\*\*Creative North Star: "(.+?)"\*\*$', line.strip())
    if m:
        north_star = m.group(1)

    if mode == "dos" and line.startswith("- **Do** "):
        dos.append(line[len("- **Do** ") :].strip())
    if mode == "donts" and line.startswith("- **Don't** "):
        donts.append(line[len("- **Don't** ") :].strip())

    if mode is None and section == "overview" and line.startswith("- **"):
        chars.append(re.sub(r"^- \*\*(.+?)\*\* ", r"\1 ", line).strip())

    if rule_body is None:
        m = re.match(r"^\*\*(The .+? Rule)\.\*\* (.*)$", line.strip())
        if m:
            rule_body = [m.group(1), [m.group(2)], SECTION_ALIASES.get(section, section)]
    else:
        if line.strip() == "" or line.startswith("#"):
            flush_rule()
        elif not line.startswith("- **"):
            rule_body[1].append(line.strip())
flush_rule()

start = next(i for i, l in enumerate(lines) if "Creative North Star" in l) + 1
end = next(i for i, l in enumerate(lines) if l.startswith("**Key Characteristics:**"))
para = []
for l in lines[start:end]:
    if l.strip() == "":
        if para:
            overview_paras.append(" ".join(para))
            para = []
    else:
        para.append(l.strip())
if para:
    overview_paras.append(" ".join(para))

out = {
    "schemaVersion": 2,
    "generatedAt": datetime.now(timezone.utc).astimezone().isoformat(timespec="seconds"),
    "title": "Design System: Lite Point of Sale",
    "extensions": {
        "colorMeta": {
            key: {"role": role, "displayName": name, "canonical": canon, "tonalRamp": ramp}
            for key, role, name, canon, ramp in COLORS
        },
        "typographyMeta": {k: {"displayName": k.capitalize(), "purpose": v} for k, v in TYPO_META.items()},
        "shadows": SHADOWS,
        "motion": MOTION,
        "breakpoints": BREAKPOINTS,
    },
    "components": COMPONENTS,
    "narrative": {
        "northStar": north_star,
        "overview": "\n\n".join(overview_paras),
        "keyCharacteristics": chars,
        "rules": rules,
        "dos": dos,
        "donts": donts,
    },
}

target = ROOT / ".impeccable" / "design.json"
target.write_text(json.dumps(out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"wrote {target}")
print(f"colors={len(COLORS)} components={len(COMPONENTS)} rules={len(rules)} chars={len(chars)} dos={len(dos)} donts={len(donts)}")
print(f"northStar={north_star!r} overview_paras={len(overview_paras)}")
print("rules:", [r["name"] + "/" + r["section"] for r in rules])
