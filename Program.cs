using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Windows.Forms;

namespace KakaoTalkAdBlock
{
    class Program
    {
        [DllImport("user32.dll")]
        static extern int GetClassName(IntPtr hWnd, StringBuilder lpClassName, int nMaxCount);

        [DllImport("user32.dll")]
        static extern bool EnumChildWindows(IntPtr WindowHandle, EnumWindowProcess Callback, IntPtr lParam);

        [DllImport("user32.dll")]
        static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);

        [DllImport("user32.dll")]
        static extern IntPtr FindWindow(string lpClassName, string lpWindowName);

        [DllImport("user32.dll")]
        static extern IntPtr GetParent(IntPtr hWnd);

        [DllImport("user32.dll", EntryPoint = "SetWindowPos", SetLastError = false)]
        static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, int uFlags);

        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);

        [DllImport("user32.dll", SetLastError = false)]
        static extern bool GetWindowRect(IntPtr hwnd, out RECT lpRect);

        [DllImport("user32.dll")]
        static extern bool UpdateWindow(IntPtr hWnd);

        static class SetWindowPosFlags
        {
            public const int SWP_NOMOVE = 0x0002;
        }

        [StructLayout(LayoutKind.Sequential)]
        struct RECT
        {
            public int Left;
            public int Top;
            public int Right;
            public int Bottom;
        }

        delegate bool EnumWindowProcess(IntPtr Handle, IntPtr Parameter);

        static bool EnumWindow(IntPtr Handle, IntPtr Parameter)
        {
            List<IntPtr> target = (List<IntPtr>)GCHandle.FromIntPtr(Parameter).Target;
            if (target == null)
                throw new Exception("GCHandle Target could not be cast as List(Of IntPtr)");
            target.Add(Handle);
            return true;
        }

        static IntPtr hwnd = IntPtr.Zero;
        static Container container = new Container();

        static ContextMenuStrip buildContextMenu()
        {
            var contextMenu = new ContextMenuStrip();
            var exitItem = new ToolStripMenuItem();
            exitItem.Text = "종료(&x)";
            exitItem.Click += new EventHandler(delegate (object sender, EventArgs e)
            {
                Environment.Exit(0);
            });

            contextMenu.Items.Add(exitItem);

            return contextMenu;
        }

        static void Main(string[] args)
        {
            

            string[] KAKAOTALK_TITLE_STRING = { "카카오톡", "Kakaotalk", "カカオトーク" };
            
            // find kakaotalk window
            foreach (string titleCandidate in KAKAOTALK_TITLE_STRING)
            {
                hwnd = FindWindow(null, titleCandidate);
                if (hwnd != IntPtr.Zero) break;
            }


            //// build trayicon
            NotifyIcon tray = new NotifyIcon(container);
            tray.Visible = true;
            tray.Icon = Properties.Resources.icon;
            tray.ContextMenuStrip = buildContextMenu();
            
            // exit program if kakaotalk is not found
            if (hwnd == IntPtr.Zero)
            {
                MessageBox.Show("카카오톡이 실행중이지 않은 것 같습니다.");
                return;
            }

            Thread runnerThread = new Thread(new ThreadStart(removeAd));
            runnerThread.Start();
            Application.Run();
        }

        static void removeAd()
        {
            // get handles from child windows
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
                // get rect of kakaotalk
                RECT rectKakaoTalk = new RECT();
                GetWindowRect(hwnd, out rectKakaoTalk);
                // iterate all child windows of kakaotalk
                foreach (var childHwnd in childHwnds)
                {
                    GetClassName(childHwnd, windowClass, windowClass.Capacity);
                    GetWindowText(childHwnd, windowCaption, windowCaption.Capacity);
                    
                    // hide ad
                    if (windowClass.ToString().Equals("EVA_Window") && GetParent(childHwnd) == hwnd)
                    {
                        ShowWindow(childHwnd, 0);
                        SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                    }
                    
                    // remove white area
                    if (windowCaption.ToString().StartsWith("OnlineMainView") && GetParent(childHwnd) == hwnd)
                    {
                        var width = rectKakaoTalk.Right - rectKakaoTalk.Left;
                        var height = (rectKakaoTalk.Bottom - rectKakaoTalk.Top) - 38; // 38; there might be dragon. don't touch it.
                        UpdateWindow(hwnd);
                        SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
                    }
                }
                Thread.Sleep(1000);
            }
        }
    }
}
