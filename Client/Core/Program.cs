using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Velopack;

namespace Client.Core
{
    class Program
    {
        [STAThread]
        public static void Main(string[] args)
        {
            VelopackApp.Build().Run();
            var application = new App();
            application.InitializeComponent();
            application.Run();

            UpdateMyApp();
        }

        private static async Task UpdateMyApp()
        {
            var mgr = new UpdateManager("https://vukovlevi.dev/content/battleship");

            // check for new version
            var newVersion = await mgr.CheckForUpdatesAsync();
            if (newVersion == null)
                return; // no update available

            // download new version
            await mgr.DownloadUpdatesAsync(newVersion);

            // install new version and restart app
            mgr.ApplyUpdatesAndRestart(newVersion);
        }
    }
}
