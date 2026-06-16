// Package sample is a capindex extraction fixture (not built into the binary).
package sample

// Tagged does a thing worth indexing.
//
//aep:cap domain=layer-create tier=stable verify=ae-accept gate=TestFoo alias="测试,thing"
func Tagged(name string) error { return nil }

// Untagged has prose but no aep:cap directive.
func Untagged() {}
