# ⚠ SetLayerTransform encodes an AV/precomp layer's Anchor Point as fraction-of-source, not pixels

SetLayerTransform encodes an AV/precomp layer's Anchor Point as fraction-of-source, not pixels

## Signature
- symptom: `precomp/footage layer renders blank/off-screen after SetLayerTransform sets a non-zero anchor; AE DOM reads anchor ≈ value × source-dimension (960 → 1843200, 540 → 583200); parser round-trip reads the anchor back correctly`
- error_type: —
- where: SetLayerTransform anchor encoding (internal/serializer) for AV-typed layers · surfaced in flightdeck/showcase/booyah-clone/gen_precomp1.go
- trigger: set Anchor Point on a precomp/footage/AV layer via SetLayerTransform with a PIXEL value (e.g. comp-centre 960,540)

## Symptom / repro
Booyah 复刻 comp ⑤ プリコンポジション 1 (3 precomp layers → comp ①). Set each layer's
anchor to the source centre (960,540) via `SetLayerTransform.AnchorPoint().SetStaticValue`,
position to comp-centre (960,540). **Go round-trip reads anchor back as [960,540,0]** — looks
correct. But AE renders the comp **blank**; `verify`/DOM dump shows
`anchor.valueAtTime = 1843200, 583200` = `960×1920, 540×1080`. The pivot lands ~1.8M px away,
so the nested comp ① renders far off-screen → blank frame.

The bug is invisible at anchor (0,0) (0×N = 0): an earlier pass left anchor at
NewLayerTransform's (0,0) default and the layer rendered — just mis-placed to the
bottom-right quadrant (source top-left pivoting at the comp-centre position), which the
user caught visually ("中点在右下顶角"). Centering to (960,540) is what exposed the ×dim blow-up.

Classic red-line-1 false-green: parser round-trip is clean, only AE's DOM/ render reveals it.

## Root cause (observed, not yet fixed in lib)
`SetLayerTransform`'s Anchor Point encoding for an **AV / precomp / footage** layer treats the
input as a **fraction of the source dimensions** (0.5 = centre), whereas for **shape / text**
layers it is in **pixels** (comp ② set a pixel box-centre anchor, comp ④ used pixel (0,0) — both
correct). Writing pixels (960,540) on an AV layer makes AE read `960×W, 540×H`.

Empirically confirmed via AE DOM: writing anchor (0.5, 0.5) → AE reads (960, 540); writing
(960, 540) → AE reads (1843200, 583200).

## Workaround (in gen, until the lib is fixed)
Write the **source-centre as a FRACTION** for AV/precomp layers:
`tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5})` → AE reads source-centre. comp ⑤ then
renders the staggered comp ① copies centred (eyeballed at t=0.6, matches the original layout).

## Library follow-up (not done)
SetLayerTransform should normalize anchor units across layer kinds (AV vs shape/text) so a
caller can pass pixels uniformly — needs the encoding fix + a double-version render-pixel gate
on a precomp layer with a non-zero, non-centre anchor (the existing precomp gate
[[precomp-layer-source-id-re]] uses a bare precomp = template anchor, so it never exercised a
SetLayerTransform anchor on an AV layer). Until then, gens write the fraction and link here.

## Cases
- 2026-06-22 first — Booyah comp ⑤; bisected (bare precomp renders ✓ → +SetLayerTransform blank, with/without SetStartTime) then AE-DOM dump pinned anchor=1843200,583200. Workaround: anchor fraction (0.5,0.5). Render-confirmed centred.
