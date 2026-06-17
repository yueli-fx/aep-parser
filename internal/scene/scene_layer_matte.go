package scene

import "fmt"

// SetTrackMatteSource designates `src` as this layer's explicit track-matte
// source and writes the matte mode. Mirrors AE ScriptingAPI 23+
// layer.setTrackMatte(srcLayer, type). Requires AE 23+ ldta (the @0xA0
// slot — re-save through AE 23+ first on AE 22 / 2020 files).
//
// On success, writes ldta @0xA0..0xA3 = src.ID (4 bytes BE) and
// @0x6B = mode (1 byte). length-preserving. Delegates byte writes to
// SetTrackMatteLayer; this wrapper provides the *Layer-arg parity with AE
// scripting plus early validation (nil / cross-comp / self) with clearer
// error messages.
//
// Refuse-cases:
//   - src == nil
//   - either layer lacks a comp back-pointer (built outside parser)
//   - cross-comp matte (src.comp != l.comp); AE 23+ requires same-comp
//   - self-matte (src.ID == l.ID)
//   - ldta too short — caught and surfaced by SetTrackMatteLayer
//
// Pass mode=TrackMatteNone with non-nil src for "preserve target, no
// matte applied" — AE allows that (source pointer stored, matte channel
// disabled). To fully clear, use ClearTrackMatteLayer().
//
//aep:cap domain=layer-set tier=stable verify=roundtrip minver=2023 boundary="length-preserving;AE 23+ ldta 专属;委托 SetTrackMatteLayer;nil/跨 comp/自 matte 拒绝;AE2025 单版本已验(track_matte_explicit,DOM trackMatteLayer+type 对);双版本不可达(fingerprint,同 SetTrackMatteLayer)→ 留 roundtrip" alias="set track matte source,track matte,setTrackMatte,设置遮罩来源,遮罩图层"
func (l *Layer) SetTrackMatteSource(src *Layer, mode TrackMatteType) error {
	if src == nil {
		return fmt.Errorf("SetTrackMatteSource: matte source is nil")
	}
	if l.comp == nil {
		return fmt.Errorf("SetTrackMatteSource: layer %q has no comp back-ref (built outside parser?)", l.Name)
	}
	if src.comp == nil {
		return fmt.Errorf("SetTrackMatteSource: matte source %q has no comp back-ref (built outside parser?)", src.Name)
	}
	if l.comp != src.comp {
		return fmt.Errorf("SetTrackMatteSource: cross-comp matte not allowed — src in %q, layer in %q", src.comp.Name, l.comp.Name)
	}
	if src.ID == l.ID {
		return fmt.Errorf("SetTrackMatteSource: self-matte (sourceID == own ID = %d) not allowed", l.ID)
	}
	return l.SetTrackMatteLayer(src.ID, mode)
}
