package beta

import (
	"kakaotalkadblock/internal/win/winapi"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	executable            = "kakaotalkui.exe"
	qt6WindowClassPrefix  = "Qt6"
	qtWindowClassSuffix   = "QWindowIcon"
	chromeWidgetWindow    = "Chrome_WidgetWin_1"
	chromeWidgetContainer = "Chrome_WidgetWin_0"
	momentAdWindowTitle   = "MOMENT 광고"
	chromeLegacyTitle     = "Chrome Legacy Window"
	qt6AdContainerTitle   = "KakaoTalkUI"
	clipExtraPadding      = 6
	cornerDiameter        = 16
	referenceDPI          = 144
)

type windowClip struct {
	adHeight            int32
	bottomInset         int32
	dpi                 uint32
	adWindow            windows.HWND
	originalStyle       uintptr
	styleChanged        bool
	dwmMarginsSupported int8
	appliedWidth        int32
	appliedHeight       int32
	appliedCropHeight   int32
	appliedDpi          uint32
	refreshTicks        uint8
}

var mutex = &sync.Mutex{}
var mainWindowHandleMap = make(map[windows.HWND]struct{})
var windowClipMap = make(map[windows.HWND]windowClip)
var childWindowCollector []windows.HWND
var childWindowCollectorCallback = syscall.NewCallback(func(handle windows.HWND, _ uintptr) uintptr {
	childWindowCollector = append(childWindowCollector, handle)
	return 1
})

func IsExecutable(name string) bool {
	return strings.EqualFold(name, executable)
}

func TrackWindow(handle windows.HWND, className string, parentHandle windows.HWND) {
	if !isQt6WindowClass(className) || parentHandle != 0 {
		return
	}
	mutex.Lock()
	mainWindowHandleMap[handle] = struct{}{}
	mutex.Unlock()
}

func RemoveAds() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	previousDpiContext := winapi.EnableThreadPerMonitorV2DpiAwareness()
	defer winapi.RestoreThreadDpiAwareness(previousDpiContext)

	mutex.Lock()
	defer mutex.Unlock()
	for mainWindow := range mainWindowHandleMap {
		if !winapi.IsWindow(mainWindow) {
			delete(mainWindowHandleMap, mainWindow)
			delete(windowClipMap, mainWindow)
			continue
		}
		if !winapi.IsWindowVisible(mainWindow) {
			continue
		}
		hideMomentAds(mainWindow)
	}
}

func Restore() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	previousDpiContext := winapi.EnableThreadPerMonitorV2DpiAwareness()
	defer winapi.RestoreThreadDpiAwareness(previousDpiContext)

	mutex.Lock()
	defer mutex.Unlock()
	for mainWindow := range windowClipMap {
		clearWindowClip(mainWindow)
	}
}

func isQt6WindowClass(className string) bool {
	return strings.HasPrefix(className, qt6WindowClassPrefix) && strings.HasSuffix(className, qtWindowClassSuffix)
}

func hideMomentAds(mainWindow windows.HWND) {
	if clip, known := windowClipMap[mainWindow]; known {
		if winapi.IsWindow(clip.adWindow) {
			clipWindowForAd(mainWindow, clip.adWindow)
			return
		}
	}

	for _, handle := range getChildWindows(mainWindow) {
		if winapi.GetClassName(handle) != chromeWidgetWindow || winapi.GetWindowText(handle) != momentAdWindowTitle {
			continue
		}
		if !containsChromeLegacyWindow(handle) {
			continue
		}

		chromeHost := winapi.GetParent(handle)
		if winapi.GetClassName(chromeHost) != chromeWidgetContainer {
			continue
		}
		adWindow := winapi.GetParent(chromeHost)
		if !isQt6WindowClass(winapi.GetClassName(adWindow)) || winapi.GetWindowText(adWindow) != qt6AdContainerTitle {
			continue
		}

		clipWindowForAd(mainWindow, adWindow)
	}
}

func clipWindowForAd(mainWindow, adWindow windows.HWND) {
	if !winapi.IsWindow(mainWindow) || !winapi.IsWindowVisible(mainWindow) || !winapi.IsWindow(adWindow) {
		return
	}
	mainRect := new(winapi.Rect)
	if !winapi.GetWindowRect(mainWindow, mainRect) {
		return
	}
	mainWidth := mainRect.Right - mainRect.Left
	mainHeight := mainRect.Bottom - mainRect.Top
	if mainWidth < 1 || mainHeight < 1 {
		return
	}

	currentDpi := winapi.GetDpiForWindow(mainWindow)
	if currentDpi == 0 {
		currentDpi = 96
	}

	clip, known := windowClipMap[mainWindow]
	clip.adWindow = adWindow

	adRect := new(winapi.Rect)
	measured := false
	if winapi.GetWindowRect(adWindow, adRect) {
		if adHeight, bottomInset, ok := measureAd(mainRect, adRect); ok {
			clip.adHeight = adHeight
			clip.bottomInset = bottomInset
			clip.dpi = currentDpi
			measured = true
		}
	}

	if !measured {
		if !known || clip.adHeight < 1 || clip.dpi == 0 {
			return
		}
		if clip.dpi != currentDpi {
			clip.adHeight = scaleDpiValue(clip.adHeight, clip.dpi, currentDpi)
			clip.bottomInset = scaleDpiValue(clip.bottomInset, clip.dpi, currentDpi)
			clip.dpi = currentDpi
		}
	}

	extraPadding := scaleDpiValue(clipExtraPadding, referenceDPI, currentDpi)
	cropHeight := clip.adHeight + clip.bottomInset + extraPadding
	visibleHeight := mainHeight - cropHeight
	if visibleHeight < 1 {
		return
	}

	if measured || winapi.IsWindowVisible(adWindow) {
		collapseAdWindow(adWindow)
	}
	if !winapi.IsWindow(mainWindow) || !winapi.IsWindowVisible(mainWindow) {
		return
	}
	geometryChanged := clip.appliedWidth != mainWidth ||
		clip.appliedHeight != mainHeight ||
		clip.appliedCropHeight != cropHeight ||
		clip.appliedDpi != currentDpi
	delayedRefresh := clip.refreshTicks > 0
	frameChanged := false
	useDwmMargins := clip.dwmMarginsSupported == 1
	if clip.dwmMarginsSupported == 0 || (clip.dwmMarginsSupported == 1 && geometryChanged) {
		useDwmMargins = winapi.SetDwmBorderMargins(mainWindow, cropHeight)
		if useDwmMargins {
			clip.dwmMarginsSupported = 1
		} else {
			clip.dwmMarginsSupported = -1
		}
	}
	if useDwmMargins {
		if clip.styleChanged {
			currentStyle := winapi.GetWindowStyle(mainWindow)
			restoredStyle := clip.originalStyle | currentStyle&winapi.WindowStateStyles
			winapi.SetWindowStyle(mainWindow, restoredStyle)
			winapi.SetDwmCornerPreference(mainWindow, false)
			clip.styleChanged = false
			frameChanged = true
		}
	} else {
		currentStyle := winapi.GetWindowStyle(mainWindow)
		if !clip.styleChanged {
			clip.originalStyle = currentStyle &^ winapi.WindowStateStyles
		}
		targetStyle := (clip.originalStyle | currentStyle&winapi.WindowStateStyles) &^ winapi.DwmFrameWindowStyles
		if winapi.IsZoomed(mainWindow) {
			targetStyle |= clip.originalStyle & winapi.DwmFrameWindowStyles
		}
		if currentStyle != targetStyle && winapi.SetWindowStyle(mainWindow, targetStyle) {
			frameChanged = true
		}
		if winapi.GetWindowStyle(mainWindow) == targetStyle {
			styleChanged := targetStyle&winapi.DwmFrameWindowStyles != clip.originalStyle&winapi.DwmFrameWindowStyles
			if clip.styleChanged != styleChanged {
				winapi.SetDwmCornerPreference(mainWindow, styleChanged)
				frameChanged = true
			}
			clip.styleChanged = styleChanged
		}
	}

	if frameChanged || delayedRefresh || (clip.dwmMarginsSupported == 1 && geometryChanged) {
		winapi.RefreshWindowFrame(mainWindow)
	}
	if geometryChanged || frameChanged {
		if !winapi.IsWindow(mainWindow) || !winapi.IsWindowVisible(mainWindow) {
			return
		}
		cornerSize := scaleDpiValue(cornerDiameter, referenceDPI, currentDpi)
		region := winapi.CreateRoundRectRgn(0, 0, mainWidth, visibleHeight, cornerSize, cornerSize)
		if region == 0 {
			return
		}
		if !winapi.SetWindowRgn(mainWindow, region, true) {
			winapi.DeleteObject(region)
			return
		}
	}
	if geometryChanged || frameChanged {
		clip.refreshTicks = 2
	}
	if geometryChanged || frameChanged || delayedRefresh {
		winapi.RedrawWindow(mainWindow)
	}
	if delayedRefresh {
		clip.refreshTicks--
	}
	clip.appliedWidth = mainWidth
	clip.appliedHeight = mainHeight
	clip.appliedCropHeight = cropHeight
	clip.appliedDpi = currentDpi
	windowClipMap[mainWindow] = clip
}

func measureAd(mainRect, adRect *winapi.Rect) (adHeight, bottomInset int32, ok bool) {
	if mainRect == nil || adRect == nil {
		return 0, 0, false
	}
	mainWidth := mainRect.Right - mainRect.Left
	mainHeight := mainRect.Bottom - mainRect.Top
	adWidth := adRect.Right - adRect.Left
	adHeight = adRect.Bottom - adRect.Top
	bottomInset = mainRect.Bottom - adRect.Bottom
	horizontalOverlap := min(mainRect.Right, adRect.Right) - max(mainRect.Left, adRect.Left)

	if mainWidth < 1 || mainHeight < 1 || adWidth < 1 || adHeight < 1 {
		return 0, 0, false
	}
	if adHeight >= mainHeight/2 || bottomInset < 0 || bottomInset >= mainHeight/3 {
		return 0, 0, false
	}
	if horizontalOverlap < mainWidth/2 {
		return 0, 0, false
	}
	return adHeight, bottomInset, true
}

func scaleDpiValue(value int32, fromDpi, toDpi uint32) int32 {
	if value <= 0 || fromDpi == 0 || toDpi == 0 || fromDpi == toDpi {
		return value
	}
	return int32((int64(value)*int64(toDpi) + int64(fromDpi)/2) / int64(fromDpi))
}

func collapseAdWindow(adWindow windows.HWND) {
	if !winapi.IsWindow(adWindow) {
		return
	}
	winapi.ShowWindow(adWindow, 0)
	if !winapi.IsWindow(adWindow) {
		return
	}
	flags := uint32(winapi.SwpNomove | winapi.SwpNozorder | winapi.SwpNoactivate)
	winapi.SetWindowPos(adWindow, 0, 0, 0, 0, 0, flags)
}

func clearWindowClip(mainWindow windows.HWND) {
	if !winapi.IsWindow(mainWindow) {
		return
	}
	if clip, known := windowClipMap[mainWindow]; known && clip.styleChanged {
		currentStyle := winapi.GetWindowStyle(mainWindow)
		restoredStyle := clip.originalStyle | currentStyle&winapi.WindowStateStyles
		winapi.SetWindowStyle(mainWindow, restoredStyle)
	}
	winapi.SetWindowRgn(mainWindow, 0, true)
	winapi.SetDwmBorderMargins(mainWindow, 0)
	winapi.SetDwmCornerPreference(mainWindow, false)
	winapi.RefreshWindowFrame(mainWindow)
	winapi.RedrawWindow(mainWindow)
}

func containsChromeLegacyWindow(handle windows.HWND) bool {
	for _, child := range getChildWindows(handle) {
		if winapi.GetWindowText(child) == chromeLegacyTitle {
			return true
		}
	}
	return false
}

func getChildWindows(parent windows.HWND) []windows.HWND {
	childWindowCollector = childWindowCollector[:0]
	winapi.EnumChildWindows(parent, childWindowCollectorCallback, 0)
	return append([]windows.HWND(nil), childWindowCollector...)
}
