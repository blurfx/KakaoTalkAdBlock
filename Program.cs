using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;

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
            public const int SWP_NOMOVE = 0x0002;
            //Retains the current position (ignores X and Y parameters).
        }
        
        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);


        [DllImport("user32.dll", SetLastError = false)]
        public static extern bool GetWindowRect(IntPtr hwnd, out RECT lpRect);

        [DllImport("user32.dll")]
        static extern bool UpdateWindow(IntPtr hWnd);

        [StructLayout(LayoutKind.Sequential)]
        public struct RECT
        {
            public int Left;        // x position of upper-left corner
            public int Top;         // y position of upper-left corner
            public int Right;       // x position of lower-right corner
            public int Bottom;      // y position of lower-right corner
        }
               

        private delegate bool EnumWindowProcess(IntPtr Handle, IntPtr Parameter);

        private static bool EnumWindow(IntPtr Handle, IntPtr Parameter)
        {
            List<IntPtr> target = (List<IntPtr>)GCHandle.FromIntPtr(Parameter).Target;
            if (target == null)
                throw new Exception("GCHandle Target could not be cast as List(Of IntPtr)");
            target.Add(Handle);
            return true;
        }

        static IntPtr hwnd;
        static void Main(string[] args)
        {
            string[] KAKAOTALK_TITLE_STRING = { "카카오톡", "Kakaotalk", "カカオトーク" };
            //카카오톡 윈도우 찾기
            hwnd = IntPtr.Zero;
            foreach (string titleCandidate in KAKAOTALK_TITLE_STRING)
            {
                hwnd = FindWindow(null, titleCandidate);
                if (hwnd != IntPtr.Zero) break;
            }
            //모든 언어에 해당하는 이름의 윈도우가 없는 경우 종료
            if (hwnd == IntPtr.Zero) return;

            //이 경우 한국어/영어/일본어의 윈도우는 보장됨
            Console.WriteLine("Passed");


            Thread runnerThread = new Thread(new ThreadStart(Adremove));
            runnerThread.Start();

        }


        private static void Adremove()
        {
            var childHwnds = new List<IntPtr>();
            var gcHandle = GCHandle.Alloc(childHwnds);
            try
            {
                EnumChildWindows(hwnd, new EnumWindowProcess(EnumWindow), GCHandle.ToIntPtr(gcHandle));
            }
            finally
            {
                if (gcHandle.IsAllocated) gcHandle.Free();
            }

            var windowClass = new StringBuilder(256);
            var windowCaption = new StringBuilder(256);
            while (true)
            {
                //카카오톡의 윈도우 크기를 가져옴
                RECT rectKakaoTalkMain = new RECT();
                GetWindowRect(hwnd, out rectKakaoTalkMain);
                //모든 자식 윈도우에 대해서
                foreach (var childHwnd in childHwnds)
                {
                    GetClassName(childHwnd, windowClass, windowClass.Capacity);
                    GetWindowText(childHwnd, windowCaption, windowCaption.Capacity);
                    //광고 클래스인 경우
                    if (windowClass.ToString().Equals("EVA_Window") && GetParent(childHwnd) == hwnd)
                    {
                        ShowWindow(childHwnd, 0);
                        SetWindowPos(childHwnd, HwndInsertAfterInt.Bottom, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                    }
                    //카카오톡 친구 화면인 경우
                    if (windowClass.ToString().Equals("EVA_ChildWindow") && GetParent(childHwnd) == hwnd && windowCaption.ToString().StartsWith("OnlineMainView"))
                    {
                        var width = rectKakaoTalkMain.Right - rectKakaoTalkMain.Left;
                        var height = (rectKakaoTalkMain.Bottom - rectKakaoTalkMain.Top)-38;
                        UpdateWindow(hwnd);
                        SetWindowPos(childHwnd,IntPtr.Zero,0,0,width, height, SetWindowPosFlags.SWP_NOMOVE);
                    }
                }
                Thread.Sleep(200);
            }
        }

    }
}
