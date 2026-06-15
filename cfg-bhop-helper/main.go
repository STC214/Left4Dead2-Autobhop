package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	appTitle = "L4D2 CFG Bhop Helper"

	idPathEdit   = 101
	idDetectBtn  = 102
	idInstallBtn = 103
	idCleanBtn   = 104
	idOpenBtn    = 105
	idStatus     = 106

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_COMMAND        = 0x0111
	WM_CTLCOLORSTATIC = 0x0138
	WM_CTLCOLOREDIT   = 0x0133
	WM_SETFONT        = 0x0030
	WM_GETTEXT        = 0x000D
	WM_GETTEXTLENGTH  = 0x000E
	BN_CLICKED        = 0
	WS_OVERLAPPED     = 0x00000000
	WS_CAPTION        = 0x00C00000
	WS_SYSMENU        = 0x00080000
	WS_MINIMIZEBOX    = 0x00020000
	WS_VISIBLE        = 0x10000000
	WS_CHILD          = 0x40000000
	WS_TABSTOP        = 0x00010000
	WS_BORDER         = 0x00800000
	ES_AUTOHSCROLL    = 0x0080
	BS_PUSHBUTTON     = 0x00000000
	SS_LEFT           = 0x00000000
	SW_SHOW           = 5
)

const cfgBlock = `// >>> L4D2_AUTO_BHOP_CFG_HELPER_BEGIN
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
`

const (
	blockBegin = "// >>> L4D2_AUTO_BHOP_CFG_HELPER_BEGIN"
	blockEnd   = "// <<< L4D2_AUTO_BHOP_CFG_HELPER_END"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procLoadCursorW      = user32.NewProc("LoadCursorW")
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procSendMessageW     = user32.NewProc("SendMessageW")
	procSetWindowTextW   = user32.NewProc("SetWindowTextW")
	procShowWindow       = user32.NewProc("ShowWindow")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreateFontW      = gdi32.NewProc("CreateFontW")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procGetLogicalDrives = kernel32.NewProc("GetLogicalDrives")
	procShellExecuteW    = shell32.NewProc("ShellExecuteW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
)

type point struct{ x, y int32 }
type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}
type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

var (
	controls = map[int]uintptr{}
	font     uintptr
	bgBrush  uintptr
)

func main() {
	runtime.LockOSThread()

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("L4D2CfgBhopHelper")
	cursor, _, _ := procLoadCursorW.Call(0, uintptr(32512))
	bgBrush, _, _ = procCreateSolidBrush.Call(0x202020)

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInstance,
		hCursor:       cursor,
		hbrBackground: bgBrush,
		lpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(appTitle))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX|WS_VISIBLE,
		260, 160, 720, 380,
		0, 0, hInstance, 0,
	)
	procShowWindow.Call(hwnd, SW_SHOW)

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CREATE:
		createUI(hwnd)
		if cfg := firstDetectedCfgDir(); cfg != "" {
			setWindowText(controls[idPathEdit], cfg)
			setStatus("已自动找到 cfg 目录。建议 Steam 启动项添加：+exec autoexec.cfg")
		} else {
			setStatus("未自动找到 cfg 目录。请手动填入 left4dead2\\cfg 路径。")
		}
		return 0
	case WM_COMMAND:
		id := int(wParam & 0xffff)
		code := int((wParam >> 16) & 0xffff)
		if code == BN_CLICKED {
			switch id {
			case idDetectBtn:
				detectPath()
			case idInstallBtn:
				installCFG()
			case idCleanBtn:
				cleanCFG()
			case idOpenBtn:
				openCfgDir()
			}
		}
		return 0
	case WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT:
		hdc := wParam
		procSetBkMode.Call(hdc, 1)
		procSetTextColor.Call(hdc, 0xF2F2F2)
		return bgBrush
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func createUI(hwnd uintptr) {
	font, _, _ = procCreateFontW.Call(neg(16), 0, 0, 0, 500, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))))

	createText(hwnd, 24, 20, 660, 22, "L4D2 CFG 连跳助手：写入 autoexec.cfg，清理时只删除本工具写入的标记块。")
	createText(hwnd, 24, 54, 660, 22, "提示：不推荐在 VAC 服务器中使用。CFG wait 连跳依赖服务器允许 wait 指令。")
	createText(hwnd, 24, 92, 80, 22, "cfg 路径")
	controls[idPathEdit] = createControl("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_AUTOHSCROLL, 104, 88, 470, 28, hwnd, idPathEdit)
	controls[idDetectBtn] = createControl("BUTTON", "自动查找", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 588, 88, 96, 30, hwnd, idDetectBtn)
	controls[idInstallBtn] = createControl("BUTTON", "一键写入", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 104, 142, 120, 36, hwnd, idInstallBtn)
	controls[idCleanBtn] = createControl("BUTTON", "一键清理", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 244, 142, 120, 36, hwnd, idCleanBtn)
	controls[idOpenBtn] = createControl("BUTTON", "打开目录", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 384, 142, 120, 36, hwnd, idOpenBtn)
	controls[idStatus] = createControl("STATIC", "", WS_CHILD|WS_VISIBLE|SS_LEFT, 24, 210, 660, 90, hwnd, idStatus)
}

func createText(parent uintptr, x, y, w, h int32, text string) uintptr {
	return createControl("STATIC", text, WS_CHILD|WS_VISIBLE|SS_LEFT, x, y, w, h, parent, 0)
}

func createControl(class, text string, style uintptr, x, y, w, h int32, parent uintptr, id int) uintptr {
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(class))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		style,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		parent, uintptr(id), 0, 0,
	)
	if font != 0 {
		procSendMessageW.Call(hwnd, WM_SETFONT, font, 1)
	}
	return hwnd
}

func detectPath() {
	if cfg := firstDetectedCfgDir(); cfg != "" {
		setWindowText(controls[idPathEdit], cfg)
		setStatus("已找到 cfg 目录。点击“一键写入”会备份并更新 autoexec.cfg。")
		return
	}
	setStatus("未找到 cfg 目录。请手动填入 SteamLibrary\\steamapps\\common\\Left 4 Dead 2\\left4dead2\\cfg。")
}

func installCFG() {
	cfgDir := strings.TrimSpace(getControlText(controls[idPathEdit]))
	if !isCfgDir(cfgDir) {
		setStatus("路径无效：请选择 Left 4 Dead 2\\left4dead2\\cfg 目录。")
		return
	}
	autoexec := filepath.Join(cfgDir, "autoexec.cfg")
	original, _ := os.ReadFile(autoexec)
	if len(original) > 0 {
		backup := filepath.Join(cfgDir, fmt.Sprintf("autoexec.cfg.bak.%s", time.Now().Format("20060102-150405")))
		if err := os.WriteFile(backup, original, 0644); err != nil {
			setStatus("备份失败：" + err.Error())
			return
		}
	}
	next := removeBlock(string(original))
	if strings.TrimSpace(next) != "" && !strings.HasSuffix(next, "\n") {
		next += "\r\n"
	}
	next += "\r\n" + cfgBlock
	if err := os.WriteFile(autoexec, []byte(next), 0644); err != nil {
		setStatus("写入失败：" + err.Error())
		return
	}
	setStatus("写入完成：autoexec.cfg 已更新。Steam 启动项建议添加：+exec autoexec.cfg")
}

func cleanCFG() {
	cfgDir := strings.TrimSpace(getControlText(controls[idPathEdit]))
	if !isCfgDir(cfgDir) {
		setStatus("路径无效：请选择 Left 4 Dead 2\\left4dead2\\cfg 目录。")
		return
	}
	autoexec := filepath.Join(cfgDir, "autoexec.cfg")
	original, err := os.ReadFile(autoexec)
	if err != nil {
		setStatus("读取失败：" + err.Error())
		return
	}
	next := removeBlock(string(original))
	if next == string(original) {
		setStatus("未发现本工具写入的 CFG 标记块，无需清理。")
		return
	}
	backup := filepath.Join(cfgDir, fmt.Sprintf("autoexec.cfg.cleanbak.%s", time.Now().Format("20060102-150405")))
	if err := os.WriteFile(backup, original, 0644); err != nil {
		setStatus("清理前备份失败：" + err.Error())
		return
	}
	if err := os.WriteFile(autoexec, []byte(strings.TrimSpace(next)+"\r\n"), 0644); err != nil {
		setStatus("清理失败：" + err.Error())
		return
	}
	setStatus("清理完成：已删除本工具写入的 CFG 块，并保留清理前备份。")
}

func openCfgDir() {
	cfgDir := strings.TrimSpace(getControlText(controls[idPathEdit]))
	if !isCfgDir(cfgDir) {
		setStatus("路径无效，无法打开。")
		return
	}
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("open"))), uintptr(unsafe.Pointer(utf16Ptr(cfgDir))), 0, 0, SW_SHOW)
}

func removeBlock(s string) string {
	start := strings.Index(s, blockBegin)
	end := strings.Index(s, blockEnd)
	if start < 0 || end < 0 || end < start {
		return s
	}
	end += len(blockEnd)
	for end < len(s) && (s[end] == '\r' || s[end] == '\n') {
		end++
	}
	return strings.TrimRight(s[:start], "\r\n") + "\r\n" + s[end:]
}

func firstDetectedCfgDir() string {
	for _, dir := range candidateCfgDirs() {
		if isCfgDir(dir) {
			return dir
		}
	}
	return ""
}

func candidateCfgDirs() []string {
	var out []string
	for _, base := range []string{
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("ProgramFiles"),
	} {
		if base != "" {
			out = append(out, filepath.Join(base, "Steam", "steamapps", "common", "Left 4 Dead 2", "left4dead2", "cfg"))
		}
	}
	for _, drive := range logicalDrives() {
		out = append(out,
			filepath.Join(drive, "SteamLibrary", "steamapps", "common", "Left 4 Dead 2", "left4dead2", "cfg"),
			filepath.Join(drive, "Steam", "steamapps", "common", "Left 4 Dead 2", "left4dead2", "cfg"),
		)
	}
	return dedupe(out)
}

func logicalDrives() []string {
	mask, _, _ := procGetLogicalDrives.Call()
	var drives []string
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) != 0 {
			drives = append(drives, fmt.Sprintf("%c:\\", 'A'+i))
		}
	}
	return drives
}

func isCfgDir(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	normalized := strings.ToLower(filepath.Clean(path))
	return strings.HasSuffix(normalized, filepath.Join("left4dead2", "cfg"))
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if v == "" {
			continue
		}
		key := strings.ToLower(filepath.Clean(v))
		if !seen[key] {
			seen[key] = true
			out = append(out, v)
		}
	}
	return out
}

func setStatus(text string) {
	setWindowText(controls[idStatus], text+"\r\nSteam 启动项建议：+exec autoexec.cfg")
}

func setWindowText(hwnd uintptr, text string) {
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getControlText(hwnd uintptr) string {
	length := send(hwnd, WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, length+1)
	send(hwnd, WM_GETTEXT, uintptr(len(buf)), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func send(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func neg(v uintptr) uintptr {
	return ^(v - 1)
}
