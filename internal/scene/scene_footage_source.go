package scene

// FootageSource is the interface for typed footage source metadata.
// Implemented by *FileSource, *SolidSource, and *PlaceholderSource.
type FootageSource interface {
	SourceType() string
}

// FileSource describes footage that comes from a file.
type FileSource struct {
	Path      string   // full file path
	FileNames []string // per-frame filenames for image sequences (empty for single files)
}

// SourceType returns "file".
func (f *FileSource) SourceType() string { return "file" }

// SolidSource describes a solid-color footage source.
type SolidSource struct {
	Color [3]float64 // RGB color in 0..1 range
}

// SourceType returns "solid".
func (s *SolidSource) SourceType() string { return "solid" }

// PlaceholderSource describes a placeholder footage source.
type PlaceholderSource struct{}

// SourceType returns "placeholder".
func (p *PlaceholderSource) SourceType() string { return "placeholder" }

// MainSource returns the typed source metadata for this footage item.
// Returns *FileSource for file footage, *SolidSource for solids,
// *PlaceholderSource for placeholders. Returns nil if the footage
// was built outside the parser (no source chunks available).
func (f *Footage) MainSource() FootageSource {
	switch {
	case f.IsPlaceholder:
		return &PlaceholderSource{}
	case f.IsSolid:
		// Solid color is stored in the opti chunk. We don't currently
		// parse the solid color from the binary, so return a zero-value
		// SolidSource. The color can be read from the raw sspc chunk
		// if needed.
		return &SolidSource{}
	default:
		return &FileSource{
			Path: f.Path,
		}
	}
}
