# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (workshop → flightdeck v1.0 rebrand migration)
**Active focus**: V3 Phase 5 闭环；下一步从三等量候选挑：Phase 5C InsertLayer / Project.DuplicateItem / V2.2.1 ShapeLayer 拓展。

## Next session

**挑下一 milestone**（按用户偏好 / 当下兴趣，三个等量候选）：

**Phase 5 残余候选**（从 `specs/2026-05-27-v3-deep-think.md` § 4 picklist）：
- **`Composition.InsertLayer(src *Layer, atIdx int)`** — 跨 comp deep-clone；扩展 DuplicateLayer 到 cross-tree（要解 SourceID 冲突 / cross-comp footage ref）
- **`Project.DuplicateItem(item Item, name string)`** — comp / footage / folder 通用；扩展 V2.1 NewComposition + V3 DuplicateLayer 到 item 级
- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 / keyframe 持久化）

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
