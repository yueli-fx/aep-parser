package scene

// scene_guide.go — composition ruler guides (AE's drag-from-ruler alignment
// lines). UI-only: they don't affect rendering. Stored in the comp's Item-level
// LIST:Gide → LIST:list → lhd3(count) + ldat(GuideItem × N, 16B each). No
// ExtendScript equivalent; mirrors py-aep's Guide model. //nolint:jargon

// GuideOrientation is a guide's binary orientation code (as stored in ldat @0).
type GuideOrientation uint32

const (
	// GuideHorizontal is a horizontal guide — Position is pixels from the top.
	GuideHorizontal GuideOrientation = 2
	// GuideVertical is a vertical guide — Position is pixels from the left.
	GuideVertical GuideOrientation = 1
)

// String renders the orientation as "horizontal" / "vertical" (or the raw code).
func (o GuideOrientation) String() string {
	switch o {
	case GuideHorizontal:
		return "horizontal"
	case GuideVertical:
		return "vertical"
	default:
		return "unknown"
	}
}

// Guide is one composition ruler guide.
type Guide struct {
	// Orientation is GuideHorizontal (2) or GuideVertical (1).
	Orientation GuideOrientation

	// Position is the guide's offset in pixels — from the top edge for a
	// horizontal guide, from the left edge for a vertical guide.
	Position float64

	// block is this guide's scene-owned copy of its 16-byte GuideItem — the
	// single source of truth. The length-preserving setters mutate this copy;
	// syncGuides copies it back into the owning comp's Gide ldat slot (paired by
	// index) at WriteAEP time. Nil for guides built outside the parser.
	block []byte
}
