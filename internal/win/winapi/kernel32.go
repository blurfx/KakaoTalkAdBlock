package winapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type WindowClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type ProcessEntry32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      uint32
	DwFlags             uint32
	SzExeFile           [MaxPath]uint8
}

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	createToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	process32First           = kernel32.NewProc("Process32First")
	process32Next            = kernel32.NewProc("Process32Next")
	getModuleHandle          = kernel32.NewProc("GetModuleHandleW")
)

func CreateToolhelp32Snapshot(flags uint32, pid uint32) windows.HWND {
	handle, _, _ := createToolhelp32Snapshot.Call(uintptr(flags), uintptr(pid))
	return windows.HWND(handle)
}

func Process32First(hSnapshot uintptr, pe *ProcessEntry32) bool {
	ret, _, _ := process32First.Call(
		hSnapshot,
		uintptr(unsafe.Pointer(pe)),
	)
	return ret != 0
}

func Process32Next(hSnapshot uintptr, pe *ProcessEntry32) bool {
	ret, _, _ := process32Next.Call(
		hSnapshot,
		uintptr(unsafe.Pointer(pe)),
	)
	return ret != 0
}

func GetModuleHandle(lpModuleName *uint16) (uintptr, error) {
	r, _, err := getModuleHandle.Call(uintptr(unsafe.Pointer(lpModuleName)))
	if r == 0 {
		return 0, err
	}
	return r, nil
}
