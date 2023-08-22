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
	user32                   = windows.NewLazySystemDLL("user32.dll")
	getClassName             = user32.NewProc("GetClassNameW")
	enumChildWindows         = user32.NewProc("EnumChildWindows")
	enumWindows              = user32.NewProc("EnumWindows")
	showWindow               = user32.NewProc("ShowWindow")
	findWindowEx             = user32.NewProc("FindWindowExW")
	getParent                = user32.NewProc("GetParent")
	setWindowPos             = user32.NewProc("SetWindowPos")
	getWindowText            = user32.NewProc("GetWindowTextW")
	getWindowRect            = user32.NewProc("GetWindowRect")
	updateWindow             = user32.NewProc("UpdateWindow")
	sendMessage              = user32.NewProc("SendMessageW")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	loadIcon                 = user32.NewProc("LoadIconW")
	postQuitMessage          = user32.NewProc("PostQuitMessage")
	defWindowProc            = user32.NewProc("DefWindowProcW")
	registerClassEx          = user32.NewProc("RegisterClassExW")
	createWindowEx           = user32.NewProc("CreateWindowExW")
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

func PostQuitMessage(nExitCode int32) {
	r, _, err := postQuitMessage.Call(uintptr(nExitCode))
	if r == 0 {
		panic(err)
	}
}

func DefWindowProc(
	hWnd uintptr,
	Msg uint32,
	wParam, lParam uintptr) uintptr {
	r, _, _ := defWindowProc.Call(
		hWnd,
		uintptr(Msg),
		wParam,
		lParam)
	return r
}

func RegisterClassEx(Arg1 *WindowClassEx) (uint16, error) {
	r, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(Arg1)))
	if r == 0 {
		return 0, err
	}
	return uint16(r), nil
}

func CreateWindowEx(
	dwExStyle uint32,
	lpClassName, lpWindowName *uint16,
	dwStyle uint32,
	X, Y, nWidth, nHeight int32,
	hWndParent, hMenu, hInstance uintptr,
	lpParam unsafe.Pointer) (uintptr, error) {
	r, _, err := createWindowEx.Call(
		uintptr(dwExStyle),
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)),
		uintptr(dwStyle),
		uintptr(X),
		uintptr(Y),
		uintptr(nWidth),
		uintptr(nHeight),
		hWndParent,
		hMenu,
		hInstance,
		uintptr(lpParam))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func LoadIcon(instance uintptr, iconName *uint16) (uintptr, error) {
	ret, _, err := loadIcon.Call(
		instance,
		uintptr(unsafe.Pointer(iconName)),
	)
	if ret == 0 {
		return 0, err
	}
	return ret, nil
}
