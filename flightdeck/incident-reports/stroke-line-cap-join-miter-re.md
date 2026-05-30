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
