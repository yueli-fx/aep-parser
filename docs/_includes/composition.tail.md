<!-- Hand-authored reference table. -->

## Renderer engines

`Composition.Renderer` / `SetRenderer` use the binary match-name stored in the
`prin` chunk. AE Scripting's `CompItem.renderer` exposes a different module name
in one case (Advanced 3D); the others match.

| binary match_name | ExtendScript | UI |
|---|---|---|
| `ADBE Escher` | `ADBE Advanced 3d` | Advanced 3D (Classic 3D on older AE) |
| `ADBE Calder` | `ADBE Calder` | Advanced 3D (AE 2025) |
| `ADBE Ernst` | `ADBE Ernst` | Cinema 4D |
| `ADBE Picasso` | `ADBE Picasso` | Ray-traced 3D |

Available engines differ per AE version (AE 2020 = Escher / Standard / Ernst;
AE 2025 = Calder / Ernst / Picasso, and on load it auto-promotes the deprecated
Escher / Picasso to Advanced 3D). An empty string means the comp has no `PRin`
LIST (rare; only programmatically built comps that skipped AE serialization).
