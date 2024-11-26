using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Client.Core
{
    internal static class Asserter
    {
        static Tcp tcp;
        static string debugFile = "debug.txt";

        public static void SetTcp()
        {
            tcp = GlobalData.Instance.Tcp;
        }

        public static void SetDebugFile(string debugFile)
        {
            Asserter.debugFile = debugFile;
        }

        public static void Assert(bool statement, string message, params string[] data)
        {
            if (data.Length % 2 != 0)
            {
                RunAssert($"[DEBUG - {DateTime.Now}]: Incorrectly formatted data in assert");
            }

            if (!statement)
            {
                string msg = CreateMessage(message, data);
                RunAssert(msg);
            }
        }

        static string CreateMessage(string message, string[] data)
        {
            string msg = $"[DEBUG - {DateTime.Now}]: Assert triggered\n{message}\n";
            for (int i = 0; i < data.Length; i += 2)
            {
                msg += $"\t{data[i]}: {data[i + 1]}\n";
            }
            return msg;
        }

        static void RunAssert(string message)
        {
            StreamWriter writer = new StreamWriter(debugFile, true, Encoding.UTF8);
            writer.WriteLine(message);
            writer.Close();
            MessageBox.Show("An error occured, the program is goint to close!");
            tcp.Close();
            Environment.Exit(1);
        }
    }
}
