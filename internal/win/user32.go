package win

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

const (
	SWP_NOMOVE = 0x0002
	WM_CLOSE   = 0x10
)

var (
	libuser32                = windows.NewLazySystemDLL("user32.dll")
	getClassName             = libuser32.NewProc("GetClassNameW")
	enumChildWindows         = libuser32.NewProc("EnumChildWindows")
	enumWindows              = libuser32.NewProc("EnumWindows")
	showWindow               = libuser32.NewProc("ShowWindow")
	findWindowEx             = libuser32.NewProc("FindWindowExW")
	getParent                = libuser32.NewProc("GetParent")
	setWindowPos             = libuser32.NewProc("SetWindowPos")
	getWindowText            = libuser32.NewProc("GetWindowTextW")
	getWindowRect            = libuser32.NewProc("GetWindowRect")
	updateWindow             = libuser32.NewProc("UpdateWindow")
	sendMessage              = libuser32.NewProc("SendMessageW")
	getWindowThreadProcessId = libuser32.NewProc("GetWindowThreadProcessId")
)

func cStr(str string) uintptr {
	strPtr, err := syscall.UTF16PtrFromString(str)
	if err != nil {
		panic(err)
	}
	return uintptr(unsafe.Pointer(strPtr))
}

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

func FindWindowEx(parent, child windows.HWND, className, windowName string) windows.HWND {
	r, _, _ := findWindowEx.Call(uintptr(parent), uintptr(child), cStr(className), cStr(windowName))
	return windows.HWND(r)
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
