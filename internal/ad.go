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

func HidePopupAd() {
	var popupHandle windows.HWND
	for {
		popupHandle = winapi.FindWindowEx(0, popupHandle, "", "")
		if popupHandle == 0 {
			break
		}
		if winapi.GetParent(popupHandle) != 0 {
			continue
		}
		className := winapi.GetClassName(popupHandle)
		if !strings.Contains(className, "RichPopWnd") {
			continue
		}
		rect := new(winapi.Rect)
		_ = winapi.GetWindowRect(popupHandle, rect)
		width := rect.Right - rect.Left
		height := rect.Bottom - rect.Top
		if width == 300 && height == 150 {
			winapi.SendMessage(popupHandle, winapi.WmClose, 0, 0)
		}
	}
}

func HideMainWindowAd(windowClass string, handle windows.HWND) {
	// @deprecated
	if windowClass == "BannerAdWnd" {
		winapi.ShowWindow(handle, 0)
		winapi.SetWindowPos(handle, 0, 0, 0, 0, 0, winapi.SwpNomove)
	}
	if windowClass == "BannerAdContainer" {
		parentHandle := winapi.GetParent(handle)
		winapi.ShowWindow(parentHandle, 0)
		winapi.SetWindowPos(parentHandle, 0, 0, 0, 0, 0, winapi.SwpNomove)
	}
}

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
