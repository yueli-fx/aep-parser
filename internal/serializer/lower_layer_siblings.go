// internal/aep/lower_layer_siblings.go
//
// The "layer sibling chunks" serializer primitive. AE writes each user Layr
// inside a comp Item LIST as: the Layr LIST, an empty Ewst LIST, then two
// repetitions of fvdv / fiop / ftts / foac / fiac / fipc / fifl. The Ewst
// alone lets AE accept a single user layer, but every layer past the first is
// silent-dropped from comp.layers when the fvdv… siblings are absent — AE uses
// them to delimit one layer's serialized unit from the next.
//
// Byte values are constants lifted from AE-saved multi-shape-layer baselines
// (re_multi_shape_layer.aep); they are identical across layers and AE versions
// (AE 2020 baseline opens unchanged in AE 2025), mirroring the Fold-level
// item-sibling primitive in lower_item_siblings.go.
package serializer

import "github.com/yueli-fx/aep-parser/internal/rifx"

// lowerLayerSiblings returns the 14 sibling chunks (two 7-chunk groups) that
// follow each user Layr's Ewst inside a comp Item LIST.
func lowerLayerSiblings() []*rifx.Chunk {
	group := func() []*rifx.Chunk {
		return []*rifx.Chunk{
			{ID: rifx.ChunkID{'f', 'v', 'd', 'v'}, Data: []byte{0x00, 0x00, 0x00, 0x03}},
			{ID: rifx.ChunkID{'f', 'i', 'o', 'p'}, Data: []byte{0x00}},
			{ID: rifx.ChunkID{'f', 't', 't', 's'}, Data: []byte{0x00, 0x00, 0x00, 0x00}},
			{ID: rifx.ChunkID{'f', 'o', 'a', 'c'}, Data: []byte{0x00}},
			{ID: rifx.ChunkID{'f', 'i', 'a', 'c'}, Data: []byte{0x00}},
			{ID: rifx.ChunkID{'f', 'i', 'p', 'c'}, Data: []byte{0x00, 0x00}},
			{ID: rifx.ChunkID{'f', 'i', 'f', 'l'}, Data: []byte{0x00, 0x00, 0x00, 0x00}},
		}
	}
	return append(group(), group()...)
}
