<!-- Hand-authored note. -->

## Building & editing shapes (alpha)

Constructing or restructuring shape layers from scratch goes through the
`ShapeLayer` wrapper (`WrapShapeLayer`) and the `VectorGroup` scene graph — an
alpha API tracked under the V3 scene-graph work, not covered on this page. The
`ShapePath` / `ShapePrimitive` types above are the **read model** produced by
parsing an existing shape layer.

### Current limitations

- Editing path vertices / geometry of a parsed `ShapePath` is not implemented
  (it would require rewriting the variable-length `shap` keyframe stream).
- A `ShapePrimitive`'s parametric sub-properties (`Size`, `Position`,
  `Roundness`, the star fields) are nil when AE elided their `cdat` (defaulted
  values); when present they read/write through the usual `Property` API.
