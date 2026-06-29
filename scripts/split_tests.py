#!/usr/bin/env python3
"""One-shot: extract test functions from aep_test.go into domain files.

Mapping is hardcoded below. Imports are detected by scanning extracted text
for token use. Each func's leading // comment block is taken with it.

Run from repo root: python scripts/split_tests.py
"""
from __future__ import annotations
import pathlib
import re
import sys

SRC = pathlib.Path("internal/aep/aep_test.go")
DST_DIR = pathlib.Path("internal/aep")

# Test → target file. Helpers (non-Test funcs) are kept in aep_test.go.
# expression_test.go takes 1 test + 3 helpers (buildEffectParade method /
# buildTextSourceWrapper method / buildExtendedAEP free func) since they're
# only used by that one test now (property_test imports it too — but same
# package, just stays accessible).
TARGETS = {
    "expression_test.go": [
        "buildEffectParade",
        "buildTextSourceWrapper",
        "buildExtendedAEP",
        "TestExpressionAndEffectsAndMarkersAndText",
    ],
    "composition_test.go": [
        "buildAEPWithCustomCdta",
        "TestPerCompTickRate",
        "TestCompositionBGColorRegression",
        "TestCompositionCDTAFields",
        "TestCompositionCDTARealFixture",
        "TestCompositionLayerByID",
        "TestCompositionSetters",
        "TestCompositionNameAndTiming",
        "TestCompositionSettersRejectMissingCdta",
        "TestCompositionExtraFlagSetters",
        "TestPixelAspectReadReal",
        "TestCompositionActiveCamera",
        "TestCompositionFlagSetters",
        "TestDisplayStartTimeReal",
        "TestSetDisplayStartTimeRoundtrip",
    ],
    "layer_test.go": [
        "TestLayerFieldDecoding",
        "TestLayerComment",
        "TestLayerCommentEmpty",
        "TestLayerAutoOrient",
        "TestLayerParent",
        "TestLayerParentNoComp",
        "TestLayerSourceComposition",
        "TestLayerSourceCompositionNoOwner",
        "TestJSONParentName",
        "TestLayerPropertyAccessors",
        "TestLayerSetters",
        "TestLayerTimeSetters",
        "TestLayerSetParentAndSource",
        "TestLayerSetAutoOrient",
        "TestLayerSetNameAndComment",
        "TestLayerSettersRejectMissingLdta",
        "TestTrackMatteLayerReal",
        "TestSetTrackMatteLayerRoundtrip",
        "TestSetTrackMatteLayerRejectsInvalid",
        "TestLightKindReal",
        "TestSetLightKindRoundtrip",
        "TestTimeRemapEnabledReal",
        "TestCameraLightAccessorsReal",
        "TestAlternateSourceReal",
        "TestSetAlternateSourceRoundtrip",
        "TestSetAlternateSourceRejectsInvalid",
        "TestLayerAddFontAndUse",
    ],
    "text_test.go": [
        "TestTextSourceDecodeSynthetic",
        "TestTextSourceDecodeReal",
        "TestTextSourceExtendedFieldsReal",
        "TestSetTextLengthPreserving",
        "TestTextPerRunSetters",
        "TestTextPerRunSettersRejectOutOfRange",
        "TestTextCapsBaselineStrokeReal",
        "TestTextParagraphFieldsReal",
        "TestTextWave1Roundtrip",
        "TestTextWave1RejectInvalid",
        "TestTextAE24MoreRunFieldsReal",
        "TestTextAE24MoreParagraphFieldsReal",
        "TestTextAE24MoreSettersRoundtrip",
        "TestTextManualKerningReal",
        "TestSetManualKerningRoundtrip",
        "TestSetManualKerningRejectsInvalid",
        "TestFontAxesReal",
    ],
}

# Detect imports by scanning func bodies for these tokens.
IMPORT_RULES = [
    ("bytes", re.compile(r"\bbytes\.")),
    ("encoding/binary", re.compile(r"\bbinary\.")),
    ("errors", re.compile(r"\berrors\.")),
    ("flag", re.compile(r"\bflag\.")),
    ("fmt", re.compile(r"\bfmt\.")),
    ("io", re.compile(r"\bio\.")),
    ("io/fs", re.compile(r"\bfs\.")),
    ("math", re.compile(r"\bmath\.")),
    ("os", re.compile(r"\bos\.")),
    ("strings", re.compile(r"\bstrings\.")),
    ("testing", re.compile(r"\btesting\.")),
    ("aep", re.compile(r"\baep\.")),
]

FUNC_HEADER = re.compile(
    r"^func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?\s+(\w+)\s*\(",
)


def parse_funcs(text: str):
    """Return list of (name, start_line_inclusive_with_comments, end_line_exclusive)."""
    lines = text.split("\n")
    funcs = []
    i = 0
    while i < len(lines):
        m = FUNC_HEADER.match(lines[i])
        if not m:
            i += 1
            continue
        name = m.group(1)
        # walk back through // comment block (no blank gap)
        cs = i
        while cs - 1 >= 0 and lines[cs - 1].startswith("//"):
            cs -= 1
        # find closing brace
        j = i + 1
        while j < len(lines):
            if lines[j] == "}":
                break
            j += 1
        funcs.append((name, cs, j + 1))
        i = j + 1
    return lines, funcs


def gather_imports(body: str) -> list[str]:
    imps = []
    for name, rx in IMPORT_RULES:
        if rx.search(body):
            imps.append(name)
    return imps


def render_imports(imps: list[str]) -> str:
    std = [i for i in imps if i != "aep"]
    has_aep = "aep" in imps
    lines = ["import ("]
    for s in std:
        lines.append(f'\t"{s}"')
    if has_aep:
        if std:
            lines.append("")
        lines.append('\t"github.com/yueli-fx/aep-parser/internal/aep"')
    lines.append(")")
    return "\n".join(lines)


def main():
    text = SRC.read_text(encoding="utf-8")
    lines, funcs = parse_funcs(text)
    name_to_range = {n: (cs, ce) for n, cs, ce in funcs}

    drop_lines: set[int] = set()
    summary = []

    for fname, names in TARGETS.items():
        chunks = []
        for n in names:
            if n not in name_to_range:
                print(f"WARN: {n} not found in aep_test.go", file=sys.stderr)
                continue
            cs, ce = name_to_range[n]
            chunks.append("\n".join(lines[cs:ce]))
            drop_lines.update(range(cs, ce))
        body = "\n\n".join(chunks)
        imps = gather_imports(body)
        out = ["package aep_test", "", render_imports(imps), "", body, ""]
        path = DST_DIR / fname
        path.write_text("\n".join(out), encoding="utf-8")
        summary.append((fname, len(names), len(body.split("\n"))))

    # Rewrite aep_test.go without the moved blocks. Collapse consecutive
    # blank lines that result from removals.
    kept = [ln for i, ln in enumerate(lines) if i not in drop_lines]
    # collapse 3+ blank lines → 2
    out = []
    blank = 0
    for ln in kept:
        if ln == "":
            blank += 1
            if blank > 2:
                continue
        else:
            blank = 0
        out.append(ln)
    SRC.write_text("\n".join(out), encoding="utf-8")

    print(f"aep_test.go: {len(lines)} → {len(out)} lines (-{len(lines) - len(out)})")
    for fname, nt, nl in summary:
        print(f"  {fname}: {nt} tests/helpers, {nl} lines")


if __name__ == "__main__":
    main()
