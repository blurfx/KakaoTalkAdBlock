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


        const int UPDATE_RATE = 100;

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

        #region Constants
        static string[] KAKAOTALK_TITLE_STRING = { "카카오톡", "Kakaotalk", "カカオトーク" };

        static string APP_NAME = "KakaoTalkAdBlock";
        #endregion

        #region Global Variables
        static IntPtr hwnd = IntPtr.Zero;
        static Container container = new Container();

        static Thread watcherThread = new Thread(new ThreadStart(watchProcess));
        static Thread runnerThread = new Thread(new ThreadStart(removeAd));

        static bool isKakaotalkRunning = false;
        #endregion

        static ContextMenuStrip buildContextMenu()
        {
            var contextMenu = new ContextMenuStrip();
            var versionItem = new ToolStripMenuItem();
            var exitItem = new ToolStripMenuItem();
            var startupItem = new ToolStripMenuItem();

            // version
            versionItem.Text = "v0.0.6";
            versionItem.Enabled = false;

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
                MessageBox.Show("이미 실행중입니다");
                return;
            }

            // build trayicon
            NotifyIcon tray = new NotifyIcon(container);
            tray.Visible = true;
            tray.Icon = Properties.Resources.icon;
            tray.ContextMenuStrip = buildContextMenu();

            watcherThread.Start();
            runnerThread.Start();
            Application.Run();
            mutex.ReleaseMutex();
        }


    static void watchProcess()
        {
            while (true)
            {
                while (!isKakaotalkRunning)
                {
                    System.Diagnostics.Debug.WriteLine("watching");
                    hwnd = IntPtr.Zero;

                    // find kakaotalk window
                    foreach (string titleCandidate in KAKAOTALK_TITLE_STRING)
                    {
                        hwnd = FindWindow(null, titleCandidate);
                        if (hwnd != IntPtr.Zero) break;
                    }

                    if (hwnd != IntPtr.Zero)
                    {
                        isKakaotalkRunning = true;
                    }
                    Thread.Sleep(UPDATE_RATE);
                }
                Thread.Sleep(UPDATE_RATE);
            }
        }

        static void removeAd()
        {
            
            var childHwnds = new List<IntPtr>();
            var windowClass = new StringBuilder(256);
            var windowCaption = new StringBuilder(256);
            var windowParentCaption = new StringBuilder(256);

            while (true)
            {
                while (isKakaotalkRunning)
                {
                    childHwnds.Clear();
                    var gcHandle = GCHandle.Alloc(childHwnds);

                    // get handles from child windows
                    try
                    {
                        EnumChildWindows(hwnd, new EnumWindowProcess(EnumWindow), GCHandle.ToIntPtr(gcHandle));
                    }
                    finally
                    {
                        if (gcHandle.IsAllocated) gcHandle.Free();
                        if(childHwnds.Count == 0)
                        {
                            isKakaotalkRunning = false;
                        }
                    }

                    // get rect of kakaotalk
                    RECT rectKakaoTalk = new RECT();
                    GetWindowRect(hwnd, out rectKakaoTalk);
                    // iterate all child windows of kakaotalk
                    foreach (var childHwnd in childHwnds)
                    {
                        GetClassName(childHwnd, windowClass, windowClass.Capacity);
                        GetWindowText(childHwnd, windowCaption, windowCaption.Capacity);

                        // hide ad
                        if (windowClass.ToString().Equals("EVA_Window") )
                        {
                            GetWindowText(GetParent(childHwnd), windowParentCaption, windowParentCaption.Capacity);

                            if(GetParent(childHwnd) == hwnd|| windowParentCaption.ToString().StartsWith("LockModeView")) { 
                                ShowWindow(childHwnd, 0);
                                SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, 0, 0, SetWindowPosFlags.SWP_NOMOVE);
                            }
                        }

                        // remove white area
                        if (windowCaption.ToString().StartsWith("OnlineMainView") && GetParent(childHwnd) == hwnd)
                        {
                            var width = rectKakaoTalk.Right - rectKakaoTalk.Left;
                            var height = (rectKakaoTalk.Bottom - rectKakaoTalk.Top) - 38; // 38; there might be dragon. don't touch it.
                            UpdateWindow(hwnd);
                            SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
                        }

                        if (windowCaption.ToString().StartsWith("LockModeView") && GetParent(childHwnd) == hwnd){
                            var width = rectKakaoTalk.Right - rectKakaoTalk.Left;
                            var height = (rectKakaoTalk.Bottom - rectKakaoTalk.Top); // 38; there might be dragon. don't touch it.
                            UpdateWindow(hwnd);
                            SetWindowPos(childHwnd, IntPtr.Zero, 0, 0, width, height, SetWindowPosFlags.SWP_NOMOVE);
                        }
                    }
                    Thread.Sleep(UPDATE_RATE);
                }
                Thread.Sleep(UPDATE_RATE);
            }
        }
    }
}
