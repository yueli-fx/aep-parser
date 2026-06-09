package main

type symKind int

const (
	kindField  symKind = iota // struct 字段 → Attributes
	kindGetter                // getter 方法 → Attributes
	kindMethod                // 动作方法 → Methods
)

type example struct {
	suffix string // "" 或 "merge"/"separate"
	code   string // 去掉 // Output 后的函数体源码
}

type symbol struct {
	name          string
	kind          symKind
	signature     string // 规范化单行（方法/getter）；字段为 ""
	fieldDecl     string // "Name string"（字段）；其它为 ""
	doc           string // directive 已剥离的 prose
	jsonName      string // struct tag json 名；无则 ""
	readWrite     bool   // Attributes 专用：RW=true / R=false
	fieldRWForced bool   // 字段 R/RW 被 //docgen:rw|ro 显式锁定
	examples      []example
}

type constBlock struct {
	doc  string
	code string // const (...) 块源码（go/printer 渲）
}

type docType struct {
	name       string
	doc        string
	attributes []symbol // 字段（结构声明序）后接 getter（源序）
	methods    []symbol // 动作方法（源序）
	consts     []constBlock
	isAlias    bool // `type X = pkg.Y` 别名（facade 包里的壳）；合并时输给真定义
}
