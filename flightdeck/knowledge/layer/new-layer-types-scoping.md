# ⚠ New layer types — scoping → ALL SHIPPED (2026-06-10)

SUMMARY: New layer types — scoping → ALL SHIPPED (2026-06-10)
READ WHEN: implementing NewSolidLayer / NewNullLayer / NewAdjustmentLayer / NewCameraLayer / NewLightLayer; scoping a new from-scratch layer-creation path; deciding embed-template vs from-scratch for a footage-backed layer; RE'ing the solid footage Item (opti "Soli" / Pin / sspc); patching solid color / dimensions / footage name

---

Camera/Light shipped first ([[camera-light-layer-create-re]]); Solid/Null/
Adjustment shipped same day — AE 2020 + 2025 ship-gate PASS
(`TestNewSolidNull_AEShipGate_AE20{20,25}`). § "As shipped" below records the
actual mechanism (it differs from the original recommendation in one key way:
whole-template-project + InsertLayer reuse, NOT per-chunk .bin extraction).
The whole New\* family (Camera/Light/Solid/Null/Adjustment/Text) was promoted
Alpha → Stable (structural op, CLAUDE.md #2) in the 2026-06-11 audit on the
dual-version gate evidence.

## As shipped — solid family (mutate_layer_solid.go)

- **Embed a whole AE-2020 template PROJECT** (`templates/solidnull_2020.aep`,
  one comp "TplComp" with layers TplSolid/TplNull/TplAdj) and **reuse
  `insertLayerCrossProject`** — the existing gated machinery already does
  footage-closure import with fresh item IDs + SourceID remap, which is exactly
  solid creation. No per-chunk extraction, no new splice code. (Called directly,
  not via the public `InsertLayer` wrapper — its refuse-matrix only admits AV
  src, and a template null parses as `LayerTypeNull`.)
- **Per-instance patches, all length-preserving**: layer Utf8 name (managed
  splice) · time-span re-home 0→comp duration (`Set{StartTime,InPoint,OutPoint}`)
  · opti name @0x1A · opti color · sspc W/H · ldta pad/trim to capability
  LdtaSize · cdta @0x18 "user content" bump.
- **opti "Soli" layout (282B, RE: re_solidnull.aep)**: tag @0x00, **ARGB
  4×float32 BE @0x0A** (A=1.0 always), **name @0x1A** NUL-terminated in a fixed
  256B tail → rename is length-preserving. Solid footage's Item-level Utf8 is
  EMPTY — the display name lives only in opti.
- **JSX gotcha**: a layer never explicitly renamed stores an empty Utf8 (AE
  shows the source name) — fixture JSX must set `layer.name = ...` or the Go
  side can't find template layers by name.
- The returned layer is already **parsed** (InsertLayer re-parses) — setters
  work immediately, no [[add-effect-splice-re]] Reopen needed.
- **ID-namespace trap (the one real reject)**: Go-built dest comps carry dummy
  service layers DLay/SLay/CLay/SecL with IDs 2..12 invisible to `c.Layers`;
  `allocItemID` handed out 2/3 → **AE 2025 rejects** with `unexpected match
  name searched for in group` (AE 2020 tolerates). Fixed in initDerived +
  NewComposition — details in [[nextitemid-must-include-layer-ids]].

## Original scoping (kept for the record)

## Two families

1. **Footage-backed: Solid / Null / Adjustment.** Each is an AV layer whose
   `ldta @0x28` SourceID points at a **solid footage Item** living in an
   auto-created "Solids" **folder** Item. Null = a 100×100 solid + `isNull`
   ldta flag; Adjustment = a white solid + `isAdjust` flag. So solving **Solid**
   solves all three (Null/Adjustment are flag + dimension/color variants).
2. **Source-less: Camera / Light.** No footage Item — defined entirely by ldta
   (layer subtype byte @0x83) + a layer-specific property group (Camera Options
   / Light Options). Closer to the ShapeLayer model (embed a property-tree
   template, no project-level item to create). Likely the *easier* family.

## Solid footage Item structure (RE'd from re_effect_library.aep)

A solid footage Item is `LIST(Item)` with children:

```
iide(4) idpc(8) idta(84, type byte @1 == 7 footage) Utf8(name)
LIST(Pin ):
  sspc(222)            ← footage spatial: width/height/pixel-aspect (real-AE 222B layout, W/H @0x20/0x24 — see Footage.sspcChunk)
  Utf8(0)
  opti(282, tag "Soli")  ← SOLID DESCRIPTOR: "Soli" + color + name (the bytes to RE for SetColor/dims)
  pgui(16)
  LIST(CLRS): epid apid linl embp ipws dcui prgb   ← color-management metadata
  LIST(mnfo): strt drop                            ← misc footage info
  Utf8(0)
ftgi(16)             ← footage general info
Utf8(0)
LIST(Gide): gdta(8) + LIST(list){lhd3(52)}        ← same Gide boilerplate as every Layr
```

The "Solids" parent is a **folder Item** (idta type byte == 1), name "Solids" —
AE auto-creates it; new solids nest under it.

## Recommended approach (when built)

**Embed-template + ID-remap, NOT from-scratch byte synthesis.** Same lesson as
shape layers (from-scratch bodies trigger AE silent-drop; embedding AE-native
bytes bypasses byte-level RE — see [[v2-2-aelayer-structure]]) and the same
machinery as cross-Project `InsertLayer` (import an item closure with fresh IDs,
remap SourceID). Steps:

1. Extract a full solid footage Item + its "Solids" folder + the AV Layr from an
   AE fixture; embed both.
2. On `NewSolidLayer`: ensure/create the "Solids" folder, splice a cloned
   footage Item with a fresh project item ID, splice the Layr into the comp with
   a fresh layer ID + SourceID @0x28 remapped to the new footage ID. Reuse
   `allocItemID` + the InsertLayer closure-import infra.
3. **AE-acceptance ship-gate is mandatory + high-risk** — footage-Item creation
   has the same silent-drop / "项目文件似乎已损坏" surface that took
   NewComposition 5 RE phases ([[ae25-acceptance-gate]]). Budget for it.

## Color / dimensions

`SetColor` + dimensions need RE of `opti` "Soli" (282B) + `sspc` (222B; W/H
offsets already known @0x20/0x24 from Footage convenience work). A fixed-color
fixed-size solid (template defaults) is a low-value MVP — `SetColor`/`SetSize`
are what make it useful, so RE those offsets in the same pass.

## Why deferred (2026-06-10)

Substantial multi-chunk feature with a high-risk AE-acceptance gate; not cleanly
completable in a single unattended pass without risking rejected output. Camera/
Light (source-less) is the lower-risk entry point if picking one — it follows the
ShapeLayer embed-template model without project-level item creation.
