---
showcase: camera-light
direction: 相机/灯光层 from-scratch 创建 — NewCameraLayer / NewLightLayer，纯 Go 从零建层后 AE DOM 确认层类型（📋 读值档，不看图）
capabilities: [new-camera-layer, new-light-layer, source-less-layer]
gates: [TestNewCameraLayer_AEShipGate, TestNewLightLayer_AEShipGate, camera-light-layer-create-re]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/camera-light  +  scripts/ae_run.ps1 verify.jsx"
---

# camera-light — 相机/灯光层创建 showcase（📋 读值档）

## 这个方向测什么

`NewCameraLayer` / `NewLightLayer` 从零建层（embed 整块 AE-native Layr 模板）。相机/灯光只作用于 3D 几何（本档无），无渲染视觉 → 读值档：AE 打开后 DOM 确认**两层被接受 + 类型正确**。

## 验证方式（📋 读值，非看图）

`verify.jsx` 打开 .aep,确认 `Cam01 instanceof CameraLayer` / `Light01 instanceof LightLayer` + dump 层类型/默认选项到 `.done`。无 png。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `camera_light.aep`（BG + Cam01 + Light01） |
| `verify.jsx` | 读值脚本(tracked) | dump 层类型 + 选项 DOM → `.done`（无 png） |
| `camera_light.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 |
| `camera_light.done` | readback 日志 (gitignored) | 用户读这个核值 |

## 期望 readback（AE2020 实测一致）

| 字段 | 期望值 | 说明 |
|---|---|---|
| comp layers | 3 | BG + Cam01 + Light01 |
| Cam01 instanceof CameraLayer | true | 相机层创建被 AE 认（设置面板=「双节点摄像机」） |
| Cam01 模板默认 | zoom=1000 · DOF=on · focus=1500 · aperture=50 · blur=75 | 选项 setter 从零 elide → 全是模板默认值 |
| Light01 instanceof LightLayer | true | 灯光层创建被 AE 认 |
| Light01 lightType | 4415 = **AMBIENT（环境光）** | 模板默认 = **环境光**（非点光！） |
| Light01 intensity | 40 | 模板默认 |

> 审核要点：两个 `instanceof` 全 true + layers=3 即通过(create 路径可用)。**注意：灯光默认是环境光,相机默认值是模板内置（非本库设的）。**

## ⚠ 真实边界（诚实标注）

**选项 setter 从零不可用**——`SetCameraZoom/FocusDistance/Aperture/BlurLevel/DepthOfField` 与 `SetLightIntensity/Color/ConeAngle/ConeFeather` 在 from-scratch 模板层上全部报 **"property not present"**（选项属性流在 embed 模板里被 elide,同 effect param elision）。故本档只能演示**建层 + 类型**,选项值是模板默认(zoom=1000 / intensity=40 / type=POINT)。这些 setter 在 **parsed 层**(属性已存在)上才生效——属 fixture-mutation 能力,非 from-scratch。改光型(无 SetLightType)同理需 parsed 层或扩模板。

## 溯源

source-less 层创建（embed-whole-Layr）：`incidents/camera-light-layer-create-re.md`；FilmSize 写阻塞负发现 `camera-filmsize-ldta-write-blocked.md`；ship-gate `layer_light_test.go` 等。
