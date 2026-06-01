package aep

import "encoding/binary"

// write_render_queue.go — length-preserving in-place setters for render queue
// item settings. Each patches a fixed-width field inside the 2246-byte
// render-settings ldat block (settingsBlock aliases the chunk bytes), so no
// chunk size changes and WriteAEP re-emits the mutation. Non-structural, so no
// AE ship-gate (same model as the other Set* value patches).
//
// Alpha: this write surface has not been double-version ship-gated. Values
// round-trip byte-stably; AE acceptance is presumed (fields AE itself writes)
// but not yet validated in-app.
//
// Concurrency: like all Set* patches these mutate shared chunk bytes; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

// patchU16 writes a big-endian u16 at the given field offset in the item's
// settings block, no-op when the item has no backing block.
func (it *RenderQueueItem) patchU16(fieldOffset int, v uint16) bool {
	if it == nil || len(it.settingsBlock) < fieldOffset+2 {
		return false
	}
	binary.BigEndian.PutUint16(it.settingsBlock[fieldOffset:], v)
	return true
}

// sentinelU16 maps the "current settings" value -1 back to the binary 0xFFFF
// sentinel; other values pass through truncated to u16.
func sentinelU16(v int) uint16 {
	if v < 0 {
		return 0xFFFF
	}
	return uint16(v)
}

// --- render settings setters ------------------------------------------------

// SetQuality sets the render quality (-1 current / 0 wireframe / 1 draft /
// 2 best). Alpha.
func (it *RenderQueueItem) SetQuality(v int) {
	if it.patchU16(rsQuality, sentinelU16(v)) {
		it.RenderSettings.Quality = v
	}
}

// SetColorDepth sets the color depth (-1 current / 0 8bpc / 1 16bpc / 2 32bpc).
func (it *RenderQueueItem) SetColorDepth(v int) {
	if it.patchU16(rsColorDepth, sentinelU16(v)) {
		it.RenderSettings.ColorDepth = v
	}
}

// SetEffects sets the effects render setting (0 all-off / 1 all-on / 2 current).
func (it *RenderQueueItem) SetEffects(v int) {
	if it.patchU16(rsEffects, sentinelU16(v)) {
		it.RenderSettings.Effects = v
	}
}

// SetFieldRender sets field rendering (0 off / 1 upper-first / 2 lower-first).
func (it *RenderQueueItem) SetFieldRender(v int) {
	if it.patchU16(rsFieldRender, sentinelU16(v)) {
		it.RenderSettings.FieldRender = v
	}
}

// SetPulldown sets the 3:2 pulldown phase (0 off / 1..5).
func (it *RenderQueueItem) SetPulldown(v int) {
	if it.patchU16(rsPulldown, sentinelU16(v)) {
		it.RenderSettings.Pulldown = v
	}
}

// SetFrameBlending sets frame blending (0 off-all / 1 on-checked / 2 current).
func (it *RenderQueueItem) SetFrameBlending(v int) {
	if it.patchU16(rsFrameBlending, sentinelU16(v)) {
		it.RenderSettings.FrameBlending = v
	}
}

// SetMotionBlur sets motion blur (0 off-all / 1 on-checked / 2 current).
func (it *RenderQueueItem) SetMotionBlur(v int) {
	if it.patchU16(rsMotionBlur, sentinelU16(v)) {
		it.RenderSettings.MotionBlur = v
	}
}

// SetProxyUse sets proxy use (0 none / 1 all / 2 current / 3 comp-only).
func (it *RenderQueueItem) SetProxyUse(v int) {
	if it.patchU16(rsProxyUse, sentinelU16(v)) {
		it.RenderSettings.ProxyUse = v
	}
}

// SetSoloSwitches sets solo switches (0 off / 2 current).
func (it *RenderQueueItem) SetSoloSwitches(v int) {
	if it.patchU16(rsSoloSwitches, sentinelU16(v)) {
		it.RenderSettings.SoloSwitches = v
	}
}

// SetGuideLayers sets guide layers (0 off / 2 current).
func (it *RenderQueueItem) SetGuideLayers(v int) {
	if it.patchU16(rsGuideLayers, sentinelU16(v)) {
		it.RenderSettings.GuideLayers = v
	}
}

// SetDiskCache sets disk cache (0 read-only / 2 current).
func (it *RenderQueueItem) SetDiskCache(v int) {
	if it.patchU16(rsDiskCache, sentinelU16(v)) {
		it.RenderSettings.DiskCache = v
	}
}

// SetFrameRate sets the frame-rate source (0 use comp / 1 use this).
func (it *RenderQueueItem) SetFrameRate(v int) {
	if it.patchU16(rsUseThisFrameRate, sentinelU16(v)) {
		it.RenderSettings.FrameRate = v
	}
}

// SetResolution sets the [x, y] resolution divisors (>= 1).
func (it *RenderQueueItem) SetResolution(x, y int) {
	if x < 1 || y < 1 {
		return
	}
	if it.patchU16(rsResolutionX, uint16(x)) && it.patchU16(rsResolutionY, uint16(y)) {
		it.RenderSettings.Resolution = [2]int{x, y}
	}
}

// SetSkipExistingFiles toggles "skip existing files".
func (it *RenderQueueItem) SetSkipExistingFiles(v bool) {
	n := uint16(0)
	if v {
		n = 1
	}
	if it.patchU16(rsSkipExistingFiles, n) {
		it.RenderSettings.SkipExistingFiles = v
	}
}
