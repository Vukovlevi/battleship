using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;
using Client.Core;

namespace Client
{
    /// <summary>
    /// Interaction logic for GamePage.xaml
    /// </summary>
    public partial class GamePage : Page
    {
        Tcp tcp;
        string username;
        string enemyUsername;
        public GamePage()
        {
            InitializeComponent();
        }

        internal void SetTcp(Tcp tcp)
        {
            this.tcp = tcp;
        }

        public void SetUsernames(string username, string enemyUsername)
        {
            this.username = username;
            this.enemyUsername = enemyUsername;

            Dispatcher.Invoke(() =>
            {
                EnemyUsernameLabel.Content = $"Az ellenfeled: {this.enemyUsername}";
            });
        }
    }
}
