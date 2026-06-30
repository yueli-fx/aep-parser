// Code moved from scene_shape_graph.go; keep behavior-only edits out of split commits.
package scene

import (
	"fmt"
)

// PropertyGroup is the escape-hatch surface. Currently the minimal struct —
// Name and the empty Children / streams maps — so the field exists on
// VectorGroup.Transform / node Properties() but none of the typed lookup methods
// are wired yet.
type PropertyGroup struct {
	Name      string
	Children  map[string]*PropertyGroup
	streams   map[string]any
	Separated bool // dimension separation reserved for nested-group support
}

// newGroupTransform constructs the identity Transform PropertyGroup placeholder
// for a fresh VectorGroup. A RootGroup default-serialized form has no
// `ADBE Vector Transform Group` — this placeholder stays nil-children until the
// escape hatch wires it.
func newGroupTransform() *PropertyGroup {
	return &PropertyGroup{Name: "Transform"}
}

// Child returns the nested PropertyGroup by name, or nil if not present.
// Use for walking deeper-than-leaf escape-hatch trees.
func (pg *PropertyGroup) Child(name string) *PropertyGroup {
	if pg == nil || pg.Children == nil {
		return nil
	}
	return pg.Children[name]
}

// Float64Stream returns the PropertyStream[float64] under the given name,
// or an error if no stream by that name exists or it isn't the expected type.
// Mutations on the returned stream are visible through the typed accessor.
func (pg *PropertyGroup) Float64Stream(name string) (*PropertyStream[float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not float64", pg.Name, name)
	}
	return ps, nil
}

// Vec2Stream returns the PropertyStream[[2]float64] under the given name.
func (pg *PropertyGroup) Vec2Stream(name string) (*PropertyStream[[2]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[2]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [2]float64", pg.Name, name)
	}
	return ps, nil
}

// Vec3Stream returns the PropertyStream[[3]float64] under the given name.
func (pg *PropertyGroup) Vec3Stream(name string) (*PropertyStream[[3]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[3]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [3]float64", pg.Name, name)
	}
	return ps, nil
}

// ColorStream returns the PropertyStream[[4]float64] (RGBA) under the
// given name.
func (pg *PropertyGroup) ColorStream(name string) (*PropertyStream[[4]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[4]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [4]float64", pg.Name, name)
	}
	return ps, nil
}

// PathStream returns the PropertyStream[BezierPath] under the given
// name.
func (pg *PropertyGroup) PathStream(name string) (*PropertyStream[BezierPath], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[BezierPath])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not BezierPath", pg.Name, name)
	}
	return ps, nil
}
