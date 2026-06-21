package aep

import (
	"io"

	"github.com/example/aep-parser/internal/serializer"
)

// Application wraps a parsed Project, mirroring py-aep's top-level //nolint:jargon
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

// @summary    Open an .aep file and return an Application wrapping the parsed project
// @param      path  the filesystem path to the .aep file
// @returns    an Application wrapping the parsed project
// @domain     meta
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @alias      parse,application,read,读取
func Parse(path string) (*Application, error) {
	p, err := Open(path)
	if err != nil {
		return nil, err
	}
	return &Application{Project: p}, nil
}

// @summary    Parse an .aep file from a reader into an Application
// @param      r  the reader positioned at the start of the .aep file
// @returns    an Application wrapping the parsed project
// @domain     meta
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @alias      parse reader,application,流读取,io.ReadSeeker
func ParseReader(r io.ReadSeeker) (*Application, error) {
	p, err := FromReader(r)
	if err != nil {
		return nil, err
	}
	return &Application{Project: p}, nil
}

// @summary    Report the AE version that wrote the file
// @returns    a version string like "17.7x101" (AE 2020 build 101) or "25.6x57" (AE 2025), or "" when unavailable
// @domain     meta
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @description Decoded from the root head chunk's packed version word at
//   offset 4-7 (uint32 big-endian): bits 30-26 are major_a (5 bits), bits
//   21-19 are major_b (3 bits, major = major_a*8 + major_b), bits 18-15 are
//   minor (4 bits), bit 9 is the beta flag (1 = release), and bits 7-0 are
//   the build number. Returns "" when the file has no head chunk or its
//   data is too short, such as for synthesized fixtures or corrupt files.
// @alias      version,AE 版本,build,head chunk
func (a *Application) Version() string {
	if a.Project == nil {
		return ""
	}
	return serializer.ProjectVersionString(a.Project)
}
