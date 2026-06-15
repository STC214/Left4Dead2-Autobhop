# Left4Dead2-Autobhop / 求生之路 2 自动连跳外置工具

这是一个用于《Left 4 Dead 2 / 求生之路 2》的外置自动连跳辅助工具。运行工具后，在游戏中按住空格键即可持续发送跳跃输入。

External auto-bhop helper for Left 4 Dead 2. When the tool is running, hold the
spacebar in game to perform continuous jump input.

本 fork 在保留原 C 控制台版本的基础上，新增了一个原生 Go + Win32 桌面 UI。

This fork adds a native Go + Win32 desktop UI alongside the original C console
implementation.

## 致谢与引用 / Attribution

本仓库 fork 自：

This repository is a fork of:

- 原始上游仓库 / Original upstream repository: https://github.com/Efah-Dev/Left4Dead2-Autobhop

原项目提供了外置自动连跳逻辑、读内存方式、已知偏移以及 Visual Studio C 实现。本 fork 保留原实现，并在 `ui-go/` 中新增独立的 Go + Win32 UI 包装层。

The original project provides the external auto-bhop logic, memory-reading
approach, known offsets, and Visual Studio C implementation. This fork keeps
that original implementation intact and adds a separate Go + Win32 UI wrapper in
`ui-go/`.

原 README 也引用并感谢了以下相关项目：

The original README also credits these related projects:

- https://github.com/mads256h/l4d2-autobhop
- https://github.com/KnifeLemon/Left4Dead2-Autobhop-CSharp

请参考上游仓库和相关项目了解原始代码来源与背景。

Please refer to the upstream repository and the related projects for the
original code lineage and background.

## 警告 / Warning

请自行承担使用风险。不要在 VAC 保护服务器上使用此工具。外置读内存工具和输入辅助仍可能违反服务器或平台规则，并可能导致封禁。

Use at your own risk. Do not use this tool on VAC-protected servers. External
memory tools and input helpers may still violate server or platform rules and
can result in bans.

## 本 fork 新增内容 / What's New In This Fork

- `ui-go/` 中新增原生 Go + Win32 UI。
- 使用进程/窗口选择器，不再依赖固定窗口标题。
- 可编辑运行参数：
  - 原项目列出的 `PlayerBase` 偏移预设。
  - `mFlags` 偏移。
  - 轮询间隔。
- 等待游戏、等待模块、读内存、发送按键都在后台 worker 中执行，避免 UI 线程卡死。
- 程序图标内置到 exe 资源中。
- 图标可显示在 exe 文件、标题栏、Windows 任务栏和系统托盘中。
- 支持最小化到系统托盘。

- Native Go + Win32 UI in `ui-go/`.
- Process/window selector instead of relying on a fixed window title.
- Editable runtime settings:
  - `PlayerBase` offset presets from the original project.
  - `mFlags` offset.
  - polling interval.
- Background worker for game/module waiting, memory reads, and key posting, so
  the UI thread stays responsive.
- Program icon embedded into the exe resource.
- Icon shown in the exe file, title bar, Windows taskbar, and system tray.
- Minimize-to-tray support.

## Go Win32 UI

### 构建 / Build

要求：

Requirements:

- Windows
- Go 1.26 或更新版本 / Go 1.26 or newer

构建 UI：

Build the UI:

```powershell
cd ui-go
go build -ldflags "-H=windowsgui" -o l4d2-autobhop-ui.exe .
```

`app.syso` 已提交到仓库中，因此 Go 构建时会自动链接图标资源。

`app.syso` is committed so the icon resource is linked automatically by Go.

### 运行 / Run

1. 启动《Left 4 Dead 2 / 求生之路 2》。
2. 运行 `ui-go/l4d2-autobhop-ui.exe`。
3. 点击进程选择按钮，选择游戏窗口。
4. 保持默认偏移，或在需要时选择其他 `PlayerBase` 预设。
5. 点击 Start / 启动。
6. 在游戏中按住空格键。

1. Launch Left 4 Dead 2.
2. Run `ui-go/l4d2-autobhop-ui.exe`.
3. Click the process selection button and choose the game window.
4. Keep the default offsets, or choose another `PlayerBase` preset if needed.
5. Click Start.
6. Hold spacebar in game.

最小化 UI 后，窗口会隐藏到系统托盘。点击托盘图标可以恢复窗口。

Minimizing the UI hides it to the system tray. Click the tray icon to restore
the window.

## 原控制台版本 / Original Console Version

仓库中仍保留原始 C 实现：

The original C implementation is still available in the repository:

- `main.c`
- `utils.c`
- `utils.h`
- `l4d2-autobhop.sln`
- `l4d2-autobhop.vcxproj`

可以使用 Visual Studio 2022 的 x64 Release 配置进行构建。

Build it with Visual Studio 2022 using the x64 Release configuration.

## 已知偏移 / Known Offsets

默认值：

Default values:

```text
player base offset = 0x726BD8
mFlags offset      = 0xF0
```

目前列出的 `PlayerBase` 预设：

Currently listed `PlayerBase` presets:

```text
0x726BD8
0x73A574
0x7C4424
0x7C4450
0x7C4644
```

## 本 fork 的改动 / Fork Changes

本 fork 保留原项目的外置读内存实现方式，并新增更安全、响应更好的 Win32 UI 包装层。UI 代码、托盘支持、图标资源接入和更新后的文档由本 fork 维护。

This fork keeps the original external-memory approach and adds a safer,
responsive Win32 UI wrapper. The UI code, tray support, icon resource wiring,
and updated documentation are maintained in this fork.
