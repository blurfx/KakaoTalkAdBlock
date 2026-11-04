//go:build windows && amd64

package beta

import (
	"kakaotalkadblock/internal/win/winapi"
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

type windowClassEx struct {
	Size        uint32
	Style       uint32
	WindowProc  uintptr
	ClassExtra  int32
	WindowExtra int32
	Instance    uintptr
	Icon        uintptr
	Cursor      uintptr
	Background  uintptr
	Menu        *uint16
	ClassName   *uint16
	SmallIcon   uintptr
}

func TestRemoveAdsWindowLayouts(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	previousDpi := winapi.EnableThreadPerMonitorV2DpiAwareness()
	defer winapi.RestoreThreadDpiAwareness(previousDpi)
	user32 := windows.NewLazySystemDLL("user32.dll")
	registerClass := user32.NewProc("RegisterClassExW")
	unregisterClass := user32.NewProc("UnregisterClassW")
	createWindow := user32.NewProc("CreateWindowExW")
	destroyWindow := user32.NewProc("DestroyWindow")
	defWindowProc := user32.NewProc("DefWindowProcW")
	instance, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW").Call(0)
	windowProc := syscall.NewCallback(func(hwnd, message, wParam, lParam uintptr) uintptr {
		result, _, _ := defWindowProc.Call(hwnd, message, wParam, lParam)
		return result
	})
	for _, name := range []string{"Qt6AdBlockTestQWindowIcon", chromeWidgetContainer, chromeWidgetWindow, "Chrome_RenderWidgetHostHWND"} {
		className, _ := syscall.UTF16PtrFromString(name)
		class := windowClassEx{Size: uint32(unsafe.Sizeof(windowClassEx{})), WindowProc: windowProc, Instance: instance, ClassName: className}
		atom, _, err := registerClass.Call(uintptr(unsafe.Pointer(&class)))
		if atom == 0 {
			t.Fatalf("RegisterClassExW %s: %v", name, err)
		}
		defer unregisterClass.Call(uintptr(unsafe.Pointer(className)), instance)
	}
	create := func(class, title string, style uintptr, parent windows.HWND, x, y, width, height int32) windows.HWND {
		classPtr, _ := syscall.UTF16PtrFromString(class)
		titlePtr, _ := syscall.UTF16PtrFromString(title)
		handle, _, err := createWindow.Call(0x08000080, uintptr(unsafe.Pointer(classPtr)), uintptr(unsafe.Pointer(titlePtr)), style, uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(parent), 0, instance, 0)
		if handle == 0 {
			t.Fatalf("CreateWindowExW %s: %v", class, err)
		}
		return windows.HWND(handle)
	}
	for _, test := range []struct {
		name    string
		owned   bool
		nested  bool
		title   string
		blocked bool
	}{
		{"child-moment", false, false, momentAdWindowTitle, true},
		{"owned-moment", true, false, momentAdWindowTitle, true},
		{"child-adfit", false, false, "AdFit NAS Advertisement", true},
		{"owned-adfit", true, false, "AdFit NAS Advertisement", true},
		{"owned-adfit-nested", true, true, "AdFit NAS Advertisement", true},
		{"owned-video", true, false, "Video player", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			previousDpi := winapi.EnableThreadPerMonitorV2DpiAwareness()
			defer winapi.RestoreThreadDpiAwareness(previousDpi)
			mainWindowHandleMap = make(map[windows.HWND]struct{})
			ownedWindowHandleMap = make(map[windows.HWND]windows.HWND)
			windowClipMap = make(map[windows.HWND]windowClip)
			main := create("Qt6AdBlockTestQWindowIcon", "Main", 0x90000000, 0, -30000, -30000, 400, 600)
			defer destroyWindow.Call(uintptr(main))
			var ad windows.HWND
			if test.owned {
				ad = create("Qt6AdBlockTestQWindowIcon", qt6AdContainerTitle, 0x90000000, main, -30000, -29500, 400, 100)
				defer destroyWindow.Call(uintptr(ad))
			} else {
				ad = create("Qt6AdBlockTestQWindowIcon", qt6AdContainerTitle, 0x50000000, main, 0, 500, 400, 100)
			}
			container := ad
			if test.nested {
				container = create("Qt6AdBlockTestQWindowIcon", qt6AdContainerTitle, 0x50000000, ad, 4, 4, 392, 92)
			}
			host := create(chromeWidgetContainer, "", 0x50000000, container, 0, 0, 392, 92)
			renderer := create(chromeWidgetWindow, test.title, 0x50000000, host, 0, 0, 400, 100)
			create("Chrome_RenderWidgetHostHWND", chromeLegacyTitle, 0x50000000, renderer, 0, 0, 400, 100)
			TrackWindow(main, winapi.GetClassName(main), winapi.GetParent(main))
			if test.owned {
				TrackWindow(ad, winapi.GetClassName(ad), winapi.GetParent(ad))
			}
			RemoveAds()
			if blocked := !winapi.IsWindowVisible(ad); blocked != test.blocked {
				t.Errorf("ad hidden=%t, want %t", blocked, test.blocked)
			}
			if _, clipped := windowClipMap[main]; clipped != test.blocked {
				t.Errorf("main clipped=%t, want %t", clipped, test.blocked)
			}
			region := winapi.CreateRoundRectRgn(0, 0, 0, 0, 0, 0)
			defer winapi.DeleteObject(region)
			getWindowRgn := user32.NewProc("GetWindowRgn")
			regionType, _, _ := getWindowRgn.Call(uintptr(main), region)
			if test.blocked {
				var bounds winapi.Rect
				_, _, _ = windows.NewLazySystemDLL("gdi32.dll").NewProc("GetRgnBox").Call(region, uintptr(unsafe.Pointer(&bounds)))
				if regionType == 0 || bounds.Bottom < 495 || bounds.Bottom > 496 {
					t.Errorf("region type=%d bounds=%+v, want bottom between 495 and 496", regionType, bounds)
				}
			} else if regionType != 0 {
				t.Errorf("non-ad window region changed: type=%d", regionType)
			}
			Restore()
			if restored, _, _ := getWindowRgn.Call(uintptr(main), region); restored != 0 {
				t.Errorf("window region was not restored: type=%d", restored)
			}
		})
	}
}
