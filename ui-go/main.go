package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	appTitle = "L4D2 Autobhop Control"

	wmAppStatus = 0x8001
	wmTrayIcon  = 0x8002

	idProcessCombo = 101
	idRefreshBtn   = 102
	idStartBtn     = 103
	idStopBtn      = 104
	idTitleEdit    = 105
	idBaseCombo    = 106
	idFlagsEdit    = 107
	idPollEdit     = 108
	idStatusText   = 109

	statusIdle = iota
	statusScanning
	statusRunning
	statusStopped
	statusErrorOpenProcess
	statusErrorClientDLL
	statusGameClosed
	statusBadSettings

	PROCESS_VM_READ                   = 0x0010
	PROCESS_QUERY_LIMITED_INFO        = 0x1000
	TH32CS_SNAPMODULE          uint32 = 0x00000008
	TH32CS_SNAPMODULE32        uint32 = 0x00000010
	INVALID_HANDLE_VALUE              = ^uintptr(0)

	CB_ADDSTRING       = 0x0143
	CB_GETCURSEL       = 0x0147
	CB_SETCURSEL       = 0x014E
	CB_RESETCONTENT    = 0x014B
	CB_SHOWDROPDOWN    = 0x014F
	BN_CLICKED         = 0
	CBN_SELCHANGE      = 1
	WM_CLOSE           = 0x0010
	WM_COMMAND         = 0x0111
	WM_CREATE          = 0x0001
	WM_DESTROY         = 0x0002
	WM_CTLCOLORSTATIC  = 0x0138
	WM_CTLCOLOREDIT    = 0x0133
	WM_CTLCOLORLISTBOX = 0x0134
	WM_SETFONT         = 0x0030
	WM_GETTEXT         = 0x000D
	WM_GETTEXTLENGTH   = 0x000E
	WM_SETTEXT         = 0x000C
	WM_SETICON         = 0x0080
	WM_SYSCOMMAND      = 0x0112
	WM_LBUTTONUP       = 0x0202
	WM_RBUTTONUP       = 0x0205
	WM_LBUTTONDBLCLK   = 0x0203
	BM_SETCHECK        = 0x00F1
	BST_CHECKED        = 1
	SC_MINIMIZE        = 0xF020
	WS_OVERLAPPED      = 0x00000000
	WS_CAPTION         = 0x00C00000
	WS_SYSMENU         = 0x00080000
	WS_MINIMIZEBOX     = 0x00020000
	WS_VISIBLE         = 0x10000000
	WS_CHILD           = 0x40000000
	WS_TABSTOP         = 0x00010000
	WS_VSCROLL         = 0x00200000
	WS_BORDER          = 0x00800000
	CBS_DROPDOWNLIST   = 0x0003
	ES_AUTOHSCROLL     = 0x0080
	ES_READONLY        = 0x0800
	BS_PUSHBUTTON      = 0x00000000
	SS_LEFT            = 0x00000000
	IMAGE_ICON         = 1
	ICON_SMALL         = 0
	ICON_BIG           = 1
	LR_DEFAULTCOLOR    = 0x0000
	SW_SHOW            = 5
	SW_HIDE            = 0
	SW_RESTORE         = 9
	COLOR_WINDOW       = 5
	VK_SPACE           = 0x20
	WM_KEYDOWN         = 0x0100
	WM_KEYUP           = 0x0101
	NIM_ADD            = 0x00000000
	NIM_DELETE         = 0x00000002
	NIF_MESSAGE        = 0x00000001
	NIF_ICON           = 0x00000002
	NIF_TIP            = 0x00000004
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindow                 = user32.NewProc("IsWindow")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procLoadImageW               = user32.NewProc("LoadImageW")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procSendMessageW             = user32.NewProc("SendMessageW")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procSetBkMode                = gdi32.NewProc("SetBkMode")
	procSetTextColor             = gdi32.NewProc("SetTextColor")
	procSetWindowTextW           = user32.NewProc("SetWindowTextW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procTranslateMessage         = user32.NewProc("TranslateMessage")

	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procReadProcessMemory        = kernel32.NewProc("ReadProcessMemory")
	procModule32FirstW           = kernel32.NewProc("Module32FirstW")
	procModule32NextW            = kernel32.NewProc("Module32NextW")

	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreateFontW      = gdi32.NewProc("CreateFontW")

	procShellNotifyIconW = syscall.NewLazyDLL("shell32.dll").NewProc("Shell_NotifyIconW")
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
type moduleEntry32 struct {
	dwSize        uint32
	th32ModuleID  uint32
	th32ProcessID uint32
	glblcntUsage  uint32
	proccntUsage  uint32
	modBaseAddr   uintptr
	modBaseSize   uint32
	hModule       uintptr
	szModule      [256]uint16
	szExePath     [260]uint16
}
type guid struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}
type notifyIconData struct {
	cbSize           uint32
	hwnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         guid
	hBalloonIcon     uintptr
}
type targetWindow struct {
	hwnd  uintptr
	pid   uint32
	title string
}
type settings struct {
	hwnd             uintptr
	pid              uint32
	playerBaseOffset uintptr
	mFlagsOffset     uintptr
	pollInterval     time.Duration
}
type appState struct {
	hwnd        uintptr
	controls    map[int]uintptr
	targets     []targetWindow
	stop        chan struct{}
	running     bool
	mu          sync.Mutex
	font        uintptr
	bgBrush     uintptr
	fieldBrush  uintptr
	iconBig     uintptr
	iconSmall   uintptr
	trayVisible bool
}

var state = &appState{controls: make(map[int]uintptr)}

func main() {
	runtime.LockOSThread()

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("L4D2AutobhopUI")
	cursor, _, _ := procLoadCursorW.Call(0, uintptr(32512))
	state.iconBig = loadAppIcon(hInstance, 256)
	state.iconSmall = loadAppIcon(hInstance, 16)
	bg, _, _ := procCreateSolidBrush.Call(0x202020)
	field, _, _ := procCreateSolidBrush.Call(0x2C2C2C)
	state.bgBrush = bg
	state.fieldBrush = field

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInstance,
		hIcon:         state.iconBig,
		hCursor:       cursor,
		hbrBackground: bg,
		lpszClassName: className,
		hIconSm:       state.iconSmall,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(appTitle))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX|WS_VISIBLE,
		200, 120, 620, 430,
		0, 0, hInstance, 0,
	)
	state.hwnd = hwnd
	applyWindowIcons(hwnd)
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

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		createUI(hwnd)
		refreshTargets()
		setStatus(statusIdle)
		return 0
	case WM_COMMAND:
		id := int(wParam & 0xffff)
		code := int((wParam >> 16) & 0xffff)
		if id == idProcessCombo && code == CBN_SELCHANGE {
			syncSelectedTitle()
		} else if code == BN_CLICKED {
			switch id {
			case idRefreshBtn:
				refreshTargets()
				procSendMessageW.Call(state.controls[idProcessCombo], CB_SHOWDROPDOWN, 1, 0)
			case idStartBtn:
				startWorker()
			case idStopBtn:
				stopWorker(statusStopped)
			}
		}
		return 0
	case wmAppStatus:
		setStatus(int(wParam))
		return 0
	case wmTrayIcon:
		switch uint32(lParam) {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK, WM_RBUTTONUP:
			restoreFromTray()
		}
		return 0
	case WM_SYSCOMMAND:
		if wParam&0xfff0 == SC_MINIMIZE {
			minimizeToTray()
			return 0
		}
	case WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT, WM_CTLCOLORLISTBOX:
		hdc := wParam
		procSetBkMode.Call(hdc, 1)
		procSetTextColor.Call(hdc, 0xF2F2F2)
		if msg == WM_CTLCOLORSTATIC {
			return state.bgBrush
		}
		return state.fieldBrush
	case WM_CLOSE:
		deleteTrayIcon()
		stopWorker(statusStopped)
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		deleteTrayIcon()
		stopWorker(statusStopped)
		user32.NewProc("PostQuitMessage").Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func createUI(hwnd uintptr) {
	state.font, _, _ = procCreateFontW.Call(
		neg(16), 0, 0, 0, 500, 0, 0, 0, 1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))),
	)
	createText(hwnd, 22, 18, 430, 26, "L4D2 Autobhop Control", true)
	createText(hwnd, 22, 52, 530, 20, "选择游戏窗口，设置偏移，然后启动。后台运行，不阻塞界面。", false)

	createText(hwnd, 26, 92, 110, 22, "进程窗口", false)
	state.controls[idProcessCombo] = createControl("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST|WS_VSCROLL, 140, 88, 300, 260, hwnd, idProcessCombo)
	state.controls[idRefreshBtn] = createControl("BUTTON", "选择进程", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 455, 87, 120, 30, hwnd, idRefreshBtn)

	createText(hwnd, 26, 136, 110, 22, "窗口标题", false)
	state.controls[idTitleEdit] = createControl("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_AUTOHSCROLL|ES_READONLY, 140, 132, 435, 28, hwnd, idTitleEdit)

	createText(hwnd, 26, 180, 110, 22, "PlayerBase", false)
	state.controls[idBaseCombo] = createControl("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST, 140, 176, 160, 180, hwnd, idBaseCombo)
	for _, v := range []string{"0x726BD8", "0x73A574", "0x7C4424", "0x7C4450", "0x7C4644"} {
		sendString(state.controls[idBaseCombo], CB_ADDSTRING, v)
	}
	procSendMessageW.Call(state.controls[idBaseCombo], CB_SETCURSEL, 0, 0)

	createText(hwnd, 322, 180, 92, 22, "mFlags", false)
	state.controls[idFlagsEdit] = createControl("EDIT", "0xF0", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_AUTOHSCROLL, 410, 176, 165, 28, hwnd, idFlagsEdit)

	createText(hwnd, 26, 224, 110, 22, "轮询间隔 ms", false)
	state.controls[idPollEdit] = createControl("EDIT", "1", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_AUTOHSCROLL, 140, 220, 160, 28, hwnd, idPollEdit)

	state.controls[idStartBtn] = createControl("BUTTON", "启动", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 140, 278, 130, 36, hwnd, idStartBtn)
	state.controls[idStopBtn] = createControl("BUTTON", "停止", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 286, 278, 130, 36, hwnd, idStopBtn)
	state.controls[idStatusText] = createControl("STATIC", "", WS_CHILD|WS_VISIBLE|SS_LEFT, 26, 338, 548, 42, hwnd, idStatusText)
}

func loadAppIcon(hInstance uintptr, size int32) uintptr {
	icon, _, _ := procLoadImageW.Call(
		hInstance,
		1,
		IMAGE_ICON,
		uintptr(size),
		uintptr(size),
		LR_DEFAULTCOLOR,
	)
	return icon
}

func applyWindowIcons(hwnd uintptr) {
	if state.iconBig != 0 {
		procSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, state.iconBig)
	}
	if state.iconSmall != 0 {
		procSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, state.iconSmall)
	}
}

func minimizeToTray() {
	if state.hwnd == 0 {
		return
	}
	if addTrayIcon() {
		procShowWindow.Call(state.hwnd, SW_HIDE)
	}
}

func restoreFromTray() {
	if state.hwnd == 0 {
		return
	}
	deleteTrayIcon()
	procShowWindow.Call(state.hwnd, SW_RESTORE)
	procSetForegroundWindow.Call(state.hwnd)
}

func addTrayIcon() bool {
	if state.trayVisible || state.hwnd == 0 {
		return state.trayVisible
	}
	nid := newTrayIconData()
	ok, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	state.trayVisible = ok != 0
	return state.trayVisible
}

func deleteTrayIcon() bool {
	if !state.trayVisible || state.hwnd == 0 {
		return !state.trayVisible
	}
	nid := newTrayIconData()
	ok, _, _ := procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
	if ok != 0 {
		state.trayVisible = false
	}
	return !state.trayVisible
}

func newTrayIconData() notifyIconData {
	var nid notifyIconData
	nid.cbSize = uint32(unsafe.Sizeof(nid))
	nid.hwnd = state.hwnd
	nid.uID = 1
	nid.uFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.uCallbackMessage = wmTrayIcon
	nid.hIcon = state.iconSmall
	if nid.hIcon == 0 {
		nid.hIcon = state.iconBig
	}
	copyUTF16(nid.szTip[:], appTitle)
	return nid
}

func createText(parent uintptr, x, y, w, h int32, text string, title bool) uintptr {
	ctrl := createControl("STATIC", text, WS_CHILD|WS_VISIBLE|SS_LEFT, x, y, w, h, parent, 0)
	if title {
		font, _, _ := procCreateFontW.Call(neg(22), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))))
		procSendMessageW.Call(ctrl, WM_SETFONT, font, 1)
	}
	return ctrl
}

func createControl(class, text string, style uintptr, x, y, w, height int32, parent uintptr, id int) uintptr {
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(class))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		style,
		uintptr(x), uintptr(y), uintptr(w), uintptr(height),
		parent, uintptr(id), 0, 0,
	)
	if state.font != 0 {
		procSendMessageW.Call(hwnd, WM_SETFONT, state.font, 1)
	}
	return hwnd
}

func refreshTargets() {
	combo := state.controls[idProcessCombo]
	procSendMessageW.Call(combo, CB_RESETCONTENT, 0, 0)
	state.targets = enumerateGameWindows()
	for _, t := range state.targets {
		sendString(combo, CB_ADDSTRING, fmt.Sprintf("%s  [pid %d]", t.title, t.pid))
	}
	if len(state.targets) > 0 {
		procSendMessageW.Call(combo, CB_SETCURSEL, 0, 0)
		syncSelectedTitle()
		setStatus(statusIdle)
		return
	}
	sendString(combo, CB_ADDSTRING, "未找到可见的 Left 4 Dead 2 窗口")
	procSendMessageW.Call(combo, CB_SETCURSEL, 0, 0)
	setWindowText(state.controls[idTitleEdit], "")
	setStatus(statusScanning)
}

func syncSelectedTitle() {
	idx := int(send(state.controls[idProcessCombo], CB_GETCURSEL, 0, 0))
	if idx >= 0 && idx < len(state.targets) {
		setWindowText(state.controls[idTitleEdit], getWindowTitle(state.targets[idx].hwnd))
		return
	}
	setWindowText(state.controls[idTitleEdit], "")
}

func enumerateGameWindows() []targetWindow {
	var out []targetWindow
	cb := syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		if isWindowVisible(hwnd) && getWindowTextLength(hwnd) > 0 {
			title := getWindowTitle(hwnd)
			if strings.Contains(strings.ToLower(title), "left 4 dead 2") {
				var pid uint32
				procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
				out = append(out, targetWindow{hwnd: hwnd, pid: pid, title: title})
			}
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return out
}

func startWorker() {
	state.mu.Lock()
	if state.running {
		state.mu.Unlock()
		return
	}
	cfg, ok := readSettings()
	if !ok {
		state.mu.Unlock()
		setStatus(statusBadSettings)
		return
	}
	state.stop = make(chan struct{})
	stop := state.stop
	state.running = true
	state.mu.Unlock()

	postStatus(statusScanning)
	go runBhop(cfg, stop)
}

func stopWorker(status int) {
	state.mu.Lock()
	if state.running && state.stop != nil {
		close(state.stop)
	}
	state.running = false
	state.stop = nil
	state.mu.Unlock()
	setStatus(status)
}

func readSettings() (settings, bool) {
	var cfg settings
	if idx := int(send(state.controls[idProcessCombo], CB_GETCURSEL, 0, 0)); idx >= 0 && idx < len(state.targets) {
		cfg.hwnd = state.targets[idx].hwnd
		cfg.pid = state.targets[idx].pid
		setWindowText(state.controls[idTitleEdit], getWindowTitle(cfg.hwnd))
	}
	if cfg.hwnd == 0 || cfg.pid == 0 {
		return cfg, false
	}
	cfg.playerBaseOffset = parseHex(comboText(state.controls[idBaseCombo]))
	cfg.mFlagsOffset = parseHex(getControlText(state.controls[idFlagsEdit]))
	ms, err := strconv.Atoi(strings.TrimSpace(getControlText(state.controls[idPollEdit])))
	if err != nil || ms < 1 || cfg.playerBaseOffset == 0 || cfg.mFlagsOffset == 0 {
		return cfg, false
	}
	if ms > 50 {
		ms = 50
	}
	cfg.pollInterval = time.Duration(ms) * time.Millisecond
	return cfg, true
}

func runBhop(cfg settings, stop <-chan struct{}) {
	hProcess, _, _ := procOpenProcess.Call(PROCESS_VM_READ|PROCESS_QUERY_LIMITED_INFO, 0, uintptr(cfg.pid))
	if hProcess == 0 {
		postStatus(statusErrorOpenProcess)
		markStopped()
		return
	}
	defer procCloseHandle.Call(hProcess)

	clientBase, stopped := waitForClientBase(cfg.pid, stop)
	if stopped {
		postStatus(statusStopped)
		markStopped()
		return
	}
	if clientBase == 0 {
		postStatus(statusErrorClientDLL)
		markStopped()
		return
	}
	postStatus(statusRunning)

	ticker := time.NewTicker(cfg.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			postKey(cfg.hwnd, WM_KEYUP)
			return
		case <-ticker.C:
			if !isWindow(cfg.hwnd) {
				postStatus(statusGameClosed)
				markStopped()
				return
			}
			key, _, _ := procGetAsyncKeyState.Call(VK_SPACE)
			if key&0x8000 == 0 {
				continue
			}
			playerPtr, ok := readUint32(hProcess, clientBase+cfg.playerBaseOffset)
			if !ok || playerPtr == 0 {
				continue
			}
			flags, ok := readUint32(hProcess, uintptr(playerPtr)+cfg.mFlagsOffset)
			if !ok {
				continue
			}
			if onGround(flags) {
				postKey(cfg.hwnd, WM_KEYUP)
			} else {
				postKey(cfg.hwnd, WM_KEYDOWN)
			}
		}
	}
}

func waitForClientBase(pid uint32, stop <-chan struct{}) (uintptr, bool) {
	for {
		select {
		case <-stop:
			return 0, true
		default:
		}
		base := clientDLLBase(pid)
		if base != 0 {
			return base, false
		}
		time.Sleep(750 * time.Millisecond)
	}
}

func clientDLLBase(pid uint32) uintptr {
	snap, _, _ := procCreateToolhelp32Snapshot.Call(uintptr(TH32CS_SNAPMODULE|TH32CS_SNAPMODULE32), uintptr(pid))
	if snap == INVALID_HANDLE_VALUE || snap == 0 {
		return 0
	}
	defer procCloseHandle.Call(snap)

	var me moduleEntry32
	me.dwSize = uint32(unsafe.Sizeof(me))
	ok, _, _ := procModule32FirstW.Call(snap, uintptr(unsafe.Pointer(&me)))
	for ok != 0 {
		if strings.EqualFold(syscall.UTF16ToString(me.szModule[:]), "client.dll") {
			return me.modBaseAddr
		}
		ok, _, _ = procModule32NextW.Call(snap, uintptr(unsafe.Pointer(&me)))
	}
	return 0
}

func readUint32(process uintptr, addr uintptr) (uint32, bool) {
	var value uint32
	var read uintptr
	ok, _, _ := procReadProcessMemory.Call(process, addr, uintptr(unsafe.Pointer(&value)), unsafe.Sizeof(value), uintptr(unsafe.Pointer(&read)))
	return value, ok != 0 && read == unsafe.Sizeof(value)
}

func onGround(flags uint32) bool {
	return flags == 0x80 || flags == 0x82 || flags == 0x280 || flags == 0x282
}

func postKey(hwnd uintptr, message uint32) {
	procPostMessageW.Call(hwnd, uintptr(message), VK_SPACE, 0x390000)
}

func markStopped() {
	state.mu.Lock()
	state.running = false
	state.stop = nil
	state.mu.Unlock()
}

func setStatus(code int) {
	text := map[int]string{
		statusIdle:             "就绪：选择进程后启动。",
		statusScanning:         "等待游戏窗口或 client.dll 加载中...",
		statusRunning:          "运行中：按住空格时自动连跳。",
		statusStopped:          "已停止。",
		statusErrorOpenProcess: "无法打开游戏进程，请尝试以相同权限运行本工具。",
		statusErrorClientDLL:   "未能读取 client.dll，可能游戏已关闭或尚未加载地图。",
		statusGameClosed:       "游戏窗口已关闭，已停止。",
		statusBadSettings:      "设置无效：请确认进程、偏移和轮询间隔。",
	}[code]
	setWindowText(state.controls[idStatusText], text)
}

func postStatus(code int) {
	procPostMessageW.Call(state.hwnd, wmAppStatus, uintptr(code), 0)
}

func send(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func sendString(hwnd uintptr, msg uint32, text string) {
	procSendMessageW.Call(hwnd, uintptr(msg), 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getControlText(hwnd uintptr) string {
	length := send(hwnd, WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, length+1)
	send(hwnd, WM_GETTEXT, uintptr(len(buf)), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func comboText(hwnd uintptr) string {
	idx := int(send(hwnd, CB_GETCURSEL, 0, 0))
	if idx < 0 {
		return ""
	}
	return getControlText(hwnd)
}

func setWindowText(hwnd uintptr, text string) {
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getWindowTextLength(hwnd uintptr) int {
	ret, _, _ := procGetWindowTextLengthW.Call(hwnd)
	return int(ret)
}

func getWindowTitle(hwnd uintptr) string {
	n := getWindowTextLength(hwnd)
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func isWindow(hwnd uintptr) bool {
	ret, _, _ := procIsWindow.Call(hwnd)
	return ret != 0
}

func isWindowVisible(hwnd uintptr) bool {
	ret, _, _ := procIsWindowVisible.Call(hwnd)
	return ret != 0
}

func parseHex(s string) uintptr {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0
	}
	return uintptr(v)
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func copyUTF16(dst []uint16, s string) {
	src := syscall.StringToUTF16(s)
	if len(src) > len(dst) {
		src = src[:len(dst)]
		src[len(src)-1] = 0
	}
	copy(dst, src)
}

func neg(v uintptr) uintptr {
	return ^(v - 1)
}
