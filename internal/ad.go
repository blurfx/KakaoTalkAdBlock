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

func HideMainWindowAd(windowClass string, handle windows.HWND) {
	// @deprecated
	if windowClass == "BannerAdWnd" {
		winapi.ShowWindow(handle, 0)
		winapi.MoveWindow(handle, 0, 0, 0, 0, true)
	}
	if windowClass == "BannerAdContainer" {
		parentHandle := winapi.GetParent(handle)
		winapi.ShowWindow(parentHandle, 0)
		winapi.MoveWindow(parentHandle, 0, 0, 0, 0, true)
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

func HidePopupAd(windowClass string, handle windows.HWND) {
	if windowClass == "AdFitWebView" {
		parentHandle := winapi.GetParent(handle)
		winapi.SendMessage(parentHandle, winapi.WmClose, 0, 0)
		winapi.ShowWindow(parentHandle, 0)
		winapi.MoveWindow(parentHandle, 0, 0, 0, 0, true)
		winapi.SendMessage(handle, winapi.WmClose, 0, 0)
		winapi.ShowWindow(handle, 0)
		winapi.MoveWindow(handle, 0, 0, 0, 0, true)
	}
}
