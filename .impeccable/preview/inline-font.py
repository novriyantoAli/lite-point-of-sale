#!/usr/bin/env python3
"""Semai Inter Variable ke style.css sebagai data URI.

Pratinjau harus bisa dibuka lewat file:// tanpa server dan tanpa jaringan, dan
Chrome tidak selalu memuat font lewat url() relatif dari file://. Menyematkan
berkasnya menghapus pertanyaan itu sekaligus membuat pratinjau ini satu berkas.

Sumbernya berkas yang sama dengan yang dipakai aplikasi:
frontend/node_modules/@fontsource-variable/inter/files/inter-latin-wght-normal.woff2
"""
import base64
import re
from pathlib import Path

HERE = Path(__file__).resolve().parent
FONT = HERE.parents[1] / "frontend/node_modules/@fontsource-variable/inter/files/inter-latin-wght-normal.woff2"
CSS = HERE / "style.css"
PLACEHOLDER = "__FONT_DATA_URI__"

if not FONT.exists():
    raise SystemExit(f"font not found: {FONT}\nJalankan `pnpm install` di frontend/ dulu.")

css = CSS.read_text(encoding="utf-8")

if PLACEHOLDER not in css:
    if "data:font/woff2;base64," in css:
        print("sudah disematkan; tidak ada yang diubah")
        raise SystemExit(0)
    raise SystemExit(f"placeholder {PLACEHOLDER} tidak ditemukan di {CSS}")

uri = "data:font/woff2;base64," + base64.b64encode(FONT.read_bytes()).decode("ascii")
CSS.write_text(css.replace(PLACEHOLDER, uri), encoding="utf-8")
print(f"disematkan {FONT.name} ({FONT.stat().st_size // 1024} KB) → {CSS} ({CSS.stat().st_size // 1024} KB)")
