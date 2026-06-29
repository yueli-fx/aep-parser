package aeptest

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
	"github.com/yueli-fx/aep-parser/internal/serializer"
)

func NextItemID(p *aep.Project) uint32 { return scene.ProjectNextItemID(p) }

func RootFold(p *aep.Project) *rifx.Chunk { return serializer.ProjectRootFold(p) }

func ItemList(c *aep.Composition) *rifx.Chunk { return serializer.CompItemListForTest(c) }

func Root(p *aep.Project) *rifx.Chunk { return serializer.ProjectRoot(p) }

func Ldta(l *aep.Layer) *rifx.Chunk { return serializer.LayerLdta(l) }

func NewTestProperty(matchName string, components int, tdb4, tdsb, tdum, tduM *rifx.Chunk) *aep.Property {
	return serializer.NewTestProperty(matchName, components, tdb4, tdsb, tdum, tduM)
}

func MakeTdb4(dims uint16, spatialStatic, cvot, noValue, typeFlags byte) *rifx.Chunk {
	return serializer.MakeTdb4(dims, spatialStatic, cvot, noValue, typeFlags)
}

func SetLayerComp(l *aep.Layer, c *aep.Composition) { scene.SetLayerComp(l, c) }

func SetCompProj(c *aep.Composition, p *aep.Project) { scene.SetCompositionProj(c, p) }

func ClearLayerLayrList(l *aep.Layer) { serializer.ClearLayerLayrList(l) }

func CorruptSrcLayrFormType(l *aep.Layer) { serializer.CorruptSrcLayrFormType(l) }

func Proj(c *aep.Composition) *aep.Project { return scene.CompositionProj(c) }

func DestFootageByPath(p *aep.Project, path string) *aep.Footage {
	return serializer.DestFootageByPath(p, path)
}

func LocateItemBlockByID(p *aep.Project, id uint32) (int, int) {
	return serializer.LocateItemBlockByID(p, id)
}

func ImportFootageBlock(dest, src *aep.Project, srcID uint32, name string) (uint32, error) {
	return serializer.ImportFootageBlock(dest, src, srcID, name)
}
