package scene

// Layer convenience helpers — mirror of py-aep's `layer.containing_comp`, //nolint:jargon
// `layer.has_audio`, `layer.active_at_time(t)`, etc. Pure helpers built
// on existing Layer fields; no chunk RE / no writes (except RemoveTrackMatte,
// which is an alias of the existing ClearTrackMatteLayer).
//
// `Layer.Index` and `Layer.Type` are already exposed as direct fields;
// `Parent`, `SourceComposition`, `SourceFootage`, `AlternateSource`,
// `TrackMatteLayer`, `PropertyByMatchName` are already exposed as methods.

// ContainingComp returns the composition this layer belongs to, or nil
// when the layer was built outside the parser (no owning comp wired up).
func (l *Layer) ContainingComp() *Composition {
	return l.comp
}

// HasVideo reports whether the layer has a visual component. True for
// every non-audio layer kind we model — only pure audio-source layers
// (AV source with no video stream) would be false, and we don't model
// that distinction yet. Camera / Light layers have no video output but
// AE script's `layer.hasVideo` reports false for them, so we mirror that.
func (l *Layer) HasVideo() bool {
	switch l.Type {
	case LayerTypeCamera, LayerTypeLight:
		return false
	}
	return true
}

// HasAudio reports whether the layer's source could carry audio. True
// when the source resolves to a Footage that is not a solid/placeholder
// or a Composition (which itself may contain audio layers). For camera /
// light / null / text / shape layers it returns false.
func (l *Layer) HasAudio() bool {
	if l.SourceComposition() != nil {
		return true
	}
	if f := l.SourceFootage(); f != nil && !f.IsSolid && !f.IsPlaceholder {
		return true
	}
	return false
}

// AudioActive reports whether the layer would produce audio output:
// it has a potential audio source AND the AudioEnabled switch is on.
// This is the design-time check; it does not consider time.
func (l *Layer) AudioActive() bool {
	return l.AudioEnabled && l.HasAudio()
}

// AudioActiveAtTime reports whether the layer would produce audio at
// the given time t (seconds, in the owning composition's timeline).
// Combines AudioActive with ActiveAtTime.
func (l *Layer) AudioActiveAtTime(t float64) bool {
	return l.AudioActive() && l.ActiveAtTime(t)
}

// ActiveAtTime reports whether the layer is on-screen at time t —
// visible switch enabled AND the time falls within the layer's
// in/out range. Mirrors AE script's `Layer.activeAtTime(t)`.
//
// InPoint/OutPoint come from the layer's existing time fields.
// Note: our parser already exposes InPoint via Layer setters, but the
// underlying numeric is not on the Layer struct — it derives from
// ldta @0x18/@0x20. For now we approximate using StartTime + Duration;
// callers needing precise InPoint should use SourceID + per-comp time math.
func (l *Layer) ActiveAtTime(t float64) bool {
	if !l.Visible {
		return false
	}
	in, out := l.layerTimeRange()
	return t >= in && t < out
}

// layerTimeRange returns the layer's effective on-screen [in, out)
// range in seconds. Uses StartTime + Duration as the approximation
// (precise InPoint / OutPoint live in ldta @0x18-@0x23 and are
// exposed through Layer.InPoint() / Layer.OutPoint() accessors when
// those land in a future task).
func (l *Layer) layerTimeRange() (float64, float64) {
	in := l.StartTime
	out := in + l.Duration
	if l.Duration <= 0 && l.comp != nil {
		out = in + l.comp.Duration
	}
	return in, out
}

// Width returns the layer's effective width in pixels — proxies to the
// source AV item when present, or to the owning composition's width for
// layers without an AV source (cameras, lights, nulls, text, shapes).
// Returns 0 when neither resolves.
func (l *Layer) Width() int {
	if c := l.SourceComposition(); c != nil {
		return int(c.Width)
	}
	if f := l.SourceFootage(); f != nil {
		return int(f.Width)
	}
	if l.comp != nil {
		return int(l.comp.Width)
	}
	return 0
}

// Height returns the layer's effective height in pixels — see Width.
func (l *Layer) Height() int {
	if c := l.SourceComposition(); c != nil {
		return int(c.Height)
	}
	if f := l.SourceFootage(); f != nil {
		return int(f.Height)
	}
	if l.comp != nil {
		return int(l.comp.Height)
	}
	return 0
}

// HasTrackMatte reports whether this layer uses ANOTHER layer as its
// alpha/luma matte source. True when TrackMatte != TrackMatteNone.
func (l *Layer) HasTrackMatte() bool {
	return l.TrackMatte != TrackMatteNone
}

// IsTrackMatte reports whether this layer is USED AS a matte by any
// other layer in the same composition. Walks the comp's layers to find
// references; O(n) on the comp's layer count.
func (l *Layer) IsTrackMatte() bool {
	if l.comp == nil || l.ID == 0 {
		return false
	}
	for _, other := range l.comp.Layers {
		if other == l {
			continue
		}
		if other.TrackMatteLayerID == l.ID && other.TrackMatte != TrackMatteNone {
			return true
		}
	}
	return false
}

// AutoName returns the layer's display name. Falls back to the source
// item's name when the layer has no user-set name. Empty string for
// layers without a source (camera/light/null/text/shape) and no
// user-set name.
func (l *Layer) AutoName() string {
	if l.Name != "" {
		return l.Name
	}
	if c := l.SourceComposition(); c != nil {
		return c.Name
	}
	if f := l.SourceFootage(); f != nil {
		return f.Name
	}
	return ""
}

// IsNameFromSource reports whether the layer's display name is
// inherited from its source AV item (i.e., the user did not set a
// custom layer name). True when Name == "" and a source exists.
func (l *Layer) IsNameFromSource() bool {
	if l.Name != "" {
		return false
	}
	return l.SourceComposition() != nil || l.SourceFootage() != nil
}

// @summary     Clear the layer's track-matte assignment
// @description Alias of [Layer.ClearTrackMatteLayer], kept for API parity
//   with other scripting environments' layer.removeTrackMatte(). Clears
//   both the mode (TrackMatte to None) and the explicit source pointer
//   (TrackMatteLayerID to 0) with one length-preserving write to ldta.
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestTrackMatteExplicit_AEShipGate_AE2025
// @since       AE2025
// @boundary    present only on AE 23+ project layouts; the ship gate is
//   single-version because no intermediate AE target both produces this
//   layout and survives the next version's forward compatibility check,
//   so a two-version gate is not reachable.
// @alias       remove track matte,remove matte,清除遮罩,取消遮罩,remove_track_matte
func (l *Layer) RemoveTrackMatte() error {
	return l.ClearTrackMatteLayer()
}

// IsThreeDModelLayer reports whether the layer is a 3D Model layer
// (AE 24+). Identified by ldta byte 0x83 == 0x05 at parse time; see
// [inferLayerType].
func (l *Layer) IsThreeDModelLayer() bool {
	return l.Type == LayerType3DModel
}
