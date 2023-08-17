package internal

import (
	"bytes"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"kakaotalkadblock/internal/win"

	"golang.org/x/sys/windows"
)

const sleepTime = 100 * time.Millisecond

var mutex = &sync.Mutex{}
var handles = make([]windows.HWND, 0)

func uint8ToStr(arr []uint8) string {
	n := bytes.Index(arr, []uint8{0})

	return string(arr[:n])
}

func watch() {
	const executeable = "kakaotalk.exe"
	var (
		pe32      win.ProcessEntry32
		szExeFile string
	)

	snapshot := win.CreateToolhelp32Snapshot(win.TH32CS_SNAPPROCESS, 0)
	pe32.DwSize = uint32(unsafe.Sizeof(pe32))

	var enumWindow = syscall.NewCallback(func(handle windows.HWND, processId uintptr) uintptr {
		win.GetWindowThreadProcessId(handle, &pe32.Th32ProcessID)

		if processId == uintptr(pe32.Th32ProcessID) {
			handles = append(handles, handle)
		}
		return 1
	})

	for {
		mutex.Lock()
		handles = handles[:0]

		if win.Process32First(uintptr(snapshot), &pe32) {
			for {
				szExeFile = uint8ToStr(pe32.SzExeFile[:])

				if strings.ToLower(szExeFile) == executeable {
					win.EnumWindows(enumWindow, uintptr(pe32.Th32ProcessID))
					break
				}

				if !win.Process32Next(uintptr(snapshot), &pe32) {
					break
				}
			}
		}
		mutex.Unlock()
		time.Sleep(sleepTime)
	}
}

func removeAd() {
	childHandles := make([]windows.HWND, 0)

	var enumWindow = syscall.NewCallback(func(handle windows.HWND, _ uintptr) uintptr {
		childHandles = append(childHandles, handle)
		return 1
	})
	for {
		mutex.Lock()
		for _, wnd := range handles {
			if wnd == 0 {
				continue
			}
			childHandles = childHandles[:0]
			var handle windows.HWND
			win.EnumChildWindows(wnd, enumWindow, uintptr(unsafe.Pointer(&handle)))

			rect := new(win.Rect)
			win.GetWindowRect(wnd, rect)
			for _, childHandle := range childHandles {
				className := win.GetClassName(childHandle)
				windowText := win.GetWindowText(childHandle)
				HideMainWindowAd(className, childHandle)
				HideMainViewAdArea(windowText, rect, childHandle)
				HideLockScreenAdArea(windowText, rect, childHandle)
			}
		}
		HidePopupAd()
		mutex.Unlock()
		time.Sleep(sleepTime)
	}
}

func Run() {
	go watch()
	go removeAd()

	select {}
}
