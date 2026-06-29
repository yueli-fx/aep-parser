#!/usr/bin/env python3
"""One-shot: split write.go into write_keyframe.go + write_property.go,
and append Layer.SetText (+ its text-encoding helpers) into write_layer.go.

Same pattern as split_tests.py / split_parse_text.py: walk top-level func
headers, slice by name, auto-detect imports.

For NEW output files, the import block is fully auto-detected.
For APPEND mode (write_layer.go), the existing import block is parsed,
merged with newly detected imports, and re-emitted at the top.

For the leftover write.go we strip the original import block and re-detect
to prune now-unused imports (e.g., `bytes` after SetText leaves).

Run from repo root: python scripts/split_write.py
"""
from __future__ import annotations
import pathlib
import re
import sys

SRC = pathlib.Path("internal/aep/write.go")
DST_DIR = pathlib.Path("internal/aep")

TARGETS = {
    "write_keyframe.go": {
        "mode": "new",
        "header": (
            "// Keyframe writers. All edits are length-preserving block writes\n"
            "// into the parent ldat. Layout dispatch (spatial vs non-spatial)\n"
            "// is owned by the Keyframe value itself.\n"
        ),
        "funcs": [
            "SetTime",
            "SetValue",
            "SetInInterp",
            "SetOutInterp",
            "setInterpByte",
            "SetInTemporalEase",
            "SetOutTemporalEase",
            "setTemporalEase",
            "SetInSpatialTangent",
            "SetOutSpatialTangent",
            "setSpatialTangent",
            "blockSlice",
            "blockSize",
        ],
    },
    "write_property.go": {
        "mode": "new",
        "header": (
            "// Property-level writers: static value, expression source &\n"
            "// enabled bit, and keyframe insert/delete (length-variable ldat\n"
            "// rebuild that re-parses the property's keyframe stream).\n"
        ),
        "funcs": [
            "SetStaticValue",
            "SetExpressionEnabled",
            "InsertKeyframe",
            "DeleteKeyframe",
            "reparseKeyframes",
            "SetExpression",
        ],
    },
    "write_layer.go": {
        "mode": "append",
        "funcs": [
            "SetText",
            "TextEncodedByteLen",
            "encodeAEPSText",
            "writeUTF16BEEscaped",
            "writeBEEscaped",
        ],
    },
}

# Pattern `\bpkg\.\w` rejects false positives in comments like "ldat bytes. Used"
# while still matching real package usage like `bytes.Buffer`.
IMPORT_RULES = [
    ("bytes", re.compile(r"\bbytes\.\w")),
    ("encoding/binary", re.compile(r"\bbinary\.\w")),
    ("fmt", re.compile(r"\bfmt\.\w")),
    ("io", re.compile(r"\bio\.\w")),
    ("math", re.compile(r"\bmath\.\w")),
    ("path/filepath", re.compile(r"\bfilepath\.\w")),
    ("strings", re.compile(r"\bstrings\.\w")),
    ("unicode/utf16", re.compile(r"\butf16\.\w")),
    (
        "github.com/yueli-fx/aep-parser/internal/rifx",
        re.compile(r"\brifx\.\w"),
    ),
]

FUNC_HEADER = re.compile(
    r"^func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?\s+(\w+)\s*\(",
)


def parse_funcs(text: str):
    """Return (lines, [(name, start_inclusive_w_comments, end_exclusive)])."""
    lines = text.split("\n")
    funcs = []
    i = 0
    while i < len(lines):
        m = FUNC_HEADER.match(lines[i])
        if not m:
            i += 1
            continue
        name = m.group(1)
        cs = i
        while cs - 1 >= 0 and lines[cs - 1].startswith("//"):
            cs -= 1
        j = i + 1
        while j < len(lines):
            if lines[j] == "}":
                break
            j += 1
        funcs.append((name, cs, j + 1))
        i = j + 1
    return lines, funcs


def detect_imports(body: str) -> list[str]:
    return [name for name, rx in IMPORT_RULES if rx.search(body)]


def render_imports(imps: list[str]) -> str:
    if not imps:
        return ""
    std = sorted(i for i in imps if not i.startswith("github.com/"))
    third = sorted(i for i in imps if i.startswith("github.com/"))
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


def trim_trailing_blanks(lines: list[str]) -> list[str]:
    while lines and lines[-1].strip() == "":
        lines.pop()
    return lines


IMPORT_OPEN = re.compile(r"^import\s*\(")


def split_off_import_block(lines: list[str]) -> tuple[list[str], list[str]]:
    """Return (lines_without_import_block, parsed_import_paths)."""
    n = len(lines)
    for i, ln in enumerate(lines):
        if IMPORT_OPEN.match(ln):
            j = i + 1
            while j < n and lines[j] != ")":
                j += 1
            # parse the paths between (i+1, j)
            paths = []
            for k in range(i + 1, j):
                m = re.match(r"\s*\"([^\"]+)\"", lines[k])
                if m:
                    paths.append(m.group(1))
            # drop lines [i, j] inclusive; also one trailing blank if present
            end = j + 1
            if end < n and lines[end].strip() == "":
                end += 1
            return lines[:i] + lines[end:], paths
    return lines, []


def main():
    text = SRC.read_text(encoding="utf-8")
    lines, funcs = parse_funcs(text)
    name_to_range = {n: (cs, ce) for n, cs, ce in funcs}

    drop: set[int] = set()
    summary = []

    for fname, cfg in TARGETS.items():
        chunks = []
        for n in cfg["funcs"]:
            if n not in name_to_range:
                print(f"WARN: {n} not found in write.go", file=sys.stderr)
                continue
            cs, ce = name_to_range[n]
            chunks.append("\n".join(lines[cs:ce]))
            drop.update(range(cs, ce))
        body = "\n\n".join(chunks)
        new_imps = detect_imports(body)

        out_path = DST_DIR / fname
        if cfg["mode"] == "new":
            header = cfg.get("header", "")
            parts = ["package aep", ""]
            imp = render_imports(new_imps)
            if imp:
                parts.append(imp)
                parts.append("")
            if header:
                parts.append(header.rstrip() + "\n")
            parts.append(body)
            parts.append("")
            out_path.write_text("\n".join(parts), encoding="utf-8")
            summary.append((fname, "new", len(body.split("\n")), new_imps))
        else:
            # append: read existing, strip its import block, merge imports, rewrite
            existing = out_path.read_text(encoding="utf-8").split("\n")
            # drop package line — we re-emit it
            assert existing[0].startswith("package "), "unexpected first line"
            pkg_line = existing[0]
            stripped, existing_imps = split_off_import_block(existing[1:])
            stripped = trim_trailing_blanks(stripped)
            merged = sorted(set(existing_imps) | set(new_imps))
            std = [i for i in merged if not i.startswith("github.com/")]
            third = [i for i in merged if i.startswith("github.com/")]
            merged_ordered = std + third
            imp = render_imports(merged_ordered)
            parts = [pkg_line, ""]
            if imp:
                parts.append(imp)
                parts.append("")
            parts.append("\n".join(stripped).lstrip("\n"))
            parts.append("")
            parts.append(body)
            parts.append("")
            out_path.write_text("\n".join(parts), encoding="utf-8")
            summary.append((fname, "append", len(body.split("\n")), new_imps))

    # rewrite write.go: keep non-dropped lines, strip its original import
    # block, re-detect imports from kept body, re-emit.
    kept = [ln for i, ln in enumerate(lines) if i not in drop]
    assert kept[0].startswith("package "), "kept[0] not package?"
    pkg_line = kept[0]
    rest, _ = split_off_import_block(kept[1:])
    rest = trim_trailing_blanks(rest)
    rest = collapse_blanks(rest)
    kept_body = "\n".join(rest)
    kept_imps = detect_imports(kept_body)
    imp = render_imports(kept_imps)
    parts = [pkg_line, ""]
    if imp:
        parts.append(imp)
        parts.append("")
    parts.append(kept_body.lstrip("\n"))
    parts.append("")
    SRC.write_text("\n".join(parts), encoding="utf-8")

    new_len = len(SRC.read_text(encoding="utf-8").split("\n")) - 1
    print(f"write.go: {len(lines)} → {new_len} lines (-{len(lines) - new_len})")
    for fname, mode, nl, imps in summary:
        print(f"  [{mode}] {fname}: +{nl} lines, imports={imps}")


if __name__ == "__main__":
    main()
