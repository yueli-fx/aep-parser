package scene

import "fmt"

// @summary     Set another layer as this layer's explicit track-matte source
// @param       src   the layer to use as the matte source
// @param       mode  the matte mode to apply
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestTrackMatteExplicit_AEShipGate_AE2025
// @since       AE2025
// @boundary    mirrors AE ScriptingAPI 23+'s layer.setTrackMatte(srcLayer,
//   type) and requires the AE 23+ ldta layout (the @0xA0 slot) — files saved
//   by AE 22 / 2020 need a resave through AE 23+ first. length-preserving:
//   writes ldta @0xA0..0xA3 = src.ID (4 bytes BE) and @0x6B = mode (1 byte),
//   delegating to SetTrackMatteLayer. Refuses a nil src, either layer
//   lacking a comp back-pointer, a cross-comp matte (AE 23+ requires the
//   same comp), and a self-matte (src.ID == the layer's own ID). Passing
//   mode=TrackMatteNone with a non-nil src preserves the source pointer
//   while disabling the matte channel; use ClearTrackMatteLayer to fully
//   clear it. The ship gate is single-version because no intermediate AE
//   target both produces this layout and survives the next version's
//   forward compatibility check, so a two-version gate is not reachable.
// @alias       set track matte source,track matte,setTrackMatte,设置遮罩来源,遮罩图层
func (l *Layer) SetTrackMatteSource(src *Layer, mode TrackMatteType) error {
	if src == nil {
		return fmt.Errorf("SetTrackMatteSource: matte source is nil")
	}
	if l.runtime.comp == nil {
		return fmt.Errorf("SetTrackMatteSource: layer %q has no comp back-ref (built outside parser?)", l.Name)
	}
	if src.runtime.comp == nil {
		return fmt.Errorf("SetTrackMatteSource: matte source %q has no comp back-ref (built outside parser?)", src.Name)
	}
	if l.runtime.comp != src.runtime.comp {
		return fmt.Errorf("SetTrackMatteSource: cross-comp matte not allowed — src in %q, layer in %q", src.runtime.comp.Name, l.runtime.comp.Name)
	}
	if src.ID == l.ID {
		return fmt.Errorf("SetTrackMatteSource: self-matte (sourceID == own ID = %d) not allowed", l.ID)
	}
	return l.SetTrackMatteLayer(src.ID, mode)
}
