package scene

// CompItem filter views — convenience accessors that return layers of a
// given kind (text layers, shape layers, etc.). Each method walks Layers
// once and returns a filtered slice. Results are computed on every call; if
// you need to iterate the same view repeatedly, cache it locally.

type compositionSourceIndex struct {
	proj         *Project
	compositions map[uint32]*Composition
	footages     map[uint32]*Footage
}

func newCompositionSourceIndex(c *Composition) compositionSourceIndex {
	if c == nil || c.proj == nil {
		return compositionSourceIndex{}
	}
	index := compositionSourceIndex{proj: c.proj}
	if len(c.proj.Compositions) > 0 {
		index.compositions = make(map[uint32]*Composition, len(c.proj.Compositions))
		for _, item := range c.proj.Compositions {
			if item == nil || item.ID == 0 {
				continue
			}
			if _, exists := index.compositions[item.ID]; !exists {
				index.compositions[item.ID] = item
			}
		}
	}
	if len(c.proj.Footage) > 0 {
		index.footages = make(map[uint32]*Footage, len(c.proj.Footage))
		for _, item := range c.proj.Footage {
			if item == nil || item.ID == 0 {
				continue
			}
			if _, exists := index.footages[item.ID]; !exists {
				index.footages[item.ID] = item
			}
		}
	}
	return index
}

func (index compositionSourceIndex) composition(id uint32) *Composition {
	if id == 0 || index.compositions == nil {
		return nil
	}
	return index.compositions[id]
}

func (index compositionSourceIndex) footage(id uint32) *Footage {
	if id == 0 || index.footages == nil {
		return nil
	}
	return index.footages[id]
}

func (index compositionSourceIndex) layerComposition(l *Layer) *Composition {
	if l == nil || l.comp == nil || l.comp.proj == nil {
		return nil
	}
	if l.comp.proj != index.proj {
		return l.SourceComposition()
	}
	return index.composition(l.SourceID)
}

func (index compositionSourceIndex) layerFootage(l *Layer) *Footage {
	if l == nil || l.comp == nil || l.comp.proj == nil {
		return nil
	}
	if l.comp.proj != index.proj {
		return l.SourceFootage()
	}
	return index.footage(l.SourceID)
}

// TextLayers returns all text layers in the composition.
func (c *Composition) TextLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.Type == LayerTypeText {
			out = append(out, l)
		}
	}
	return out
}

// ShapeLayers returns all shape layers in the composition.
func (c *Composition) ShapeLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.Type == LayerTypeShape || l.IsShapeLayer {
			out = append(out, l)
		}
	}
	return out
}

// CameraLayers returns all camera layers in the composition.
func (c *Composition) CameraLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.Type == LayerTypeCamera {
			out = append(out, l)
		}
	}
	return out
}

// LightLayers returns all light layers in the composition.
func (c *Composition) LightLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.Type == LayerTypeLight {
			out = append(out, l)
		}
	}
	return out
}

// NullLayers returns all null-object layers (IsNull flag set).
func (c *Composition) NullLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.IsNull {
			out = append(out, l)
		}
	}
	return out
}

// AdjustmentLayers returns all adjustment layers (IsAdjust flag set
// or Type classified as adjustment).
func (c *Composition) AdjustmentLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.IsAdjust || l.Type == LayerTypeAdjust {
			out = append(out, l)
		}
	}
	return out
}

// ThreeDLayers returns all layers with the 3D-layer flag enabled
// (transform interpreted in 3D space).
func (c *Composition) ThreeDLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.Is3D {
			out = append(out, l)
		}
	}
	return out
}

// GuideLayers returns all layers marked as guide (IsGuide flag set
// — invisible at render but visible in the timeline).
func (c *Composition) GuideLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.IsGuide {
			out = append(out, l)
		}
	}
	return out
}

// SoloLayers returns all layers with the Solo flag set.
func (c *Composition) SoloLayers() []*Layer {
	out := make([]*Layer, 0)
	for _, l := range c.Layers {
		if l.Solo {
			out = append(out, l)
		}
	}
	return out
}

// AVLayers returns all layers backed by an audio/video source —
// pre-comps, footage files, solids, placeholders. Excludes camera /
// light / null / text / shape / adjustment-only layers (which carry no
// AVItem source).
func (c *Composition) AVLayers() []*Layer {
	out := make([]*Layer, 0)
	sources := newCompositionSourceIndex(c)
	for _, l := range c.Layers {
		if l.SourceID != 0 && (sources.layerComposition(l) != nil || sources.layerFootage(l) != nil) {
			out = append(out, l)
		}
	}
	return out
}

// CompositionLayers returns layers whose source is another composition
// (pre-comp layers).
func (c *Composition) CompositionLayers() []*Layer {
	out := make([]*Layer, 0)
	sources := newCompositionSourceIndex(c)
	for _, l := range c.Layers {
		if sources.layerComposition(l) != nil {
			out = append(out, l)
		}
	}
	return out
}

// FootageLayers returns layers whose source is a footage item
// (file / solid / placeholder).
func (c *Composition) FootageLayers() []*Layer {
	out := make([]*Layer, 0)
	sources := newCompositionSourceIndex(c)
	for _, l := range c.Layers {
		if sources.layerFootage(l) != nil {
			out = append(out, l)
		}
	}
	return out
}

// FileLayers returns layers whose source is a file-backed footage item
// (not a solid and not a placeholder).
func (c *Composition) FileLayers() []*Layer {
	out := make([]*Layer, 0)
	sources := newCompositionSourceIndex(c)
	for _, l := range c.Layers {
		f := sources.layerFootage(l)
		if f != nil && !f.IsSolid && !f.IsPlaceholder {
			out = append(out, l)
		}
	}
	return out
}

// SolidLayers returns layers whose source is a solid footage item.
func (c *Composition) SolidLayers() []*Layer {
	out := make([]*Layer, 0)
	sources := newCompositionSourceIndex(c)
	for _, l := range c.Layers {
		f := sources.layerFootage(l)
		if f != nil && f.IsSolid {
			out = append(out, l)
		}
	}
	return out
}

// PlaceholderLayers returns layers whose source is a placeholder footage
// item (AE's "Missing Footage" placeholder, opti tag = "Plac").
func (c *Composition) PlaceholderLayers() []*Layer {
	out := make([]*Layer, 0)
	sources := newCompositionSourceIndex(c)
	for _, l := range c.Layers {
		f := sources.layerFootage(l)
		if f != nil && f.IsPlaceholder {
			out = append(out, l)
		}
	}
	return out
}
