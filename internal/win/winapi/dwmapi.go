package winapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type frameMargins struct {
	Left   int32
	Right  int32
	Top    int32
	Bottom int32
}

var dwmapi = windows.NewLazySystemDLL("dwmapi.dll")
var dwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")

func SetDwmBorderMargins(hWnd windows.HWND, bottomCrop int32) bool {
	const dwmwaBorderMargins = 40

	margins := frameMargins{}
	if bottomCrop > 0 {
		margins.Left = 1
		margins.Right = 1
		margins.Top = 1
		margins.Bottom = bottomCrop + 1
	}

	if err := dwmSetWindowAttribute.Find(); err != nil {
		return false
	}
	hresult, _, _ := dwmSetWindowAttribute.Call(
		uintptr(hWnd),
		dwmwaBorderMargins,
		uintptr(unsafe.Pointer(&margins)),
		unsafe.Sizeof(margins),
	)
	return int32(hresult) >= 0
}

func SetDwmCornerPreference(hWnd windows.HWND, doNotRound bool) bool {
	const dwmwaWindowCornerPreference = 33
	preference := int32(0)
	if doNotRound {
		preference = 1
	}
	hresult, _, _ := dwmSetWindowAttribute.Call(
		uintptr(hWnd),
		dwmwaWindowCornerPreference,
		uintptr(unsafe.Pointer(&preference)),
		unsafe.Sizeof(preference),
	)
	return int32(hresult) >= 0
}
