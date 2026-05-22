package aep_test

import (
	"testing"
)

// TestBuildCompItemMatchesGoldenStructure 是 Phase 4.7 占位 stub。
// 完整验证 builder 输出 Item LIST children 顺序 / 类型跟 AE 2025 saved 1-comp
// fixture (test_data/comp_item_children.golden.txt) 一致，需要 NewComposition
// 入口（Phase 5）建出 *Composition 才能拿到 itemList chunk 对比。
//
// 当前 Phase 4 builders 已经实现完整 buildCompItem 链路，但 unexported；
// Phase 5 TestNewComposition_Roundtrip 通过 WriteAEP + FromReader 隐式验证
// children 顺序正确（错了 AE 会拒开 / re-parse 会丢字段）。
//
// 这个 skip-stub 留作未来如果需要 explicit byte-level golden diff 时的占位。
func TestBuildCompItemMatchesGoldenStructure(t *testing.T) {
	t.Skip("Deferred to Phase 5: NewComposition roundtrip implicitly verifies golden order")
}
