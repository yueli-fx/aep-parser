<!-- Hand-authored notes. -->

## Current limitations

- Editing vertices / path geometry is not implemented (would require rewriting
  the variable-length `shap` kfl stream).
- Adding / removing masks is unsupported (structural change).
- The remaining unparsed `mkif` bytes (`0x10` / `0x18` / `0x20`–`0x27`) don't
  appear to be user-configurable constant fields, but haven't been fully
  reverse-engineered.
- Per-keyframe `Closed` flags on animated masks can't be set frame-by-frame
  (only the first snapshot).
