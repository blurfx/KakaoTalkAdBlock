package internal

import (
	"bytes"
	"context"
	"kakaotalkadblock/internal/beta"
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
var customScrollHandleMap = make(map[windows.HWND]bool)

var childWindowCollector []windows.HWND
var childWindowCollectorCallback = syscall.NewCallback(func(handle windows.HWND, _ uintptr) uintptr {
	childWindowCollector = append(childWindowCollector, handle)
	return 1
})

func getChildWindows(parent windows.HWND) []windows.HWND {
	childWindowCollector = childWindowCollector[:0]
	winapi.EnumChildWindows(parent, childWindowCollectorCallback, 0)
	return append([]windows.HWND(nil), childWindowCollector...)
}

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
			parentHandle := winapi.GetParent(handle)
			beta.TrackWindow(handle, className, parentHandle)
			if className == "EVA_Window_Dblclk" || className == "EVA_Window" {
				windowText := winapi.GetWindowText(handle)

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
	// Closure, not `defer winapi.CloseHandle(snapshot)`: deferred arguments are
	// evaluated immediately, which would capture the zero handle.
	defer func() { winapi.CloseHandle(snapshot) }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mutex.Lock()
			if lastFoundAt < time.Now().Unix()-1 {
				// Release the previous snapshot before taking a new one.
				// Without this every refresh leaks a kernel snapshot object,
				// which happens roughly every 2s while KakaoTalk is not running.
				winapi.CloseHandle(snapshot)
				snapshot = winapi.CreateToolhelp32Snapshot(winapi.Th32csSnapprocess, 0)
				lastFoundAt = time.Now().Unix()
			}
			if winapi.Process32First(uintptr(snapshot), &pe32) {
				for {
					szExeFile = uint8ToStr(pe32.SzExeFile[:])

					if strings.ToLower(szExeFile) == executable || beta.IsExecutable(szExeFile) {
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
	ticker := time.NewTicker(sleepTime)

	defer ticker.Stop()
	defer beta.Restore()
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
				childHandles := getChildWindows(wnd)
				if !isMainWindow(childHandles) {
					continue
				}

				rect := new(winapi.Rect)
				winapi.GetWindowRect(wnd, rect)
				for _, childHandle := range childHandles[1:] {
					className := getWindowClass(childHandle)
					windowText := getWindowText(childHandle)
					parentHandle := winapi.GetParent(childHandle)
					if parentHandle != wnd {
						continue
					}
					parentText := getWindowText(parentHandle)

					if className == "EVA_ChildWindow" && windowText == "" && parentText != "" {
						hasCustomScroll, ok := customScrollHandleMap[wnd]
						if !ok {
							hasCustomScroll = classNameStartsWith(childHandle, "_EVA_")
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
				if !winapi.IsWindow(wnd) {
					delete(adSubwindowCandidateMap, wnd)
					delete(windowTextMap, wnd)
					delete(windowClassMap, wnd)
					continue
				}
				if winapi.IsWindowVisible(wnd) && hasChromeLegacyWindow(wnd) {
					winapi.ShowWindow(wnd, 0)
				}
			}
			beta.RemoveAds()
			mutex.Unlock()
		}
	}
}

func classNameStartsWith(handle windows.HWND, className string) bool {
	if strings.HasPrefix(getWindowClass(handle), className) {
		return true
	}
	for _, wnd := range getChildWindows(handle) {
		if strings.HasPrefix(getWindowClass(wnd), className) {
			return true
		}
	}
	return false
}

func hasChromeLegacyWindow(handle windows.HWND) bool {
	if getWindowText(handle) == "Chrome Legacy Window" {
		return true
	}
	for _, wnd := range getChildWindows(handle) {
		if getWindowText(wnd) == "Chrome Legacy Window" {
			return true
		}
	}
	return false
}

func isMainWindow(handles []windows.HWND) bool {
	for _, wnd := range handles {
		windowClass := getWindowClass(wnd)
		if windowClass != "EVA_ChildWindow" {
			continue
		}

		windowText := winapi.GetWindowText(wnd)
		if strings.HasPrefix(windowText, "OnlineMainView") || strings.HasPrefix(windowText, "LockModeView") {
			return true
		}
	}
	return false
}

func getWindowText(handle windows.HWND) string {
	text, ok := windowTextMap[handle]
	if !ok {
		text = winapi.GetWindowText(handle)
		windowTextMap[handle] = text
	}
	return text
}

func getWindowClass(handle windows.HWND) string {
	class, ok := windowClassMap[handle]
	if !ok {
		class = winapi.GetClassName(handle)
		windowClassMap[handle] = class
	}
	return class
}

func Run(ctx context.Context) {
	go watch(ctx)
	go removeAd(ctx)
}
