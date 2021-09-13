using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Win32;

namespace KakaoTalkAdBlock
{
    class Program
    {
        #region WinAPI
        [DllImport("user32.dll")]
        static extern int GetClassName(IntPtr hWnd, StringBuilder lpClassName, int nMaxCount);

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        static extern bool EnumThreadWindows(uint dwThreadId, EnumThreadDelegate lpfn, IntPtr lParam);

        [DllImport("user32.dll")]
        static extern bool EnumChildWindows(IntPtr WindowHandle, EnumWindowProcess Callback, IntPtr lParam);

        [DllImport("user32.dll")]
        static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);

        [DllImport("user32.dll")]
        static extern IntPtr FindWindowEx(IntPtr hwndParent, IntPtr ihwndChildAfter, string lpszClass, string lpszWindow);

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

        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = false)]
        static extern IntPtr SendMessage(IntPtr hWnd, UInt32 Msg, IntPtr wParam, IntPtr lParam);
        public delegate bool EnumThreadDelegate (IntPtr hwnd, IntPtr lParam);

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
        #endregion

        #region Global Variables

        static string APP_NAME = "KakaoTalkAdBlock";

        static volatile List<IntPtr> hwnd = new List<IntPtr>();
        static Container container = new Container();

        static Thread watcherThread = new Thread(new ThreadStart(watchProcess));
        static Thread runnerThread = new Thread(new ThreadStart(removeAd));

        static readonly object hwndLock = new object();

        const int UPDATE_RATE = 100;

        const int LAYOUT_SHADOW_PADDING = 2;

        const int MAINVIEW_PADDING = 31;

        static uint WM_CLOSE = 0x10;
        #endregion

        static ContextMenuStrip buildContextMenu()
        {
            var contextMenu = new ContextMenuStrip();
            var versionItem = new ToolStripMenuItem();
            var exitItem = new ToolStripMenuItem();
            var startupItem = new ToolStripMenuItem();

            // version
            versionItem.Text = "v1.3.1";
            versionItem.Enabled = false;

            // if startup is enabled, set startup menu checked
            {
                var regStartup = Registry.CurrentUser.OpenSubKey("SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run", true);
                var regStartupValue = regStartup.GetValue(APP_NAME, false);
                if (!regStartupValue.Equals(false))
                {
                    startupItem.Checked = true;
                }
            }

            // run on startup menu
            startupItem.Text = "윈도우 시작 시 자동 실행";
            startupItem.Click += new EventHandler(delegate (object sender, EventArgs e)
            {
                var regStartup = Registry.CurrentUser.OpenSubKey("SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run", true);
                if (startupItem.Checked)
                {
                    regStartup.DeleteValue(APP_NAME, false);
                    startupItem.Checked = false;
                }
                else
                {
                    regStartup.SetValue(APP_NAME, Application.ExecutablePath);
                    startupItem.Checked = true;
                }
            });

            // exit menu
            exitItem.Text = "종료(&x)";
            exitItem.Click += new EventHandler(delegate (object sender, EventArgs e)
            {
                Environment.Exit(0);
            });

            contextMenu.Items.Add(versionItem);
            contextMenu.Items.Add(startupItem);
            contextMenu.Items.Add("-");
            contextMenu.Items.Add(exitItem);

            return contextMenu;
        }

        static void Main()
        {
            var mutex = new Mutex(true, APP_NAME, out bool isNotDuplicated);

            if (!isNotDuplicated)
            {
                MessageBox.Show("이미 실행 중입니다.");
                return;
            }

            // build trayicon
            new NotifyIcon(container)
            {
                Visible = true,
                Icon = Properties.Resources.icon,
                ContextMenuStrip = buildContextMenu()
            };

            watcherThread.Start();
            runnerThread.Start();
            Application.Run();
            mutex.ReleaseMutex();
        }

        static void watchProcess()
        {
            while (true)
            {
                // hwnd must not be changed while removing ad
                lock (hwndLock)
                {
                    hwnd.Clear();

                    var processes = Process.GetProcessesByName("kakaotalk");
                    foreach (Process proc in processes)
                    {
                        foreach(ProcessThread thread in proc.Threads)
                        {
                            EnumThreadWindows(Convert.ToUInt32(thread.Id), (twnd, _) =>
                            {
                                hwnd.Add(twnd);
                                return true;
                            }, IntPtr.Zero);
                        }
                    }
                }

                Thread.Sleep(UPDATE_RATE);
            }
        }

        static void removeAd()
        {
            var localHwnd = new List<IntPtr>();
            var childHwnds = new List<IntPtr>();
            var windowClass = new StringBuilder(256);
            var windowCaption = new StringBuilder(256);
            var windowParentCaption = new StringBuilder(256);

            while (true)
            {
                // hwnd must not be changed while removing ad
                lock (hwndLock)
                {
                    foreach (IntPtr wnd in hwnd)
                    {

                        if (wnd == IntPtr.Zero)
                        {
                            continue;
                        }

                        childHwnds.Clear();
                        var gcHandle = GCHandle.Alloc(childHwnds);

                        // get handles from child windows
                        try
                        {
                            EnumChildWindows(wnd, new EnumWindowProcess(EnumWindow), GCHandle.ToIntPtr(gcHandle));
                        }
                        finally
                        {
                            if (gcHandle.IsAllocated) gcHandle.Free();
                        }

                        // get rect of kakaotalk
                        RECT rectKakaoTalk = new RECT();
                        GetWindowRect(wnd, out rectKakaoTalk);

                        // iterate all child windows of kakaotalk
                        foreach (var childHwnd in childHwnds)
                        {
                            GetClassName(childHwnd, windowClass, windowClass.Capacity);
                            GetWindowText(childHwnd, windowCaption, windowCaption.Capacity);

                            HideMainWindowAd(windowClass, windowParentCaption, wnd, childHwnd);
                            HideMainViewAdArea(windowCaption, wnd, rectKakaoTalk, childHwnd);
                            HideLockScreenAdArea(windowCaption, wnd, rectKakaoTalk, childHwnd);
                        }
                    }
                    HidePopupAd();
                }
                Thread.Sleep(UPDATE_RATE);
            }
        }

        private static void HidePopupAd()
        {
            var popUpHwnd = IntPtr.Zero;
            while ((popUpHwnd = FindWindowEx(IntPtr.Zero, popUpHwnd, null, "")) != IntPtr.Zero)
            {
                // popup ad does not have any parent
                if (GetParent(popUpHwnd) != IntPtr.Zero) continue;

                // get class name of blank title
                var classNameSb = new StringBuilder(256);
                GetClassName(popUpHwnd, classNameSb, classNameSb.Capacity);
                string className = classNameSb.ToString();

                if (!className.Contains("RichPopWnd")) continue;

                // get rect of popup ad
                GetWindowRect(popUpHwnd, out RECT rectPopup);
                var width = rectPopup.Right - rectPopup.Left;
                var height = rectPopup.Bottom - rectPopup.Top;

                if (width.Equals(300) && height.Equals(150))
                {
                    SendMessage(popUpHwnd, WM_CLOSE, IntPtr.Zero, IntPtr.Zero);
                }
            }
        }

        private static void HideMainWindowAd(StringBuilder windowClass, StringBuilder windowParentCaption, IntPtr wnd, IntPtr childHwnd)
        {
            if (windowClass.ToString().Equals("BannerAdWnd") || windowClass.ToString().Equals("EVA_Window"))
            {
                GetWindowText(GetParent(childHwnd), windowParentCaption, windowParentCaption.Capacity);

                if (windowParentCaption.ToString().StartsWith("LockModeView"))
                {
                    ShowWindow(childHwnd, 0);
                    SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                }
            }
        }

        private static void HideLockScreenAdArea(StringBuilder windowCaption, IntPtr wnd, RECT rectKakaoTalk, IntPtr childHwnd)
        {
            if (windowCaption.ToString().StartsWith("LockModeView") && GetParent(childHwnd) == wnd)
            {
                var width = rectKakaoTalk.Right - rectKakaoTalk.Left - LAYOUT_SHADOW_PADDING;
                var height = rectKakaoTalk.Bottom - rectKakaoTalk.Top;
                UpdateWindow(childHwnd);
                SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
            }
        }

        private static void HideMainViewAdArea(StringBuilder windowCaption, IntPtr wnd, RECT rectKakaoTalk, IntPtr childHwnd)
        {
            if (windowCaption.ToString().StartsWith("OnlineMainView") && GetParent(childHwnd) == wnd)
            {
                var width = rectKakaoTalk.Right - rectKakaoTalk.Left - LAYOUT_SHADOW_PADDING;
                var height = rectKakaoTalk.Bottom - rectKakaoTalk.Top - MAINVIEW_PADDING;
                if (height < 1)
                {
                    return;
                }
                UpdateWindow(childHwnd);
                SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
            }
        }
    }
}
