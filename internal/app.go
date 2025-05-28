package internal

import (
	"bytes"
	"context"
	"kakaotalkadblock/internal/win/winapi"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const sleepTime = 100 * time.Millisecond
const executable = "kakaotalk.exe"

var mutex = &sync.Mutex{}
var mainWindowHandleMap = make(map[windows.HWND]struct{})
var adSubwindowCandidateMap = make(map[windows.HWND]struct{})
var windowTextMap = make(map[windows.HWND]string)
var windowClassMap = make(map[windows.HWND]string)
var enumWindowCallbackMap = make(map[windows.HWND]uintptr)
var customScrollHandleMap = make(map[windows.HWND]bool)

func uint8ToStr(arr []uint8) string {
	n := bytes.Index(arr, []uint8{0})

	return string(arr[:n])
}

func watch(ctx context.Context) {
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
			className := winapi.GetClassName(handle)
			if className == "EVA_Window_Dblclk" || className == "EVA_Window" {
				windowText := winapi.GetWindowText(handle)
				parentHandle := winapi.GetParent(handle)

				switch className {
				case "EVA_Window_Dblclk":
					if windowText != "" && parentHandle == 0 {
						mainWindowHandleMap[handle] = struct{}{}
					} else if windowText == "" && parentHandle != 0 {
						if _, ok := mainWindowHandleMap[parentHandle]; ok {
							adSubwindowCandidateMap[handle] = struct{}{}
						}
					}
				case "EVA_Window":
					if windowText == "" && parentHandle == 0 {
						adSubwindowCandidateMap[handle] = struct{}{}
					}
				}
			}
		}
		return 1
	})
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mutex.Lock()
			if lastFoundAt < time.Now().Unix()-1 {
				snapshot = winapi.CreateToolhelp32Snapshot(winapi.Th32csSnapprocess, 0)
				lastFoundAt = time.Now().Unix()
			}
			if winapi.Process32First(uintptr(snapshot), &pe32) {
				for {
					szExeFile = uint8ToStr(pe32.SzExeFile[:])

					if strings.ToLower(szExeFile) == executable {
						winapi.EnumWindows(enumWindow, uintptr(pe32.Th32ProcessID))
					}

					if !winapi.Process32Next(uintptr(snapshot), &pe32) {
						break
					}
				}
			}
			mutex.Unlock()
		}
	}
}

func removeAd(ctx context.Context) {
	childHandles := make([]windows.HWND, 0)
	ticker := time.NewTicker(sleepTime)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mutex.Lock()
			for wnd := range mainWindowHandleMap {
				if wnd == 0 {
					continue
				}
				childHandles = childHandles[:0]
				var handle windows.HWND
				enumWindow, ok := enumWindowCallbackMap[wnd]
				if !ok {
					enumWindow = syscall.NewCallback(func(handle windows.HWND, _ uintptr) uintptr {
						childHandles = append(childHandles, handle)
						return 1
					})
					enumWindowCallbackMap[wnd] = enumWindow
				}
				winapi.EnumChildWindows(wnd, enumWindow, uintptr(unsafe.Pointer(&handle)))

				rect := new(winapi.Rect)
				winapi.GetWindowRect(wnd, rect)
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

					if className == "EVA_ChildWindow" && windowText == "" && parentText != "" {
						hasCustomScroll, ok := customScrollHandleMap[wnd]
						if !ok {
							hasCustomScroll = hasChildrenWithClassName(childHandle, "_EVA_")
							customScrollHandleMap[wnd] = hasCustomScroll
						}
						if !hasCustomScroll {
							winapi.SendMessage(childHandle, winapi.WmClose, 0, 0)
						}
					}
					HideMainViewAdArea(windowText, rect, childHandle)
					HideLockScreenAdArea(windowText, rect, childHandle)
				}
			}
			for wnd := range adSubwindowCandidateMap {
				if hasChildrenWithClassName(wnd, "Chrome Legacy Window") {
					winapi.ShowWindow(wnd, 0)
				}
			}
			mutex.Unlock()
		}
	}
}

func hasChildrenWithClassName(handle windows.HWND, className string) bool {
	childHandles := make([]windows.HWND, 0)

	enumWindow, ok := enumWindowCallbackMap[handle]
	if !ok {
		enumWindow = syscall.NewCallback(func(handle windows.HWND, _ uintptr) uintptr {
			childHandles = append(childHandles, handle)
			return 1
		})
		enumWindowCallbackMap[handle] = enumWindow
	}
	winapi.EnumChildWindows(handle, enumWindow, uintptr(unsafe.Pointer(&handle)))

	for _, wnd := range childHandles {
		if hasChildrenWithClassName(wnd, className) {
			return true
		}
	}

	windowClass, ok := windowClassMap[handle]
	if !ok {
		windowClass = winapi.GetClassName(handle)
		windowClassMap[handle] = windowClass
	}
	return strings.HasPrefix(windowClass, className)

}

func Run(ctx context.Context) {
	go watch(ctx)
	go removeAd(ctx)
}
