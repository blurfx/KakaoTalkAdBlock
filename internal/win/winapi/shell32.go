package winapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type NotifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         GUID
	HBalloonIcon     uintptr
}

var (
	shell32         = windows.NewLazySystemDLL("shell32.dll")
	shellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")
)

func MakeIntResource(i uint16) *uint16 {
	return (*uint16)(unsafe.Pointer(uintptr(i)))
}

func ShellNotifyIcon(dwMessage uintptr, notifyIconData *NotifyIconData) error {
	r, _, err := shellNotifyIcon.Call(dwMessage, uintptr(unsafe.Pointer(notifyIconData)))
	if r == 0 {
		return err
	}
	return nil
}
