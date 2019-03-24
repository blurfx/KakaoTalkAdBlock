using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text;

namespace KakaoTalkAdBlock
{
    class Program
    {
        [DllImport("user32.dll")]
        static extern int GetClassName(IntPtr hWnd, StringBuilder lpClassName, int nMaxCount);

        [DllImport("user32.dll")]
        static extern bool EnumChildWindows(IntPtr WindowHandle, EnumWindowProcess Callback, IntPtr lParam);

        [DllImport("user32.dll")]
        static extern Boolean ShowWindow(IntPtr hWnd, int nCmdShow);

        [DllImport("user32.dll")]
        static extern IntPtr FindWindow(string lpClassName, string lpWindowName);

        [DllImport("user32.dll")]
        static extern IntPtr GetParent(IntPtr hWnd);

        [DllImport("user32.dll", EntryPoint = "SetWindowPos", SetLastError = false)]
        public static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, int uFlags);

        public static class HwndInsertAfterInt
        {
            public static readonly IntPtr NoTopMost = new IntPtr(-2);
            // Places the window at the bottom of the Z order. If the hWnd parameter identifies a topmost window, the window loses its topmost status and is placed at the bottom of all other windows.

            public static readonly IntPtr TopMost = new IntPtr(-1);
            //Places the window above all non-topmost windows (that is, behind all topmost windows). This flag has no effect if the window is already a non-topmost window.

            public static readonly IntPtr Top = new IntPtr(0);
            // Places the window at the top of the Z order.

            public static readonly IntPtr Bottom = new IntPtr(1);
            //Places the window above all non-topmost windows. The window maintains its topmost position even when it is deactivated.
        }

        public class SetWindowPosFlags
        {
            public const int SWP_ASYNCWINDOWPOS = 0x4000;
            //If the calling thread and the thread that owns the window are attached to different input queues, the system posts the request to the thread that owns the window.This prevents the calling thread from blocking its execution while other threads process the request.

            public const int SWP_DEFERERASE = 0x2000;
            //Prevents generation of the WM_SYNCPAINT message.

            public const int SWP_DRAWFRAME = 0x0020;
            // Draws a frame (defined in the window's class description) around the window.

            public const int SWP_FRAMECHANGED = 0x0020;
            //Applies new frame styles set using the SetWindowLong function.Sends a WM_NCCALCSIZE message to the window, even if the window's size is not being changed. If this flag is not specified, WM_NCCALCSIZE is sent only when the window's size is being changed.

            public const int SWP_HIDEWINDOW = 0x0080;
            //Hides the window.

            public const int SWP_NOACTIVATE = 0x0010;
            //Does not activate the window.If this flag is not set, the window is activated and moved to the top of either the topmost or non-topmost group (depending on the setting of the hWndInsertAfter parameter).

            public const int SWP_NOCOPYBITS = 0x0100;
            //Discards the entire contents of the client area.If this flag is not specified, the valid contents of the client area are saved and copied back into the client area after the window is sized or repositioned.

            public const int SWP_NOMOVE = 0x0002;
            //Retains the current position (ignores X and Y parameters).

            public const int SWP_NOOWNERZORDER = 0x0200;
            //Does not change the owner window's position in the Z order.

            public const int SWP_NOREDRAW = 0x0008;
            //Does not redraw changes.If this flag is set, no repainting of any kind occurs. This applies to the client area, the nonclient area (including the title bar and scroll bars), and any part of the parent window uncovered as a result of the window being moved.When this flag is set, the application must explicitly invalidate or redraw any parts of the window and parent window that need redrawing.

            public const int SWP_NOREPOSITION = 0x0200;
            //Same as the SWP_NOOWNERZORDER flag.

            public const int SWP_NOSENDCHANGING = 0x0400;
            //Prevents the window from receiving the WM_WINDOWPOSCHANGING message.

            public const int SWP_NOSIZE = 0x0001;
            //Retains the current size (ignores the cx and cy parameters).

            public const int SWP_NOZORDER = 0x0004;
            //Retains the current Z order (ignores the hWndInsertAfter parameter).

            public const int SWP_SHOWWINDOW = 0x0040;
            //Displays the window.
        }
        public delegate void WinEventDelegate(IntPtr hWinEventHook, uint eventType, IntPtr hwnd, int idObject, int idChild, uint dwEventThread, uint dwmsEventTime);

        [DllImport("user32.dll")]
        public static extern IntPtr SetWinEventHook(uint eventMin, uint eventMax, IntPtr hmodWinEventProc, WinEventDelegate lpfnWinEventProc, uint idProcess, uint idThread, uint dwFlags);

        const uint EVENT_OBJECT_LOCATIONCHANGE = 0x800B;

        [DllImport("user32.dll", SetLastError = false)]
        public static extern bool GetWindowRect(IntPtr hwnd, out RECT lpRect);

        [StructLayout(LayoutKind.Sequential)]
        public struct RECT
        {
            public int Left;        // x position of upper-left corner
            public int Top;         // y position of upper-left corner
            public int Right;       // x position of lower-right corner
            public int Bottom;      // y position of lower-right corner
        }
        private const int WINEVENT_INCONTEXT = 4;
        private const int WINEVENT_OUTOFCONTEXT = 0;
        private const int WINEVENT_SKIPOWNPROCESS = 2;
        private const int WINEVENT_SKIPOWNTHREAD = 1;
        private delegate bool EnumWindowProcess(IntPtr Handle, IntPtr Parameter);

        private static bool EnumWindow(IntPtr Handle, IntPtr Parameter)
        {
            List<IntPtr> target = (List<IntPtr>)GCHandle.FromIntPtr(Parameter).Target;
            if (target == null)
                throw new Exception("GCHandle Target could not be cast as List(Of IntPtr)");
            target.Add(Handle);
            return true;
        }


        static void Main(string[] args)
        {
            string[] KAKAOTALK_TITLE_STRING = { "카카오톡", "Kakaotalk", "カカオトーク" };
            //카카오톡 윈도우 찾기
            var hwnd = IntPtr.Zero;
            foreach (string titleCandidate in KAKAOTALK_TITLE_STRING)
            {
                hwnd = FindWindow(null, titleCandidate);
                if (hwnd != IntPtr.Zero) break;
            }
            //모든 언어에 해당하는 이름의 윈도우가 없는 경우 종료
            if (hwnd == IntPtr.Zero) return;

            //이 경우 한국어/영어/일본어의 윈도우는 보장됨
            Console.WriteLine("Passed");

            HookManager.SubscribeToWindowEvents(hwnd);

            RECT rectKakaoTalkMain = new RECT();
            GetWindowRect(hwnd, out rectKakaoTalkMain);

            var childHwnds = new List<IntPtr>();
            var gcHandle = GCHandle.Alloc(childHwnds);

            try
            {
                EnumChildWindows(hwnd, new EnumWindowProcess(EnumWindow), GCHandle.ToIntPtr(gcHandle));
            }
            finally
            {
                if (gcHandle.IsAllocated)
                    gcHandle.Free();
            }

            var windowClass = new StringBuilder(256);
            foreach (var childHwnd in childHwnds)
            {
                GetClassName(childHwnd, windowClass, windowClass.Capacity);

                //광고 클래스인 경우
                if (windowClass.ToString().Equals("EVA_Window") && GetParent(childHwnd) == hwnd)
                {
                    ShowWindow(childHwnd, 0);
                    SetWindowPos(childHwnd, HwndInsertAfterInt.Bottom, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                }
                //카카오톡 친구 화면인 경우
                if (windowClass.ToString().Equals("EVA_ChildWindow") && GetParent(childHwnd) == hwnd)
                {
                    SetWindowPos(
                        childHwnd,
                        HwndInsertAfterInt.Bottom, 0, 0, rectKakaoTalkMain.Right - rectKakaoTalkMain.Left, (rectKakaoTalkMain.Bottom - rectKakaoTalkMain.Top - 36), SetWindowPosFlags.SWP_NOMOVE);
                }
            }
        }
        public static class HookManager {
            public static void SubscribeToWindowEvents(IntPtr hwnd)
            {
                IntPtr windowEventHook 
                    = SetWinEventHook(EVENT_OBJECT_LOCATIONCHANGE, EVENT_OBJECT_LOCATIONCHANGE, hwnd, HandleWinResizeEvent, 0, 0, WINEVENT_OUTOFCONTEXT | WINEVENT_SKIPOWNTHREAD);
                if (windowEventHook == IntPtr.Zero)
                {
                    Console.WriteLine("event attach failed");
                }
            }
        }
        private static void HandleWinResizeEvent(IntPtr hWinEventHook, uint eventType, IntPtr hwnd, int idObject, int idChild, uint dwEventThread, uint dwmsEventTime)
        {
            RECT rectKakaoTalkMain = new RECT();
            GetWindowRect(hwnd, out rectKakaoTalkMain);

            var childHwnds = new List<IntPtr>();
            var gcHandle = GCHandle.Alloc(childHwnds);

            try
            {
                EnumChildWindows(hwnd, new EnumWindowProcess(EnumWindow), GCHandle.ToIntPtr(gcHandle));
            }
            finally
            {
                if (gcHandle.IsAllocated)
                    gcHandle.Free();
            }

            var windowClass = new StringBuilder(256);
            foreach (var childHwnd in childHwnds)
            {
                GetClassName(childHwnd, windowClass, windowClass.Capacity);

                //광고 클래스인 경우
                if (windowClass.ToString().Equals("EVA_Window") && GetParent(childHwnd) == hwnd)
                {
                    ShowWindow(childHwnd, 0);
                    SetWindowPos(childHwnd, HwndInsertAfterInt.Bottom, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                }
                //카카오톡 친구 화면인 경우
                if (windowClass.ToString().Equals("EVA_ChildWindow") && GetParent(childHwnd) == hwnd)
                {
                    SetWindowPos(
                        childHwnd,
                        HwndInsertAfterInt.Bottom, 0, 0, rectKakaoTalkMain.Right - rectKakaoTalkMain.Left, (rectKakaoTalkMain.Bottom - rectKakaoTalkMain.Top - 36), SetWindowPosFlags.SWP_NOMOVE);
                }
            }
        }
    }
}
