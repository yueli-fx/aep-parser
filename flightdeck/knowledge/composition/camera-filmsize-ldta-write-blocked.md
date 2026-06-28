# ⚠ Camera FilmSize 在 ldta @0x98 但 ScriptingAPI 不可写

SUMMARY: Camera FilmSize 在 ldta @0x98 但 ScriptingAPI 不可写
READ WHEN: adding Camera layer setter; RE'ing ldta @0x98..0x9F; tempted to ship a SetFilmSize / SetSensorSize API

---

## 现象

dump_ldta_trackmatte 抓 AE 2025 fresh camera：

```
00000090  00 00 00 00 00 00 00 01  40 59 83 06 0c 18 30 62  |........@Y....0b|
```

`@0x98..0x9F` = `40 59 83 06 0c 18 30 62` = float64 BE = **102.0472440944882**

= **36 mm × 72 PPI**（AE 默认 Film Size 36mm 转 pixels @ 72 DPI 单位）。

非 camera 层（light / AV / shape / text）这 8 字节全 0。

## 强烈疑似

camera 的 **Film Size**（Camera Settings 对话框里的字段，可改 16mm / 35mm Full Frame / 65mm / IMAX / custom 等）。

## 但写路径堵死

AE 2025 中文版的 `cameraOption` PropertyGroup（match name `ADBE Camera Options Group`）只有 13 个 sub-property：

```
ADBE Camera Zoom / Depth of Field / Focus Distance / Aperture / Blur Level
ADBE Iris Shape / Rotation / Roundness / Aspect Ratio / Diffraction Fringe
       Highlight Gain / Threshold / Saturation
```

**没有 Film Size**。`property("ADBE Camera Film Size")` 抛错。

JSX `cam.cameraOption.filmSize = 50` 不报错但 no-op（ExtendScript 允许任意 key 赋值）；ldta @0x98 字节零变化。

## 结论

跟 variable-font axes 的脚本限制同模式：AE 把字段持久化到 .aep 但 ScriptingAPI 把它隐藏。**只能读，不能 length-preservingly 写**（除非直接修 ldta 字节，但没 fixture 验证就上未免太冒险）。

若用户真要写 Film Size：

1. 用户自己在 AE Camera Settings 对话框改（UI-only）
2. 我们可暴露 `Layer.SetCameraFilmSize(mm float64)` 直接写 ldta @0x98（8 字节 length-preserving）+ 在文档里标 ⚠ unverified —— 但用户得在 AE 里打开看视觉验证，对错没法字节级 confirm

**当前选择**：不 ship setter，记 scar。如果未来 AE 版本暴露了对应 ScriptingAPI（或用户找到非 ScriptingAPI 的写路径如 AEGP / C++ SDK），再加。

## Reader 倒是没问题

可以暴露只读 accessor：

```go
// Layer.CameraFilmSize returns camera film size in pixels (36mm @ 72PPI default).
// Only present on Layer.Type == LayerTypeCamera; returns 0 otherwise.
func (l *Layer) CameraFilmSize() float64 { ... }
```

把 102.0472 这种 magic number 翻译给用户用 `mm := px / 72 * 25.4` 或类似。但即便如此读了也没法写，价值有限。**暂时也不 ship**。

## 提示

如果 AE 24+ 文档发现 `cameraOption` 隐藏字段（类似 TextDocument 那种 dict-form），重测。当前 AE 2025 中文版没出。
