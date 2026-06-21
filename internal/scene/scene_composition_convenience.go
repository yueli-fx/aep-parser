package scene

// Composition convenience helpers — mirroring AE's CompItem accessors
// (num layers, has audio, time scale). `ActiveCamera` and `Markers` already
// live on Composition itself (ActiveCamera as a method in types_core.go;
// Markers as a direct field).

// NumLayers returns the number of layers in the composition.
// Equivalent to len(c.Layers); provided as a convenience alias.
func (c *Composition) NumLayers() int {
	return len(c.Layers)
}

// HasAudio reports whether any layer in the composition has its audio
// switch enabled (ldta @0x27 bit 1). It does NOT verify the layer's
// source actually carries an audio track — only that AE would attempt
// audio playback for it.
func (c *Composition) HasAudio() bool {
	for _, l := range c.Layers {
		if l.AudioEnabled {
			return true
		}
	}
	return false
}

// TimeScale is the per-comp ticks-per-second base used for keyframe and
// marker times. Alias of TickRate, provided as a naming convenience.
// See [Composition.TickRate] for the full notes on cdta @0x08 / @0xA8
// decoding.
func (c *Composition) TimeScale() float64 {
	return c.TickRate
}
