package aep

import (
	"io"

	"github.com/example/aep-parser/internal/serializer"
)

// Application wraps a parsed Project, mirroring py-aep's top-level
// `app = py_aep.parse("x.aep")` entry. Existing Go-style entry points
// (`aep.Open`, `aep.FromReader` returning *Project) are unchanged; this
// type is an additive convenience layer.
//
// Application exists primarily to provide a stable namespace for
// AE-version metadata (Version), file-level handles (Project), and
// future runtime-style attributes that don't belong on Project itself.
type Application struct {
	Project *Project
}

// Parse opens an .aep file at path and returns an Application wrapping
// the parsed Project. Mirrors py-aep's `py_aep.parse(path)`.
//
//aep:cap domain=meta tier=stable verify=roundtrip alias="parse,application,read,py-aep parse,读取"
func Parse(path string) (*Application, error) {
	p, err := Open(path)
	if err != nil {
		return nil, err
	}
	return &Application{Project: p}, nil
}

// ParseReader parses an .aep file from an io.ReadSeeker and returns an
// Application wrapping the Project. Mirrors py-aep's reader-based parse.
//
//aep:cap domain=meta tier=stable verify=roundtrip alias="parse reader,application,流读取,io.ReadSeeker"
func ParseReader(r io.ReadSeeker) (*Application, error) {
	p, err := FromReader(r)
	if err != nil {
		return nil, err
	}
	return &Application{Project: p}, nil
}

// Version returns the AE version that wrote the file as a string like
// "17.7x101" (AE 2020 / 17.7 build 101) or "25.6x57" (AE 2025), decoded
// from the root `head` chunk's packed version word (offset 4-7).
//
// Layout per byte 4 (uint32 BE):
//
//	bits 30-26 : major_a   (5 bits)
//	bits 21-19 : major_b   (3 bits)        major = major_a*8 + major_b
//	bits 18-15 : minor     (4 bits)
//	bit  9     : beta flag (1 = release)
//	bits 7-0   : build number
//
// Returns "" when the file has no head chunk or its data is too short
// (synthesized fixtures / corrupt files).
//
// Source: py-aep `binary/item_chunks.py::HeadChunk`.
//
//aep:cap domain=meta tier=stable verify=roundtrip alias="version,AE 版本,build,head chunk"
func (a *Application) Version() string {
	if a.Project == nil {
		return ""
	}
	return serializer.ProjectVersionString(a.Project)
}
