// flightdeck/showcase/booyah-clone/main.go — rebuild the entire "Booyah Glitch"
// project from scratch via our Go API, DAG leaf-first, the original as a read-only
// value oracle. Plan: flightdeck/plans/2026-06-19-booyah-glitch-replication.md.
//
// Comps are assembled in DAG topological order (leaves first); each gen_<comp>.go
// adds one composition. Run from repo root: `go run ./flightdeck/showcase/booyah-clone`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/booyah-clone/booyah-clone.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	orc, err := openOriginal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := aep.NewProject(aep.TargetAE2020)

	// --- DAG topological order (leaves first) ---
	buildShapeIiip(p, orc) // ① シェイイイイプ！！！
	// ② テキスト変えるならココ！ / ③ マップ用フラクタルノイズ / ④ カクッ ... (later tasks)

	f, err := os.Create(outPath)
	must(err)
	defer f.Close()
	must(p.WriteAEP(f))
	fmt.Println("wrote", outPath)
}
