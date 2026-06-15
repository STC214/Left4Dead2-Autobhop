# L4D2 CFG 连跳助手 / L4D2 CFG Bhop Helper

> 不推荐在开启 VAC 的服务器中使用。封号、服务器踢出或其他后果自负。
>
> Do not use this on VAC-enabled servers. You are responsible for any ban, kick,
> or other consequence.

## wait 指令的生与死

这套 CFG 连跳脚本依赖 `wait` 指令。

- 本地单人/本地房主：通常可用。
- 官方匹配服务器：通常不可用，因为官方服务器常见配置会禁用 `wait`。
- 第三方社区服务器：看服务器配置。有些服务器允许 `wait`，有些服务器自带连跳插件，可能根本不需要脚本。

如果服务器禁用了 `wait`，脚本可能只跳一下，或者循环直接断掉。这不是安装失败，而是服务器规则导致。

## 按住空格连跳 CFG

助手写入 `autoexec.cfg` 的内容如下。脚本使用标记块包住，方便一键清理。

```text
// >>> L4D2_AUTO_BHOP_CFG_HELPER_BEGIN
// L4D2 auto bhop cfg helper block.
// Requires server-side wait command support.

alias bhop_bind "+jump; wait 2; -jump; wait 2; bhop_jump"
alias bhop_comm "alias bhop_jump bhop_bind"
alias bhop_stop "alias bhop_jump; -jump"
alias +bhop "bhop_comm; bhop_jump"
alias -bhop "bhop_stop"

unbind "SPACE"
unbind "SHIFT"
bind "SHIFT" "+speed"
bind "SPACE" "+bhop"

-jump

echo "-----------------------------------------------------"
echo ">> L4D2 CFG bhop helper loaded: hold SPACE to bhop."
echo ">> If it jumps only once, this server may block wait."
echo "-----------------------------------------------------"
// <<< L4D2_AUTO_BHOP_CFG_HELPER_END
```

为了减少编码问题，写入的 CFG 块只使用 ASCII 字符。中文说明留在本文档中。

## 手动安装

1. 找到游戏配置目录：

```text
SteamLibrary\steamapps\common\Left 4 Dead 2\left4dead2\cfg
```

2. 打开或新建：

```text
autoexec.cfg
```

3. 复制上面的 CFG 块到文件末尾。

4. 建议在 Steam 启动选项中加入：

```text
+exec autoexec.cfg
```

## 一键工具

本目录提供一个 Go + Win32 小工具：

```text
cfg-bhop-helper.exe
```

功能：

- 自动扫描常见 Steam / SteamLibrary 安装路径。
- 一键写入 CFG 块。
- 写入前备份现有 `autoexec.cfg`。
- 一键清理 CFG 块。
- 打开游戏 `cfg` 目录。
- 显示需要添加到 Steam 启动项的命令。

它不会静默修改 Steam 的用户配置文件。Steam 启动项请手动添加，避免误改 Steam 配置。

## 构建便携版

```powershell
.\package-portable.ps1
```

输出：

```text
dist/cfg-bhop-helper.exe
```
