package internal

import (
	"strings"

	"golang.org/x/sys/windows"

	"kakaotalkadblock/internal/win"
)

const (
	LayoutShadowPadding = 2
	MainViewPadding     = 31
)

func HidePopupAd() {
	var popupHandle windows.HWND
	for {
		popupHandle = win.FindWindowEx(0, popupHandle, "", "")
		if popupHandle == 0 {
			break
		}
		if win.GetParent(popupHandle) != 0 {
			continue
		}
		className := win.GetClassName(popupHandle)
		if !strings.Contains(className, "RichPopWnd") {
			continue
		}
		rect := new(win.Rect)
		_ = win.GetWindowRect(popupHandle, rect)
		width := rect.Right - rect.Left
		height := rect.Bottom - rect.Top
		if width == 300 && height == 150 {
			win.SendMessage(popupHandle, win.WM_CLOSE, 0, 0)
		}
	}
}

func HideMainWindowAd(windowClass string, handle windows.HWND) {
	if windowClass == "BannerAdWnd" {
		win.ShowWindow(handle, 0)
		win.SetWindowPos(handle, 0, 0, 0, 0, 0, win.SWP_NOMOVE)
	}
}

func HideLockScreenAdArea(windowText string, rect *win.Rect, handle windows.HWND) {
	if strings.HasPrefix(windowText, "LockModeView") {
		width := rect.Right - rect.Left - LayoutShadowPadding
		height := rect.Bottom - rect.Top
		win.UpdateWindow(handle)
		win.SetWindowPos(handle, 0, 0, 0, width, height, win.SWP_NOMOVE)
	}
}

func HideMainViewAdArea(windowText string, rect *win.Rect, handle windows.HWND) {
	if strings.HasPrefix(windowText, "OnlineMainView") {
		width := rect.Right - rect.Left - LayoutShadowPadding
		height := rect.Bottom - rect.Top - MainViewPadding
		if height < 1 {
			return
		}
		win.UpdateWindow(handle)
		win.SetWindowPos(handle, 0, 0, 0, width, height, win.SWP_NOMOVE)
	}
}
