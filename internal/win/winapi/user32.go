package winapi

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	getClassName                 = user32.NewProc("GetClassNameW")
	enumChildWindows             = user32.NewProc("EnumChildWindows")
	enumWindows                  = user32.NewProc("EnumWindows")
	showWindow                   = user32.NewProc("ShowWindow")
	getParent                    = user32.NewProc("GetParent")
	setWindowPos                 = user32.NewProc("SetWindowPos")
	getWindowLongPtr             = user32.NewProc("GetWindowLongPtrW")
	setWindowLongPtr             = user32.NewProc("SetWindowLongPtrW")
	getWindowText                = user32.NewProc("GetWindowTextW")
	getWindowRect                = user32.NewProc("GetWindowRect")
	getDpiForWindow              = user32.NewProc("GetDpiForWindow")
	isWindow                     = user32.NewProc("IsWindow")
	isWindowVisible              = user32.NewProc("IsWindowVisible")
	isZoomed                     = user32.NewProc("IsZoomed")
	updateWindow                 = user32.NewProc("UpdateWindow")
	sendMessage                  = user32.NewProc("SendMessageW")
	getWindowThreadProcessId     = user32.NewProc("GetWindowThreadProcessId")
	moveWindow                   = user32.NewProc("MoveWindow")
	setWindowRgn                 = user32.NewProc("SetWindowRgn")
	redrawWindow                 = user32.NewProc("RedrawWindow")
	setThreadDpiAwarenessContext = user32.NewProc("SetThreadDpiAwarenessContext")
)

var windowStyleIndexArg = uintptr(^uint32(15))

func GetClassName(hWnd windows.HWND) string {
	buff := make([]uint16, 255)
	_, _, _ = getClassName.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&buff[0])), 255)
	return syscall.UTF16ToString(buff)
}

func EnumWindows(lpEnumFunc, lParam uintptr) bool {
	r, _, _ := enumWindows.Call(lpEnumFunc, lParam)
	return r != 0
}
func EnumChildWindows(hWndParent windows.HWND, lpEnumFunc, lParam uintptr) bool {
	r, _, _ := enumChildWindows.Call(uintptr(hWndParent), lpEnumFunc, lParam)
	return r != 0
}

func ShowWindow(hWnd windows.HWND, nCmdShow int32) bool {
	r, _, _ := showWindow.Call(uintptr(hWnd), uintptr(nCmdShow))
	return r != 0
}

func GetParent(hWnd windows.HWND) windows.HWND {
	r, _, _ := getParent.Call(uintptr(hWnd))
	return windows.HWND(r)
}

func SetWindowPos(hWnd, hWndInsertAfter windows.HWND, x, y, cx, cy int32, uFlags uint32) bool {
	r, _, _ := setWindowPos.Call(uintptr(hWnd), uintptr(hWndInsertAfter), uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(uFlags))
	return r != 0
}

func GetWindowText(hWnd windows.HWND) string {
	buff := make([]uint16, 255)
	_, _, _ = getWindowText.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&buff[0])), 255)
	return syscall.UTF16ToString(buff)
}

func GetWindowRect(hWnd windows.HWND, lpRect *Rect) bool {
	r, _, _ := getWindowRect.Call(uintptr(hWnd), uintptr(unsafe.Pointer(lpRect)))
	return r != 0
}

func GetWindowStyle(hWnd windows.HWND) uintptr {
	style, _, _ := getWindowLongPtr.Call(uintptr(hWnd), windowStyleIndexArg)
	return style
}

func SetWindowStyle(hWnd windows.HWND, style uintptr) bool {
	setWindowLongPtr.Call(uintptr(hWnd), windowStyleIndexArg, style)
	return GetWindowStyle(hWnd) == style
}

func IsWindow(hWnd windows.HWND) bool {
	result, _, _ := isWindow.Call(uintptr(hWnd))
	return result != 0
}

func IsWindowVisible(hWnd windows.HWND) bool {
	result, _, _ := isWindowVisible.Call(uintptr(hWnd))
	return result != 0
}

func IsZoomed(hWnd windows.HWND) bool {
	result, _, _ := isZoomed.Call(uintptr(hWnd))
	return result != 0
}

func UpdateWindow(hWnd windows.HWND) bool {
	ret, _, _ := updateWindow.Call(uintptr(hWnd))
	return ret != 0
}

func SendMessage(hWd windows.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := sendMessage.Call(uintptr(hWd), uintptr(msg), wParam, lParam)
	return r
}

func GetWindowThreadProcessId(hWnd windows.HWND, dwProcessId *uint32) uint32 {
	r, _, _ := getWindowThreadProcessId.Call(uintptr(hWnd), uintptr(unsafe.Pointer(dwProcessId)))
	return uint32(r)
}

func MoveWindow(hWnd windows.HWND, x, y, width, height int32, repaint bool) bool {
	shouldRepaint := 0
	if repaint {
		shouldRepaint = 1
	}
	r, _, _ := moveWindow.Call(
		uintptr(hWnd),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(shouldRepaint),
	)
	return r != 0
}

func GetDpiForWindow(hWnd windows.HWND) uint32 {
	if err := getDpiForWindow.Find(); err != nil {
		return 0
	}
	dpi, _, _ := getDpiForWindow.Call(uintptr(hWnd))
	return uint32(dpi)
}

func SetWindowRgn(hWnd windows.HWND, region uintptr, redraw bool) bool {
	shouldRedraw := 0
	if redraw {
		shouldRedraw = 1
	}
	r, _, _ := setWindowRgn.Call(uintptr(hWnd), region, uintptr(shouldRedraw))
	return r != 0
}

func RefreshWindowFrame(hWnd windows.HWND) bool {
	flags := uint32(SwpNosize | SwpNomove | SwpNozorder | SwpNoactivate | SwpFramechanged)
	return SetWindowPos(hWnd, 0, 0, 0, 0, 0, flags)
}

func RedrawWindow(hWnd windows.HWND) bool {
	const flags = 0x0001 | 0x0080 | 0x0100 | 0x0400
	result, _, _ := redrawWindow.Call(uintptr(hWnd), 0, 0, flags)
	return result != 0
}

func EnableThreadPerMonitorV2DpiAwareness() uintptr {
	if err := setThreadDpiAwarenessContext.Find(); err != nil {
		return 0
	}
	previousContext, _, _ := setThreadDpiAwarenessContext.Call(^uintptr(3))
	return previousContext
}

func RestoreThreadDpiAwareness(previousContext uintptr) {
	if previousContext != 0 {
		setThreadDpiAwarenessContext.Call(previousContext)
	}
}
