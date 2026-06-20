---
status: active
when_to_read: a from-scratch BuildPseudoEffect control does not RENDER in AE's Effect Controls panel (the value round-trips / DOM reads fine but the row is missing); implementing/RE'ing any pseudo control's pard or value-entry bytes; deciding which controls need a value entry vs build from the pard; debugging slider value shown ×255 or with a % sign; a pseudo gate that passes on DOM/round-trip but the panel is wrong (false-green); needing the decoded pard/tdb4 field map
applies_to: [pseudo-effect, build-pseudo-effect, pard, tdb4, value-entry, tdsb, tdpi, matchname, render-not-dom, false-green, layer-picker, point, point3d, slider, label, group, keyframe, hold-keyframe, control-type, field-map, red-line-4, ae2020, ae2025]
last_updated: 2026-06-20
recurrences: 1
---

# Pseudo controls: AE renders the PANEL from the pard — the decoded field map + the render rules

## Signature
- symptom: a from-scratch `BuildPseudoEffect` control (point / 3D point / slider / layer-picker) is **missing from AE's Effect Controls panel** even though `WriteAEP`→reparse round-trips clean and ExtendScript DOM reads the value.
- where: `internal/serializer/mutate_pseudo_effect_build.go` (pard + value-entry synthesis).
- trigger: building a pseudo effect from scratch and opening it in AE.

## Why every gate was false-green (the core lesson)

The pseudo gates checked **AE-accept + DOM readback + resave**, never **panel render**. A control can exist in the property tree (DOM-readable, byte-round-trips) yet **not render** in the panel. So "Go round-trip 0 警告" and "DOM reads value=X" both passed while the user saw a blank/partial panel. Red-line-4 in its purest form: **value-correct ≠ rendered**. A pseudo gate MUST screenshot the panel (or count visible rows), not just read DOM.

## Decoded field map (multi-instance diff of rich_demo + pseudo2 + pseudo.aep)

**pard (148 B)** — only ~6 meaningful fields; everything else is padding/garbage:

| offset | field | meaning |
|---|---|---|
| `@0x04` u32 | flags | `0x20`=**dim/gray** (an OPTION on a label; default 0 = non-gray — a label is a self-closing 0x0d group, the 0x20 bit only dims it) · `0x08`=group-end · `0x200`=an *optional* slider bit (varies; removing it is safe) |
| `@0x0C` u32 | **control_type** | 03 angle·04 checkbox·05 color·06 point·07 dropdown·0a slider·0d group/label·0e group-end·12 3d·00 header/layer-picker |
| `@0x10–0x2F` | name | **system ANSI codepage** (NOT UTF-8) — see [[pseudo-control-label-ansi-codepage]] |
| `@0x30` u32 | kind | **2 = structural/reference** (header / layer-picker / group / label) · **0 = value-leaf** |
| `@0x38+` | value | angle deg(16.16) · checkbox @38 state/@3C default · color @38/@3C ARGB · **point @38 x / @3C y (16.16 fractions)** · **3d @38/@40/@48 x/y/z (f64)** · dropdown @38 sel /@3C (count<<16\|sel) · slider @38(f64) default /@68 @6C valid range /@70 @74 visible range /@78(f32) default /`@7C` display flags |
| `@0x50+` | — | **uninitialized garbage** (differs per-instance in AE's own files); we write zeros, AE accepts |

- **slider `@0x7C`**: `0x00020000` = plain numeric · `0x00050003` = percent display (value shown ×255). The "19125 instead of 75" bug was this byte.

**value-entry `tdbs`** = `[tdsb, tdsn, tdb4, cdat, (tdpi, tdps)]`:
- **`tdsb` flag: `1` = plain property · `3` = effect-header anchor.** A layer-picker flagged 3 is **hidden** — it must be 1.
- `tdsn` = label (UTF-8, unlike the pard name). `cdat` = value bytes. `tdpi` = bound layer id. `tdps` = 0.
- **`tdb4` (124 B)**: `@0x02` dimension (1 scalar/ref · 2 point · 3 3d · 4 color) · `@0x0C` = constant `0x5da8` (universal, copy it) · `@0x10..` per-dimension matrix (comp-aspect-dependent for spatial). Only the header + layer-picker carry one now, both dim-1 (comp-independent) → safe to copy verbatim.

## The render rules (what AE actually requires)

1. **AE builds the panel from the pard (`parT`), NOT the value group.** Proof: AE's own all-defaults effect has point/3d/slider in `parT` with **no** value-group entry, and they render. So the value group only carries *non-default values + the bindings*.
2. **A control's default must live IN the pard.** §9's earlier RE was wrong ("coords live only in the value entry"). Point/3d are **hidden** if their pard carries no value:
   - point → `@0x38` x, `@0x3C` y as **16.16 fixed** fractions of the layer/comp space.
   - 3d → `@0x38`/`@0x40`/`@0x48` x/y/z as **f64** fractions.
3. **Emit NO value entry for point/3d/labels/groups** — they build from the pard. A synthesized value entry for them *hides* the control. **Only the effect header and layer-pickers carry a value entry** (they hold a tdpi binding the pard can't).
4. **Layer-picker value entry: `tdsb=1` (not 3) AND `tdpi` = a resolvable layer.** A picker with `tdpi=0` (None) or pointing at a non-host layer it can't anchor is hidden; a fresh PEM picker defaults to its **host** layer. So an unset `LayerID` binds the host.
5. **matchName must be PEM-canonical `Pseudo/<uid>`** — NO `/<name>` segment. Value params tolerate an extra segment; the layer-picker does not.
6. **slider `@0x04` carries no required flag** — the old RE wrote `0x200`, which (with the elided value entry) hid the slider. Leave it 0; set `@0x7C=0x00020000` for a plain numeric display.

## Keyframe / hold-interpolation = inherent to control_type

There is **no per-param keyframe/hold flag** in the pard. All value types share `@0x04`/`@0x30` and `@0x50+` is garbage. AE derives the behavior from `@0x0C` (slider/angle/point/color = smooth keyframe; checkbox/dropdown = hold/stepped; label = static). Confirmed by elimination, not yet by a controlled toggle sample.

## Residual uncertainties (honestly not fully decoded)

- point `@0x44/@0x48` (percent-display reference fields): our `×100` renders correctly but AE's `@0x48` looks like a constant `100` regardless of y — exact semantics unresolved (rich_demo's y=5.0 instance breaks the `y×100` reading). Harmless: AE reads the value from `@0x38/@0x3C`.
- slider `@0x7C` full bit semantics (only the plain-vs-percent split is known).
- tdb4 `@0x04–0x0B` type-descriptor bits (only dimension + the 0x5da8 constant decoded).

## Fix (committed)

`mutate_pseudo_effect_build.go`: point/3d write coords into the pard; point/3d/labels/groups elide their value entry; layer-picker value entry `tdsb=1` + host-default `tdpi`; matchName = `Pseudo/<uid>`; slider drops `@0x04`, sets `@0x7C=0x00020000`. Dead synthesis paths removed. Test `TestPseudoControlEntries_ValueRules`. **User-verified in AE**: all of slider / point / 3d / layer-picker now render correctly (Strength 75, Center 200,200, Position 3D 100,300,0, Source Layer = host). See [[pseudo-layer-picker-tdpi-retarget-clobber]] for the binding-target (vs render) side of the picker.

## Cases
- 2026-06-20 — from-scratch showcase (`flightdeck/showcase/pseudo-effect`) panel render; user supplied minimal AE-authored oracles (`tmp/pseudo2.aep` layer-picker-only `Pseudo/148432`) that enabled the byte-for-byte diff isolating each cause.
