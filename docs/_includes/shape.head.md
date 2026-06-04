<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

Shape layers hold parametric primitives (rectangles, ellipses, star polygons)
and/or freeform Bezier paths. After parsing, a shape layer exposes two read
models, both reachable from the layer:

- `ShapePath` (`Layer.ShapePaths`) — a freeform Bezier path; vertices use the
  same model as mask vertices (see [mask.md](mask.md)).
- `ShapePrimitive` (`Layer.ShapePrimitives`) — a parametric Rect / Ellipse /
  Star, each sub-property a regular [`Property`](property.md).
