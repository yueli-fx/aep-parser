package projectindex

import (
	"fmt"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func BenchmarkBuild(b *testing.B) {
	project := benchmarkProject(50, 200)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Build(project)
	}
}

func BenchmarkRepeatedLookups(b *testing.B) {
	idx := Build(benchmarkProject(50, 200))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = idx.AVItemByID(uint32(i%50 + 1))
		_ = idx.LayersByName(fmt.Sprintf("Layer-%d", i%200))
		_ = idx.LayersBySourceID(uint32(i%50 + 1))
		_ = idx.LayersByEffect("ADBE Glo2")
	}
}

func benchmarkProject(compCount, layersPerComp int) *aep.Project {
	project := &aep.Project{}
	for i := 0; i < compCount; i++ {
		id := uint32(i + 1)
		project.Compositions = append(project.Compositions, &aep.Composition{
			ID:   id,
			Name: fmt.Sprintf("Comp-%d", i),
		})
		project.Footage = append(project.Footage, &aep.Footage{
			ID:   id + uint32(compCount),
			Name: fmt.Sprintf("Footage-%d", i),
		})
	}
	for ci, comp := range project.Compositions {
		for li := 0; li < layersPerComp; li++ {
			layer := &aep.Layer{
				ID:       uint32(li + 1),
				Name:     fmt.Sprintf("Layer-%d", li),
				SourceID: uint32(li%compCount + 1),
			}
			if (ci+li)%3 == 0 {
				layer.Effects = []*aep.Effect{{MatchName: "ADBE Glo2"}}
			}
			comp.Layers = append(comp.Layers, layer)
		}
	}
	return project
}
