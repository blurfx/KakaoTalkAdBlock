using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text;

namespace KakaoTalkAdBlock
{
    class Program
    {
        [DllImport("user32.dll")]
        public static extern int GetClassName(IntPtr hWnd, StringBuilder lpClassName, int nMaxCount);

        [DllImport("user32.dll")]
        private static extern bool EnumChildWindows(IntPtr WindowHandle, EnumWindowProcess Callback, IntPtr lParam);

        [DllImport("user32.dll")]
        public static extern Boolean ShowWindow(IntPtr hWnd, int nCmdShow);

        [DllImport("user32.dll")]
        static extern IntPtr FindWindow(string lpClassName, string lpWindowName);


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
            var hwnd = FindWindow(null, "카카오톡");

            if(hwnd == IntPtr.Zero)
            {
                hwnd = FindWindow(null, "KakaoTalk");
                if(hwnd == IntPtr.Zero)
                {
                    return;
                }
            }

            if (hwnd == IntPtr.Zero) return;
            Console.WriteLine("Passed");

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

                if (windowClass.ToString().Equals("EVA_Window"))
                {
                    ShowWindow(childHwnd, 0);
                }
            }
        }

    }
}
