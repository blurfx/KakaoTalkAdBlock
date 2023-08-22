package win

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"kakaotalkadblock/internal/win/winapi"
)

var quit *chan struct{}

func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case winapi.WmTrayicon:
		switch uint16(lParam) {
		case winapi.WmLbuttondblclk:
			close(*quit)
		}
	case winapi.WmDestroy:
		winapi.PostQuitMessage(0)
	default:
		return winapi.DefWindowProc(hWnd, msg, wParam, lParam)
	}
	return 0
}

func createMainWindow() (uintptr, error) {
	hInstance, err := winapi.GetModuleHandle(nil)
	if err != nil {
		return 0, err
	}

	wndClass, _ := windows.UTF16PtrFromString("KakaoTalkAdBlock")

	var windowClass winapi.WindowClassEx

	windowClass.CbSize = uint32(unsafe.Sizeof(windowClass))
	windowClass.LpfnWndProc = windows.NewCallback(wndProc)
	windowClass.HInstance = hInstance
	windowClass.LpszClassName = wndClass
	if _, err := winapi.RegisterClassEx(&windowClass); err != nil {
		return 0, err
	}

	handle, err := winapi.CreateWindowEx(
		0,
		wndClass,
		windows.StringToUTF16Ptr("KakaoTalkAdBlock"),
		winapi.WsOverlappedwindow,
		winapi.CwUsedefault,
		winapi.CwUsedefault,
		1,
		1,
		0,
		0,
		hInstance,
		nil)
	if err != nil {
		return 0, err
	}

	return handle, nil
}

type TrayIcon struct {
	notifyIconData winapi.NotifyIconData
}

func NewTrayIcon(quitChan *chan struct{}) *TrayIcon {
	var data winapi.NotifyIconData
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.UFlags = winapi.NifIcon | winapi.NifMessage | winapi.NifInfo
	data.UCallbackMessage = winapi.WmTrayicon

	hInst, err := winapi.GetModuleHandle(nil)
	if err != nil {
		panic(err)
	}
	icon, err := winapi.LoadIcon(hInst, winapi.MakeIntResource(1))
	if err != nil {
		panic(err)
	}
	data.HIcon = icon

	quit = quitChan
	return &TrayIcon{
		notifyIconData: data,
	}
}

func (t *TrayIcon) Show() {
	if t.notifyIconData.HWnd == 0 {
		handle, err := createMainWindow()
		if err != nil {
			panic(err)
		}
		t.notifyIconData.HWnd = handle
	}
	if err := winapi.ShellNotifyIcon(winapi.NimAdd, &t.notifyIconData); err != nil {
		panic(err)
	}
}

func (t *TrayIcon) Hide() {
	if err := winapi.ShellNotifyIcon(winapi.NimDelete, &t.notifyIconData); err != nil {
		panic(err)
	}
}
