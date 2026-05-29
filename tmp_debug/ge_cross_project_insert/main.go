package main

import (
	"fmt"
	"os"
	"path/filepath"

	aep "github.com/example/aep-parser/internal/aep"
)

// ge_cross_project_insert opens re_xproj_src_<mode> + re_xproj_dest_<mode>,
// inserts the src comp's first AV layer into the dest comp (cross-Project),
// writes ge_cross_project_insert_<mode>.aep. Go-side assertions guard before
// write; AE acceptance is verified by verify_ge_cross_project_insert.jsx.
func main() {
	dir := `e:\projects\tools\aep-parser\test_data`
	for _, mode := range []string{"footage", "precomp", "dedup"} {
		srcProj, err := aep.Open(filepath.Join(dir, "re_xproj_src_"+mode+".aep"))
		must(err, mode, "open src")
		destProj, err := aep.Open(filepath.Join(dir, "re_xproj_dest_"+mode+".aep"))
		must(err, mode, "open dest")

		srcComp := srcProj.CompositionByName("src_comp")
		destComp := destProj.CompositionByName("dest_comp")
		if srcComp == nil || destComp == nil {
			fail(mode, "src_comp/dest_comp not found")
		}
		var src *aep.Layer
		for _, l := range srcComp.Layers {
			if l.Type == aep.LayerTypeAV {
				src = l
				break
			}
		}
		if src == nil {
			fail(mode, "no AV layer in src_comp")
		}
		preFootage := len(destProj.Footage)
		clone, err := destComp.InsertLayer(src, 0)
		must(err, mode, "InsertLayer")
		if destProj.AVItemByID(clone.SourceID) == nil {
			fail(mode, fmt.Sprintf("clone.SourceID=%d unresolved in dest", clone.SourceID))
		}
		if mode == "dedup" && len(destProj.Footage) != preFootage {
			fail(mode, "footage duplicated in dedup mode")
		}
		out := filepath.Join(dir, "ge_cross_project_insert_"+mode+".aep")
		f, err := os.Create(out)
		must(err, mode, "create out")
		must(destProj.WriteAEP(f), mode, "WriteAEP")
		f.Close()
		fmt.Printf("[%s] OK -> %s (clone id=%d src=%d)\n", mode, out, clone.ID, clone.SourceID)
	}
}

func must(err error, mode, what string) {
	if err != nil {
		fail(mode, what+": "+err.Error())
	}
}
func fail(mode, msg string) {
	fmt.Fprintf(os.Stderr, "[%s] FAIL: %s\n", mode, msg)
	os.Exit(1)
}
