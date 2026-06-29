#!/usr/bin/env python3
"""One-shot: split parse_text.go into text_types.go + parse_text.go (slim) + postscript.go.

Slices are by line range (1-indexed inclusive). Imports per output file are
auto-detected by scanning the extracted body for known tokens.

Run from repo root: python scripts/split_parse_text.py
"""
from __future__ import annotations
import pathlib
import re
import sys

SRC = pathlib.Path("internal/aep/parse_text.go")
DST_DIR = pathlib.Path("internal/aep")

# Each entry: filename → list of (start_line_1based, end_line_1based_inclusive).
# Lines 1-40 (package + imports + btdk header doc) stay in parse_text.go.
# Lines 41-360 = enums + types (text_types.go).
# Lines 361-675 = decoders, stay in parse_text.go.
# Lines 676-1016 = PS mini-parser (postscript.go).
SLICES = {
    "text_types.go": [(41, 360)],
    "postscript.go": [(676, 1016)],
}

IMPORT_RULES = [
    ("encoding/binary", re.compile(r"\bbinary\.")),
    ("fmt", re.compile(r"\bfmt\.")),
    ("math", re.compile(r"\bmath\.")),
    ("strings", re.compile(r"\bstrings\.")),
    ("unicode/utf16", re.compile(r"\butf16\.")),
    ("unicode/utf8", re.compile(r"\butf8\.")),
    (
        "github.com/yueli-fx/aep-parser/internal/rifx",
        re.compile(r"\brifx\."),
    ),
]


def detect_imports(body: str) -> list[str]:
    return [name for name, rx in IMPORT_RULES if rx.search(body)]


def render_imports(imps: list[str]) -> str:
    if not imps:
        return ""
    std = [i for i in imps if not i.startswith("github.com/")]
    third = [i for i in imps if i.startswith("github.com/")]
    out = ["import ("]
    for s in std:
        out.append(f'\t"{s}"')
    if third:
        if std:
            out.append("")
        for t in third:
            out.append(f'\t"{t}"')
    out.append(")")
    return "\n".join(out)


def trim_trailing_blanks(lines: list[str]) -> list[str]:
    while lines and lines[-1].strip() == "":
        lines.pop()
    return lines


def collapse_blanks(lines: list[str]) -> list[str]:
    out, blank = [], 0
    for ln in lines:
        if ln.strip() == "":
            blank += 1
            if blank > 2:
                continue
        else:
            blank = 0
        out.append(ln)
    return out


def main():
    text = SRC.read_text(encoding="utf-8")
    lines = text.split("\n")
    n = len(lines)

    drop: set[int] = set()
    summary = []

    for fname, ranges in SLICES.items():
        body_lines: list[str] = []
        for s, e in ranges:
            # 1-indexed inclusive → 0-indexed slice
            body_lines.extend(lines[s - 1 : e])
            drop.update(range(s - 1, e))
        body_lines = trim_trailing_blanks(body_lines)
        body = "\n".join(body_lines)
        imps = detect_imports(body)
        parts = ["package aep", ""]
        imp_block = render_imports(imps)
        if imp_block:
            parts.append(imp_block)
            parts.append("")
        parts.append(body)
        parts.append("")  # trailing newline
        out_path = DST_DIR / fname
        out_path.write_text("\n".join(parts), encoding="utf-8")
        summary.append((fname, len(body_lines), imps))

    # Rewrite parse_text.go without the moved blocks.
    kept = [ln for i, ln in enumerate(lines) if i not in drop]
    kept = collapse_blanks(kept)
    kept = trim_trailing_blanks(kept)
    kept.append("")  # final newline
    SRC.write_text("\n".join(kept), encoding="utf-8")

    new_len = len(kept) - 1  # excluding trailing empty
    print(f"parse_text.go: {n} → {new_len} lines (-{n - new_len})")
    for fname, nl, imps in summary:
        imp_str = ", ".join(imps) if imps else "(no imports)"
        print(f"  {fname}: {nl} lines [{imp_str}]")


if __name__ == "__main__":
    main()
