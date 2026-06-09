package serializer

import "github.com/example/aep-parser/internal/scene"

// syncRenderQueue copies every render-queue item's scene-owned settings buffers
// back into the owning RIFX chunks before serialization (single source of truth
// → chunk). Length-preserving: an unmutated buffer is byte-identical to the
// parsed bytes, so a parse→write round-trip is unchanged. Called by WriteAEP;
// no-op for queues built outside the parser. Free function (serializer stage):
// it reaches the concrete render-queue / output-module back-refs.
func syncRenderQueue(p *Project) {
	rq := p.RenderQueue
	if rq == nil {
		return
	}
	for _, it := range rq.Items {
		itBlock := scene.RenderQueueItemSettings(it)
		if rb := renderQueueItemBack(it); rb != nil && len(rb.settingsSlice) == len(itBlock) {
			copy(rb.settingsSlice, itBlock)
		}
		for _, om := range it.OutputModules {
			ob := outputModuleBack(om)
			if ob == nil {
				continue
			}
			omBlock := scene.OutputModuleSettingsBlock(om)
			if len(ob.settingsSlice) == len(omBlock) {
				copy(ob.settingsSlice, omBlock)
			}
			omRoou := scene.OutputModuleRoouData(om)
			if len(ob.roouSlice) == len(omRoou) {
				copy(ob.roouSlice, omRoou)
			}
		}
	}
}
