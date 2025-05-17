package internal

import (
	"bytes"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"kakaotalkadblock/internal/win"
	"kakaotalkadblock/internal/win/winapi"

	"golang.org/x/sys/windows"
)

const sleepTime = 100 * time.Millisecond
const executable = "kakaotalk.exe"

var mutex = &sync.Mutex{}
var handles = make([]windows.HWND, 0)
var windowTextMap = make(map[windows.HWND]string)
var windowClassMap = make(map[windows.HWND]string)

func uint8ToStr(arr []uint8) string {
	n := bytes.Index(arr, []uint8{0})

	return string(arr[:n])
}

func watch() {
	var (
		pe32      winapi.ProcessEntry32
		szExeFile string
	)
	pe32.DwSize = uint32(unsafe.Sizeof(pe32))
	lastFoundAt := time.Now().Unix() - 2
	var snapshot windows.HWND
	var enumWindow = syscall.NewCallback(func(handle windows.HWND, processId uintptr) uintptr {
		winapi.GetWindowThreadProcessId(handle, &pe32.Th32ProcessID)
		if processId == uintptr(pe32.Th32ProcessID) {
			lastFoundAt = time.Now().Unix()
			handles = append(handles, handle)
		}
		return 1
	})

	for {
		mutex.Lock()
		handles = handles[:0]
		if lastFoundAt < time.Now().Unix()-1 {
			snapshot = winapi.CreateToolhelp32Snapshot(winapi.Th32csSnapprocess, 0)
			lastFoundAt = time.Now().Unix()
		}

		if winapi.Process32First(uintptr(snapshot), &pe32) {
			for {
				szExeFile = uint8ToStr(pe32.SzExeFile[:])

				if strings.ToLower(szExeFile) == executable {
					winapi.EnumWindows(enumWindow, uintptr(pe32.Th32ProcessID))
					//break
				}

				if !winapi.Process32Next(uintptr(snapshot), &pe32) {
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
			winapi.EnumChildWindows(wnd, enumWindow, uintptr(unsafe.Pointer(&handle)))

			rect := new(winapi.Rect)
			winapi.GetWindowRect(wnd, rect)
			var candidates [][]windows.HWND
			for _, childHandle := range childHandles {
				className, ok := windowClassMap[childHandle]
				if !ok {
					className = winapi.GetClassName(childHandle)
					windowClassMap[childHandle] = className
				}
				windowText, ok := windowTextMap[childHandle]
				if !ok {
					windowText = winapi.GetWindowText(childHandle)
					windowTextMap[childHandle] = windowText
				}
				parentHandle := winapi.GetParent(childHandle)
				if parentHandle != wnd {
					continue
				}
				parentText, ok := windowTextMap[parentHandle]
				if !ok {
					parentText = winapi.GetWindowText(parentHandle)
					windowTextMap[parentHandle] = parentText
				}
				if className != "EVA_ChildWindow" && windowText == "" && parentText != "" {
					winapi.SendMessage(childHandle, winapi.WmClose, 0, 0)
					candidates = append(candidates, []windows.HWND{childHandle, parentHandle})
				}
				HideMainWindowAd(className, childHandle)
				HideMainViewAdArea(windowText, rect, childHandle)
				HideLockScreenAdArea(windowText, rect, childHandle)
				HidePopupAd(className, childHandle)
			}
		}
		mutex.Unlock()
		time.Sleep(sleepTime)
	}
}

func Run() {
	var quit = make(chan struct{})
	trayIcon := win.NewTrayIcon(&quit)
	trayIcon.Show()
	defer trayIcon.Hide()
	go watch()
	go removeAd()

	select {
	case <-quit:
		return
	}
}
