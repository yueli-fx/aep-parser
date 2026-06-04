<!-- Hand-authored notes. go/doc comments can't carry these tables / caveats. -->

## Label colors

`Marker.Label` / `SetLabel` is the timeline label-color index, `0..16`. `0` is
AE's default color; the remaining indices map to AE's label palette (the same
swatches used for layer labels). Values are written verbatim, so an
out-of-range byte round-trips even though AE displays it as index 0.

## Current limitations

- Adding / removing markers within an existing set is not supported through the
  `Marker` API — it requires re-laying-out the `ldat` blocks and `Nmrd` list in
  lock-step, which breaks the length-preserving write contract. Use the
  composition / layer structural marker ops where available.
