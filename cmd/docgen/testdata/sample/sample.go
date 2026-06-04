// Package sample 是 docgen 的测试固定输入，覆盖 字段/getter/方法/常量/JSON tag/directive/Example。
package sample

import "fmt"

// Widget 是一个示例 widget。
//
// 第二段：演示 docgen 抽取多段 prose。
type Widget struct {
	// Name 是显示名。
	Name string `json:"name"`

	// Tags 是标签。
	Tags []string `json:"tags"`

	hidden int // 未导出，应被忽略
}

// SetName 设置 Name。
//
// 当 name 为空时返回 error。
func (w *Widget) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	w.Name = name
	return nil
}

// Size 返回标签数。
func (w *Widget) Size() int { return len(w.Tags) }

// Clone 返回副本。
//
//docgen:method
func (w *Widget) Clone() *Widget { return &Widget{Name: w.Name} }

// Mode 是 widget 模式。
type Mode int

// Widget 模式常量。
const (
	ModeOff Mode = iota
	ModeOn
)
