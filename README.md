# Left4Dead2-Autobhop

External auto-bhop helper for Left 4 Dead 2. When the tool is running, hold the
spacebar in game to perform continuous jump input.

This fork adds a native Go + Win32 desktop UI alongside the original C console
implementation.

## Warning

Use at your own risk. Do not use this tool on VAC-protected servers. External
memory tools and input helpers may still violate server or platform rules and
can result in bans.

## What's New In This Fork

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

### Build

Requirements:

- Windows
- Go 1.26 or newer

Build the UI:

```powershell
cd ui-go
go build -ldflags "-H=windowsgui" -o l4d2-autobhop-ui.exe .
```

`app.syso` is committed so the icon resource is linked automatically by Go.

### Run

1. Launch Left 4 Dead 2.
2. Run `ui-go/l4d2-autobhop-ui.exe`.
3. Click the process selection button and choose the game window.
4. Keep the default offsets, or choose another `PlayerBase` preset if needed.
5. Click Start.
6. Hold spacebar in game.

Minimizing the UI hides it to the system tray. Click the tray icon to restore
the window.

## Original Console Version

The original C implementation is still available in the repository:

- `main.c`
- `utils.c`
- `utils.h`
- `l4d2-autobhop.sln`
- `l4d2-autobhop.vcxproj`

Build it with Visual Studio 2022 using the x64 Release configuration.

## Known Offsets

Default values:

```text
player base offset = 0x726BD8
mFlags offset      = 0xF0
```

Currently listed `PlayerBase` presets:

```text
0x726BD8
0x73A574
0x7C4424
0x7C4450
0x7C4644
```

## Credits

Original project lineage:

- https://github.com/mads256h/l4d2-autobhop
- https://github.com/KnifeLemon/Left4Dead2-Autobhop-CSharp

This fork keeps the original external-memory approach and adds a safer,
responsive Win32 UI wrapper.
