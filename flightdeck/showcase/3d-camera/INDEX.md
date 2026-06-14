---
showcase: 3d-camera
direction: 从零（无 AE）生成一个 3D 图层 + 相机场景，一帧同时展示 Z 视差与 Rotate Y 透视 tumble
capabilities: [is3d-enable, layer-position-z, parallax, rotate-y, perspective-tumble, new-camera-layer, camera-position]
gates: [TestLayer3DEnable_AEShipGate, TestLayer3DCamDolly_AEShipGate, TestLayer3DParallax_AEShipGate, TestLayer3DRotateY_AEShipGate]
status: 待review
last_updated: 2026-06-15
regenerate: "go run ./flightdeck/showcase/3d-camera  +  AE render.jsx"
---

# 3d-camera — showcase

## 这个方向测什么

纯 Go 从零构建一个 3D 场景：4 张同尺寸（300px）卡片 flip 成 3D 图层，放在不同深度，
一个 Go 创建的相机统观。**整个 3D transform group（enable + Z + 旋转）写入零新 serializer
代码**——`SetIs3D` + reopen 后既有 `SetPosition([x,y,z])` / `SetRotateY`，因嵌入的 transform
模板本就是完整 6-axis 3D schema（详 `incidents/layer-3d-enable-bit-materializes.md`）。一帧里
同时验两个可见效果：**Z 视差**（同尺寸卡按深浅渲成大小不同）+ **Rotate Y 透视 tumble**（卡渲成梯形）。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go) | 纯 Go 构建 3d_camera.aep（6 层：BG + 4 卡 + 相机） |
| render.jsx | 渲染脚本 | AE 打开 + 读回 3D/Z/RotateY + saveFrameToPng → 3d_camera.png |
| 3d_camera.aep | 产出工程 (gitignored) | 1920×1080 / 30fps / 6 层 |
| 3d_camera.png | 渲染帧 (gitignored) | frame 0 视觉 |

## 布局（frame 0，相机 @[960,540,-1700]）

| 层 | 颜色 | Position (x,y,z) | RotateY | 期望效果 |
|---|---|---|---|---|
| Near | 红 | 560,420,**-500** | 0 | 最近 → 渲染**最大** |
| Mid | 绿 | 960,420,**+400** | 0 | 中等深度 → 中等大小 |
| Far | 蓝 | 1360,420,**+1300** | 0 | 最远 → 渲染**最小** |
| Tumble | 黄 | 960,800,0 | **50°** | 透视前缩 → 渲成**梯形** |
| BG | 深灰 | 2D 满帧 | — | 舞台背景（2D 忽略相机） |

红→绿→蓝的大小递减 = Z 视差；黄卡的梯形 = Rotate Y 透视。agent 已 AE 实渲眼验（红近大、蓝远小、黄梯形）；
**待用户真机复核**后翻 `complete`。
