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

namespace Client
{
    /// <summary>
    /// Interaction logic for SetUsernamePage.xaml
    /// </summary>
    public partial class SetUsernamePage : Page
    {
        public static string username;
        Tcp tcp;

        public SetUsernamePage()
        {
            InitializeComponent();
        }

        internal void SetTcp(Tcp tcp)
        {
            this.tcp = tcp;
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

            GameState.state = State.WaitingForMatch;
            TcpCommand command = new TcpCommand(Convert.ToByte(CommandType.JoinRequest), ASCIIEncoding.ASCII.GetBytes(username));
            tcp.Send(command.EncodeToBytes());
        }
    }
}
