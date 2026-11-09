package internal

import (
	"strings"

	"golang.org/x/sys/windows"

	"kakaotalkadblock/internal/win/winapi"
)

const (
	LayoutShadowPadding = 2
	MainViewPadding     = 31
)

func HideLockScreenAdArea(windowText string, rect *winapi.Rect, handle windows.HWND) {
	if strings.HasPrefix(windowText, "LockModeView") {
		width := rect.Right - rect.Left - LayoutShadowPadding
		height := rect.Bottom - rect.Top
		resizeView(handle, width, height)
	}
}

func HideMainViewAdArea(windowText string, rect *winapi.Rect, handle windows.HWND) {
	if strings.HasPrefix(windowText, "OnlineMainView") {
		width := rect.Right - rect.Left - LayoutShadowPadding
		height := rect.Bottom - rect.Top - MainViewPadding
		if height < 1 {
			return
		}
		resizeView(handle, width, height)
	}
}

func resizeView(handle windows.HWND, width, height int32) {
	current := new(winapi.Rect)
	// Resizing raises a WinEvent, which would trigger another pass.
	if winapi.GetWindowRect(handle, current) && current.Right-current.Left == width && current.Bottom-current.Top == height {
		return
	}
	winapi.UpdateWindow(handle)
	winapi.SetWindowPos(handle, 0, 0, 0, width, height, winapi.SwpNomove)
}
