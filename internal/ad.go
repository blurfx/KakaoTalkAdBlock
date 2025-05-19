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
		winapi.UpdateWindow(handle)
		winapi.SetWindowPos(handle, 0, 0, 0, width, height, winapi.SwpNomove)
	}
}

func HideMainViewAdArea(windowText string, rect *winapi.Rect, handle windows.HWND) {
	if strings.HasPrefix(windowText, "OnlineMainView") {
		width := rect.Right - rect.Left - LayoutShadowPadding
		height := rect.Bottom - rect.Top - MainViewPadding
		if height < 1 {
			return
		}
		winapi.UpdateWindow(handle)
		winapi.SetWindowPos(handle, 0, 0, 0, width, height, winapi.SwpNomove)
	}
}
