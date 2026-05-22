// internal/aep/framerate_canonical.go
package aep

import "math"

// frameRateEncoding 表示 cdta @0x9C..@0x9F 的 (whole, frac) 对编码。
// AE 存 frame rate 为 whole + frac/65536，但 NTSC fractions（23.976/29.97/59.94）
// 必须用 AE 内部 canonical 值，否则 AE 重新保存时 rewrite + 测试 flaky。
type frameRateEncoding struct {
	whole uint16
	frac  uint16
}

// ntscCanonical: 容差 < ntscTolerance 内的用户 fps 都映射到这些 canonical 值。
// frac 值实证来源:
//   29.97 / 59.94 来自 test_data/re_tickrate.aep RE_fps_29_97 / RE_fps_59_94 cdta @0x9E
//   23.976 数学计算 (24000/1001 - 23) * 65536 ≈ 63962 = 0xF9DA (无 fixture 实证但容差内安全)
const ntscTolerance = 1e-3

var ntscCanonical = map[float64]frameRateEncoding{
	23.976: {whole: 23, frac: 0xF9DA},
	29.97:  {whole: 29, frac: 0xF852},
	59.94:  {whole: 59, frac: 0xF0A4},
}

// encodeFrameRate 把用户输入 fps（如 29.97 或 30000/1001）normalize 为
// canonical (whole, frac) 二元组写入 cdta。
//
// 行为：
//   - 容差内匹配 ntscCanonical 表 → 用 canonical 值
//   - 否则走通用 round：whole = floor(fps), frac = round((fps - whole) * 65536)
//
// 返回的 (whole, frac) 写入 cdta @cdtaFrameRateWhole / @cdtaFrameRateFrac。
func encodeFrameRate(fps float64) frameRateEncoding {
	for canonicalFps, enc := range ntscCanonical {
		if math.Abs(fps-canonicalFps) < ntscTolerance {
			return enc
		}
	}
	whole := uint16(math.Floor(fps))
	frac := uint16(math.Round((fps - float64(whole)) * 65536))
	return frameRateEncoding{whole: whole, frac: frac}
}

// decodeFrameRate 反向解（给单测验证 round-trip）。
func decodeFrameRate(enc frameRateEncoding) float64 {
	return float64(enc.whole) + float64(enc.frac)/65536.0
}

// fpsTiming 是 cdta 里 fps-依赖的 4 个 secondary 字段集合。AE 25 写新 comp 时
// 这些字段必须填充正确值（实测 RE_fps_* fixtures from test_data/re_tickrate.aep），
// 否则 AE 25 打开会 crash —— parser 不读这些字段但 AE 自身做时间轴 sanity 检查时
// 会用到。
//
// 字段 RE 来源: dump_cdta test_data/re_tickrate.aep 各 RE_fps_* comp。
type fpsTiming struct {
	ticksPerFrame uint16 // cdta @0x06 (uint16 BE)
	tickRate      uint32 // cdta @0x08 / @0x18 / @0x30 (uint32 BE; = ticksPerFrame × fps)
	masterTicks   uint32 // cdta @0x2C (uint32 BE; = ticksPerFrame × 5 × fps_nominal_whole)
}

// canonicalFpsTiming 是 AE 25 实测 fps → cdta secondary fields 表。
// 验证: test_data/re_tickrate.aep 各 RE_fps_* comp 的 cdta @0x04..@0x33。
//
// NTSC nominal_whole: 23.976→24 / 29.97→30 / 59.94→60（@0x2C 计算用此值，不是 floor(fps)）。
//
// 23.976 行为推算（fixture 缺）: ticksPerFrame=1000 保持 TickRate ≈ 23976 与 29.97/59.94
// 共享 NTSC 时间基。不实证但 mathematically consistent，且 ship gate 测 29.97 不依赖。
var canonicalFpsTiming = map[float64]fpsTiming{
	24:     {ticksPerFrame: 1024, tickRate: 24576, masterTicks: 122880},
	25:     {ticksPerFrame: 1024, tickRate: 25600, masterTicks: 128000},
	30:     {ticksPerFrame: 1024, tickRate: 30720, masterTicks: 153600},
	50:     {ticksPerFrame: 512, tickRate: 25600, masterTicks: 128000},
	60:     {ticksPerFrame: 512, tickRate: 30720, masterTicks: 153600},
	23.976: {ticksPerFrame: 1000, tickRate: 23976, masterTicks: 120000}, // computed, no fixture
	29.97:  {ticksPerFrame: 800, tickRate: 23976, masterTicks: 120000},
	59.94:  {ticksPerFrame: 400, tickRate: 23976, masterTicks: 120000},
}

// lookupFpsTiming 返回 fps 对应的 cdta timing 字段。
//
// 行为:
//   - canonical fps（容差 ntscTolerance 内）→ 表查找
//   - 其它 fps → 启发式:
//     ticksPerFrame = 1024 if fps ≤ 30 else 512
//     tickRate = round(ticksPerFrame × fps)
//     masterTicks = ticksPerFrame × 5 × round(fps)
//
// 启发式 fallback 对 AE 25 是否完全 safe 未实证 — 但比全零安全；
// 非 canonical fps 需进一步 RE 才能保证 AE 25 不 crash。
func lookupFpsTiming(fps float64) fpsTiming {
	for canonical, t := range canonicalFpsTiming {
		if math.Abs(fps-canonical) < ntscTolerance {
			return t
		}
	}
	var tpf uint16
	if fps <= 30 {
		tpf = 1024
	} else {
		tpf = 512
	}
	rounded := math.Round(fps)
	return fpsTiming{
		ticksPerFrame: tpf,
		tickRate:      uint32(math.Round(float64(tpf) * fps)),
		masterTicks:   uint32(tpf) * 5 * uint32(rounded),
	}
}
