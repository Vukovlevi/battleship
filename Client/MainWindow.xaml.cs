using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading;
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

namespace Client
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        Tcp tcp;
        public static string username;
        public MainWindow()
        {
            InitializeComponent();
            StartGame();
            GameState.SetWindow(this);
            this.Closing += CloseWindow;
        }

        void CloseWindow(object sender, CancelEventArgs e)
        {
            tcp.Close();
        }


        void StartGame()
        {
            tcp = new Tcp();
            Asserter.SetTcp(tcp);
            tcp.Connect();
            Thread listeningThread = new Thread(tcp.Listen);
            listeningThread.Start();
            GameState.state = State.SetUsername;
        }

        private void UsernameInputKeydown(object sender, KeyEventArgs e)
        {
            if (e.Key != Key.Enter) return;
            StartMatchMaking();
        }

        private void StartMatchmakingButton(object sender, RoutedEventArgs e)
        {
            StartMatchMaking();
        }

        void StartMatchMaking()
        {
            username = usernameInput.Text;

            if (username.Length == 0)
            {
                MessageBox.Show("Állítson be felhasználónevet!");
                return;
            }

            HideSetUsername();

            TcpCommand command = new TcpCommand(Convert.ToByte(CommandType.JoinRequest), ASCIIEncoding.ASCII.GetBytes(username));
            tcp.Send(command.EncodeToBytes());
        }

        public void HideSetUsername()
        {
            GameState.state = State.WaitingForMatch;
            usernameLabel.Visibility = Visibility.Hidden;
            usernameInput.Visibility = Visibility.Hidden;
            usernameSetButton.Visibility = Visibility.Hidden;
        }

        public void ShowSetUsername()
        {
            GameState.state = State.SetUsername;
            usernameLabel.Visibility = Visibility.Visible;
            usernameInput.Visibility = Visibility.Visible;
            usernameSetButton.Visibility = Visibility.Visible;
        }
    }
}
