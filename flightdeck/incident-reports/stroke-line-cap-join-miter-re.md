---
when_to_read: implementing Stroke Line Cap / Line Join / Miter Limit setters; enriching the embedded stroke body template; debugging why a stroke enum slot is missing from a saved .aep; setting Miter Limit via JSX and hitting a "hidden property" throw
applies_to: [stroke, line-cap, line-join, miter-limit, shape-layer, enum, oned, elision, scripting-api-quirk, re-finding]
last_updated: 2026-05-31
---

# Stroke Line Cap / Line Join / Miter Limit — RE findings

RE'd via `test_data/re_stroke_linecap.jsx` (AE 2020) → `re_stroke_linecap.aep`.
Dump tool: `go run ./tmp_debug/dump_named_cdat <aep> "Stroke Line Cap" ...`.

## Corrects an earlier wrong groundwork note

The old cockpit backlog note claimed the matchName "is NOT `ADBE Vector Stroke
Line Cap` (JSX property-not-found)". **That is wrong.** `stroke.property("ADBE
Vector Stroke Line Cap")` (and `Line Join` / `Miter Limit`) resolve fine. A live
enumeration of the stroke group's 11 children confirms the documented matchNames:

```
ADBE Vector Blend Mode, ADBE Vector Composite Order,
ADBE Vector Stroke Color, ADBE Vector Stroke Opacity, ADBE Vector Stroke Width,
ADBE Vector Stroke Line Cap, ADBE Vector Stroke Line Join,
ADBE Vector Stroke Miter Limit,
ADBE Vector Stroke Dashes, ADBE Vector Stroke Taper, ADBE Vector Stroke Wave
```

The reason the slots were "missing" before: **AE elides any property at its
default value on save.** The old tolerance fixture only set Color/Width/Opacity,
so Cap/Join/Miter (all default) were elided → never appeared in the body.

## Encoding (all three)

OneD (`propertyValueType` 6417). Stored exactly like Opacity/Width: a `tdbs`
LIST per property with `tdb4`(124B) + `cdat`(40B), the value at `cdat[0:8]` as
**float64 BE**. Enums store their 1-based index as a float:

| Property | value→cdat[0:8] | default | enum |
|---|---|---|---|
| Line Cap | 2 → `4000000000000000` | 1 | 1=Butt 2=Round 3=Projecting |
| Line Join | 3 → `4008000000000000` | 1 | 1=Miter 2=Round 3=Bevel |
| Miter Limit | 4 → `4010000000000000` | 4 | (scalar, has `tdum`/`tduM` UI-range tail) |

Miter Limit's `tdbs` carries trailing `tdum`(8B)+`tduM`(8B) (UI min/max range);
the Cap/Join enums do not.

## Two behavioral gotchas

1. **AE writes Cap + Join + Miter as a unit.** Setting just Cap+Join non-default
   caused Miter Limit to persist too (at default 4). So once any of the three is
   touched, expect all three slots in the saved body.
2. **Miter Limit is hidden unless Line Join = Miter (1).** `setValue` on Miter
   throws `After Effects错误: 无法...set value...因为属性或父级属性被隐藏` if Line
   Join was already set to Bevel/Round. A JSX RE fixture must set Miter *before*
   changing Join; a runtime setter should document that Miter only applies when
   Join=Miter.

## Implication for the builder

The embedded `templates/v2_2_shape_stroke_body.bin` was extracted from a stroke
WITHOUT these slots. To emit Cap/Join/Miter, either re-extract a richer template
(stroke with all six scalar props non-default) and overwrite each `cdat[0:8]`, or
inject the three `tdbs` sub-trees on demand (mirrors AE elision). Whichever path,
it is a new structural write → AE 2020 + 2025 ship-gate required before ship.

## Addendum (2026-05-31): the rest of the shape enums

Same RE run family (`re_shape_enums.jsx` + combined template `gen_shape_all_full.jsx`)
covered the remaining OneD shape enums — all identical encoding (float64-BE @
cdat[0:8], 1-based index, default 1, AE elides default):

| Property | matchName | node(s) | default | values |
|---|---|---|---|---|
| Direction | `ADBE Vector Shape Direction` | Rect, Ellipse | 1 | 1=Normal, 3=Reversed (no 2) |
| Blend Mode | `ADBE Vector Blend Mode` | Fill, Stroke | 1 | AE 1-based index (Normal=1), large enum |
| Composite Order | `ADBE Vector Composite Order` | Fill, Stroke | 1 | 1=Above Previous, 2=Below Previous |
| Fill Rule | `ADBE Vector Fill Rule` | Fill | 1 | 1=Nonzero Winding, 2=Even-Odd |

Blend Mode / Composite Order are the first two children of every Fill/Stroke
group; Direction is the first child of every parametric shape (Rect/Ellipse).

**ExtendScript RE gotcha**: `group.addProperty(...)` reindexes the collection and
**invalidates handles obtained before later adds** — a handle to `rect` taken
before adding `stroke` throws `ReferenceError: 引用无效` on use. Re-fetch each
child by matchName after all adds (`re_shape_enums.jsx` `refetch()`).

Shipped as coverage 子项⑪ (template re-extracted for rect/ellipse/fill/stroke from
one combined `v2_2_shape_all_full.aep`; all shape ship-gates re-run dual-version).

## Addendum (2026-05-31): Stroke Taper + Wave nested groups

RE'd via `test_data/re_stroke_dtw.jsx` (AE 2020) — a probe+set fixture that
enumerates each group's children and sets them non-default. Three nested groups
hang off the stroke (`ADBE Vector Stroke Dashes / Taper / Wave`); the old stroke
template carried them as **empty placeholders** (`tdmn` immediately followed by
`ADBE Group End`) because `gen_shape_stroke_full.jsx` never set them and AE elides
default groups.

### Structure — all three are fixed-slot scalar groups (NOT free-form)

The cockpit's "variable nested group" worry was wrong. Each group's sub-children
pre-exist; you don't build them, you reveal/set them. Every sub-stream is a OneD
`float64` BE at `cdat[0:8]` — identical encoding to Cap/Join/Miter, one nesting
level deeper (inside the group's `LIST(tdgp)`, not the stroke body directly).

| Group | matchName | on-disk children | notes |
|---|---|---|---|
| Taper | `ADBE Vector Stroke Taper` | Length Units (enum), Start/End Length, StartWidthPx/EndWidthPx, Start/End Width, Start/End Ease (9 when Units≠default) | emits as a unit |
| Wave | `ADBE Vector Stroke Wave` | Amount, Units (enum), Wavelength, Cycles, Phase (angle) | emits as a unit |
| Dashes | `ADBE Vector Stroke Dashes` | Dash 1/2/3, Gap 1/2/3, Offset | **variable** — AE emits only the enabled Dash/Gap pairs |

Sub-stream tdbs shapes (same as the existing stroke scalars): scalar-with-range
= 6-child (`tdsb/tdsn/tdb4/cdat/tdum/tduM`); enum/angle = 4-child (no tdum/tduM,
tdb4 head carries `0002` / `0002ffff`).

### Elision + UI-coupling quirks (the load-bearing findings)

1. **Length Units = % (1, default) → Length Units + StartWidthPx/EndWidthPx are
   elided**, leaving Taper with 6 always-active slots (Start/End Length, Start/End
   Width, Start/End Ease, all `tdsb=00000001`). Setting Units = px (2) un-elides
   those three BUT flips Start/End Length into an inert `tdsb=00000003` state.
2. **Wave Units = Wavelength (1, default) → Units + Cycles elided**, leaving 3
   active slots (Amount, Wavelength, Phase). Setting `Wave Cycles` while
   Units=Wavelength throws *"属性或父级属性被隐藏"* (hidden-property) — same class
   as Miter-hidden-unless-Join=Miter.
3. **Dashes Offset is hidden until a dash is enabled**; `Offset.setValue` throws
   the same hidden-property error.

### V2.2 scope decision (子项⑫)

Modeled the 9 always-active `%`/Wavelength-mode scalars only — Taper Start/End
Length, Start/End Width, Start/End Ease; Wave Amount, Wavelength, Phase. Length
Units / Px mirrors / Wave Units / Cycles are deferred (their slots vanish or go
inert at the % default — same elision trap that hid Cap/Join/Miter). **Dashes is
deferred** to a follow-on: it is genuinely variable-cardinality on disk
(N enabled Dash/Gap pairs) and needs an enable/reveal runtime model.

Template re-extracted from `gen_shape_all_full.jsx` (now sets Taper+Wave with
Units left at %, so the 9 active slots emit). Serializer descends into the
Taper/Wave `LIST(tdgp)` via `findGroupBody` then reuses `overwriteShapeStreamCdat`.
Dual-version ship-gate (AE 2020 + 2025) PASS. RE fixture: `re_stroke_dtw.jsx`.

## Addendum (2026-05-31): Stroke Dashes shipped — single Dash+Gap pair (子项⑬)

Dashes turned out to be tractable for the common case despite the
variable-cardinality worry. Modeled **one Dash + Gap pair** — the dashed/dotted
line that covers most real use.

### Enable = template swap (not in-place reveal)

A solid stroke serializes the Dashes group as an **empty placeholder** (the
`tdmn` is immediately followed by `ADBE Group End`, no Dash/Gap leaves), so you
can't overwrite slots that aren't there. Rather than splice the group open at
runtime, the serializer keeps **two** embedded stroke bodies:

- `v2_2_shape_stroke_body.bin` — solid (Dashes empty placeholder), **unchanged**
- `v2_2_shape_stroke_dashed_body.bin` — superset carrying `ADBE Vector Stroke
  Dash 1` + `Gap 1` slots, extracted from an AE-saved dashed fixture
  (`v2_2_stroke_dashed.aep`)

`lowerStrokeNode` clones the dashed body iff `dashes.enabled`, then overwrites
Dash 1 / Gap 1 cdats (OneD f64-BE @ `[0:8]`, nested one level in the group's
`LIST(tdgp)` — same depth as Taper/Wave). Because the solid body is byte-identical
to before, every existing stroke-touching gate is regression-free **without
re-running them** (the `DisabledStaysSolid` round-trip proves the solid path is
unchanged). Hydrate flags `enabled` by the presence of a Dash 1 / Gap 1 leaf.

### Offset stays deferred (hard ScriptingAPI block, not a scope choice)

`v2_2_stroke_dashed.done` from the dashed-fixture RE recorded:
`offset SET failed (hidden): ... 无法对set value...因为属性或父级属性被隐藏`.
AE keeps Dashes Offset **hidden until a dash is enabled** AND refuses `setValue`
on it even then (same hidden-property class as Wave Cycles / Miter-unless-Join).
No JSX path can produce a fixture with a non-default Offset slot, so no template
can carry it — Offset is unreachable by the embed strategy, not merely descoped.
Dash 2/3 + Gap 2/3 are deferred too (each enabled pair is a distinct on-disk
cardinality → a separate template variant).

Dual-version ship-gate (AE 2020 + 2025) PASS (`TestV2_2_StrokeDashes_AEShipGate_*`,
Dash=18 / Gap=7 decoded from AE's resave). RE fixtures: `re_stroke_dtw.jsx` +
`v2_2_stroke_dashed.aep`.
