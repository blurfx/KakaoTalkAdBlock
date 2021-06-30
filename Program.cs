using System;
using System.Collections.Generic;
using System.ComponentModel;
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
        static IntPtr popUpHwnd = IntPtr.Zero;
        static Container container = new Container();

        static Thread watcherThread = new Thread(new ThreadStart(watchProcess));
        static Thread runnerThread = new Thread(new ThreadStart(removeAd));

        static readonly object hwndLock = new object();
        static bool hasRemovedPopupAd = false;

        const int UPDATE_RATE = 100;

        static uint WM_CLOSE = 0x10;
        #endregion

        static ContextMenuStrip buildContextMenu()
        {
            var contextMenu = new ContextMenuStrip();
            var versionItem = new ToolStripMenuItem();
            var exitItem = new ToolStripMenuItem();
            var startupItem = new ToolStripMenuItem();

            // version
            versionItem.Text = "v0.0.12";
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

        static void Main(string[] args)
        {
            bool isNotDuplicated = true;
            var mutex = new Mutex(true, APP_NAME, out isNotDuplicated);

            if (!isNotDuplicated)
            {
                MessageBox.Show("이미 실행 중입니다.");
                return;
            }

            // build trayicon
            NotifyIcon tray = new NotifyIcon(container)
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

        static bool hasEVAWindow(IntPtr parentHwnd)
        {
            IntPtr childHwnd = IntPtr.Zero;
            var className = new StringBuilder(256); // Class name has length limit by 256 by WNDCLASSA structure
            while ((childHwnd = FindWindowEx(parentHwnd, childHwnd, null, null)) != IntPtr.Zero)
            {
                GetClassName(childHwnd, className, className.Capacity);
                if (className.ToString().Contains("EVA_Window")) return true;
            }
            return false;
        }

        static void watchProcess()
        {
            while (true)
            {
                System.Diagnostics.Debug.WriteLine("watching");
                List<IntPtr> allHWnd = new List<IntPtr>();
                IntPtr tmpHwnd;

                // hwnd must not be changed while removing ad
                lock (hwndLock)
                {
                    hwnd.Clear();
                    allHWnd.Clear();
                    tmpHwnd = IntPtr.Zero;

                    while ((tmpHwnd = FindWindowEx(IntPtr.Zero, tmpHwnd, null, null)) != IntPtr.Zero)
                    {
                        allHWnd.Add(tmpHwnd);
                    }
                    hwnd.AddRange(allHWnd.FindAll(hasEVAWindow));
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
                System.Diagnostics.Debug.WriteLine("removing");

                // hwnd must not be changed while removing ad
                lock (hwndLock)
                {
                    foreach (IntPtr wnd in hwnd)
                    {
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

                            // hide ad
                            if (windowClass.ToString().Equals("EVA_Window"))
                            {
                                GetWindowText(GetParent(childHwnd), windowParentCaption, windowParentCaption.Capacity);

                                if (GetParent(childHwnd) == wnd || windowParentCaption.ToString().StartsWith("LockModeView"))
                                {
                                    ShowWindow(childHwnd, 0);
                                    SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                                }
                            }

                            // remove white area
                            if (windowCaption.ToString().StartsWith("OnlineMainView") && GetParent(childHwnd) == wnd)
                            {
                                var width = rectKakaoTalk.Right - rectKakaoTalk.Left;
                                var height = (rectKakaoTalk.Bottom - rectKakaoTalk.Top) - 31; // 31; there might be dragon. don't touch it.
                                UpdateWindow(wnd);
                                SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
                            }

                            if (windowCaption.ToString().StartsWith("LockModeView") && GetParent(childHwnd) == wnd)
                            {
                                var width = rectKakaoTalk.Right - rectKakaoTalk.Left;
                                var height = (rectKakaoTalk.Bottom - rectKakaoTalk.Top); // 38; there might be dragon. don't touch it.
                                UpdateWindow(wnd);
                                SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
                            }
                        }
                    }

                    // close popup ad
                    popUpHwnd = IntPtr.Zero;

                    while ((popUpHwnd = FindWindowEx(IntPtr.Zero, popUpHwnd, null, "")) != IntPtr.Zero)
                    {
                        // popup ad does not have any parent
                        if (GetParent(popUpHwnd) != IntPtr.Zero) continue;

                        // get class name of blank title
                        var classNameSb = new StringBuilder(256);
                        GetClassName(popUpHwnd, classNameSb, classNameSb.Capacity);
                        string className = classNameSb.ToString();

                        if (!className.Contains("EVA_Window_Dblclk")) continue;

                        // get rect of popup ad
                        RECT rectPopup = new RECT();
                        GetWindowRect(popUpHwnd, out rectPopup);

                        var width = rectPopup.Right - rectPopup.Left;
                        var height = rectPopup.Bottom - rectPopup.Top;

                        if (width.Equals(300) && height.Equals(150))
                        {
                            SendMessage(popUpHwnd, WM_CLOSE, IntPtr.Zero, IntPtr.Zero);
                        }
                    }
                }
                Thread.Sleep(UPDATE_RATE);
            }
        }
    }
}
