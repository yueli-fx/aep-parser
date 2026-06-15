---
status: active
when_to_read: implementing Composition.DeleteLayer or any layer-removal mutation; debugging "AE drops a layer we wrote" or "AE complains after we delete"; reasoning about what chunks make up a layer's serialized unit; deciding orphan-reference cleanup strategy (ParentID / TrackMatteLayerID / matte type byte)
applies_to: [delete-layer, layer-removal, structural-mutation, itemList, Layr, Ewst, fvdv-followers, parent-orphan, trackmatte-orphan, ae2020, ae2025, deletelayer-re]
last_updated: 2026-05-28
---

# DeleteLayer RE — AE 行为契约

V3 Phase 2 DeleteLayer 实现 reference。**所有数字来自 AE 自己 save 的 4 个 fixture diff，不是猜的**。

Fixtures: `test_data/re_delete_layer_{baseline,middle,parent,matte}.aep`
- baseline / middle / parent → **AE 2020 17.7x45**
- matte → **AE 2025 25.1x68**（`TrackMatteLayerID` 字段是 AE 23+ 才存在，AE 2020 不写）

## Layout (constant across modes)

3 solid layers in single comp `DeleteTest`:

```
index 1 (top): L1_top  id=19   ← used as implicit matte source in matte mode
index 2 (mid): L2_mid  id=17   ← matted target / parent in matte/parent modes
index 3 (bot): L3_bot  id=15   ← child of L2 in parent mode
```

Variable names match indices because the JSX adds in REVERSE order (`addSolid` puts new layer at index 1; see `test_data/re_delete_layer.jsx`).

---

## Finding 1: itemList does SPLICE (Q1)

| Mode | Item LIST children | Δ |
|---|---|---|
| baseline | 237 | — |
| middle (L2 deleted) | 221 | **-16** |
| parent (L2 deleted) | 221 | **-16** |
| matte (L1 deleted) | 221 | **-16** |

AE **does not** leave a gap. Subsequent Layr/Ewst units shift up in the Item LIST. The baseline's Layr LISTs are at child indices [9, 25, 41]; after middle-delete they're at [9, 25].

**Impl rule**: `comp.back.itemList.Children` must `slice = append(slice[:i], slice[j:]...)` (where `j-i` = 16 — see Finding 4). Not zero-out or sentinel.

---

## Finding 2: ParentID resets to 0, NOT lifted to grandparent (Q2)

`parent` mode setup: `l3.parent = l2` (so L3's ParentID = L2.id = 17 in ldta @0x84).
Then `l2.remove()`.

AE 2020 JSX log post-delete:
```
[1] L1_top
[2] L3_bot          ← NO "parent=..." annotation
```

ExtendScript's `li.parent` returns `null` after delete → L3.ParentID has been reset to 0.

**AE does NOT lift to grandparent.** L1_top wasn't anyone's parent in baseline, so "lift to L1" would be the only sane "grandparent" option, but AE chose reset-to-0 instead.

**Impl rule**: For every layer `c.Layers[k]` (k ≠ deletedIdx), if `Layer.ParentID == deletedLayer.ID`, set `Layer.ParentID = 0` and write the cleared bytes back to ldta @0x84..0x87.

---

## Finding 3: TrackMatteLayerID resets to 0; trackMatteType byte STAYS SET (Q3)

`matte` mode setup: `l2.trackMatteType = TrackMatteType.ALPHA`. Implicit matte source = layer above = L1. AE 2025 auto-populates the explicit `TrackMatteLayerID` (ldta @0xA0, AE 23+) = L1.id.

JSX log pre-delete (AE 2025):
```
[2] L2_mid matteSrc=L1_top trackMatteType=5013
```

After `l1.remove()`:
```
[1] L2_mid trackMatteType=5013      ← matteSrc gone, trackMatteType still 5013
[2] L3_bot
```

**AE clears `TrackMatteLayerID` (ldta @0xA0) → 0, but leaves `trackMatteType` byte (ldta @0x6B) untouched (5013 = ALPHA persists).**

The matted layer now has a dangling matte-intent byte. Implicit lookup would resolve to "layer above" — but L2 is now at index 1, no layer above → no matte applied at render time. AE knowingly leaves the byte; it's the user's responsibility to clear via UI if they want.

**Impl rule**:
- For every layer `c.Layers[k]` (k ≠ deletedIdx), if `Layer.TrackMatteLayerID == deletedLayer.ID`, set `Layer.TrackMatteLayerID = 0` and write zero bytes back to ldta @0xA0..0xA3.
- **Do NOT touch ldta @0x6B (trackMatteType byte)**. Match AE's behavior.

---

## Finding 4: The delete unit is 16 Item-LIST-level chunks (Q4)

Each layer occupies a contiguous 16-chunk block at the Item LIST level:

```
[N]   LIST Layr  (the layer payload)
[N+1] LIST Ewst  (empty 0-child sibling)
[N+2..N+8]  fvdv fiop ftts foac fiac fipc fifl   (post-Layr metadata, group 1)
[N+9..N+15] fvdv fiop ftts foac fiac fipc fifl   (post-Layr metadata, group 2)
```

7 chunks × 2 groups = 14 leaf followers. Plus 1 Layr + 1 Ewst = **16 chunks per layer**.

`Δ = 237 - 221 = 16` across all delete modes confirms AE removes the entire 16-chunk block.

**Crucial**: our current parser (`parse_composition.go`) only walks Layr LISTs (`item.FindAllList(rifx.IDLayr)`); it does NOT track the fvdv/fiop/ftts/foac/fiac/fipc/fifl follower chunks per-layer. These are preserved by round-trip *only because* the rifx layer keeps the full Item LIST children slice byte-identical. If we remove a Layr without also removing its 14 follower chunks, the file will be corrupted (followers now point at the wrong layer / count mismatch).

**Impl rule**: `DeleteLayer` must locate the Layr's index `i` in `comp.back.itemList.Children`, then splice out `Children[i : i+16]` (verify `Children[i].FormType == IDLayr` and `Children[i+1].FormType == IDEwst` before splicing — bail with error if invariant broken).

**Implication for V3 Phase 1 backrefs**: the 14 follower chunks per layer are currently "opaque" — not in any `Layer.back` field. For Phase 2 we don't need to model them (just splice as a contiguous block); for Phase 3+ DuplicateLayer / InsertLayer we MAY need to track them per-layer to clone correctly.

---

## Finding 5: Pseudo-layer LISTs also follow the same 16-chunk pattern

baseline Item LIST contains, beyond the 3 user Layr LISTs:
- `DLay` (1×) — index [41]
- `SLay` (6×) — indices [57, 73, 89, 105, 121, 137]
- `CLay` (3×) — indices [153, 169, 185]
- `SecL` (1×) — index [201]

Each followed by Ewst + 14 leaf chunks (same pattern). AE-only pseudo-layers (Render Queue / Essential Graphics / Markers / etc) — DeleteLayer should NEVER touch these. Filter by FormType == IDLayr only.

**Impl rule**: when locating the layer to delete, validate `Children[i].FormType == rifx.IDLayr` AND `Children[i].ID == "LIST"`. Refuse to delete DLay/SLay/CLay/SecL even if caller passes a matching index (these don't exist in `c.Layers` anyway — parseComposition only reads Layr).

---

## Finding 6: Layer IDs are monotonic-down-the-list and NOT reused (Q5 partial)

baseline IDs: L1=19, L2=17, L3=15. (Counter incremented per add: L3 added first → 15, then L2 → 17, then L1 → 19. Gaps of 2 because each solid also creates a Footage item that bumps the counter.)

After delete:
- middle: L1=19, L3=15 (L2.id=17 gone, NOT reused)
- parent: L1=19, L3=15 (same)
- matte: L2=17, L3=15 (L1.id=19 gone, NOT reused)

Confirms CLAUDE.md Invariant #9 (monotonic, no reuse). **No need to down-adjust the head chunk's nextItemID counter** after delete.

**Impl rule**: leave `proj.back.head.nextItemID` alone after DeleteLayer.

---

## Open / Not RE'd

- **Q6 (expression / render queue string refs to deleted layer ID)**: not investigated. Plan accepts this as out-of-scope for Phase 2 — user responsibility. Add doc warning on `DeleteLayer`.
- **Multi-version cross-check for matte**: matte fixture is AE 2025 only. AE 2020 + `l2.trackMatteType = ALPHA` would NOT write `TrackMatteLayerID` (field is AE 23+), so the explicit-ID cleanup path is untested on AE 2020 saves. Our DeleteLayer needs to handle both: when opening an AE 2020 file, `TrackMatteLayerID` is 0 from the start (no orphan to clear); when opening AE 23+ file with explicit ID set, clear it per Finding 3.
- **Multi-pseudo-layer-type delete**: refuse, but not formally tested by trying to delete DLay/SLay etc via JSX. Refuse-by-FormType-check is the conservative path.

---

## Related

- Plan: [`../plans/2026-05-28-v3-phase2-deletelayer-plan.md`](../archive/plans/2026-05-28-v3-phase2-deletelayer-plan.md)
- Fixture JSX: `test_data/re_delete_layer.jsx`
- Dump tool: `tmp_debug/dump_layers/main.go`
- Ship-gate scar (when AE rejects what we wrote): [`ae25-acceptance-gate.md`](ae25-acceptance-gate.md)
