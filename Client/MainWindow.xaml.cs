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
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        Tcp tcp;
        public MainWindow()
        {
            InitializeComponent();
            tcp = new Tcp();
            tcp.Connect();
            MessageBox.Show("client connected");
            //StartGame();
        }

        void StartGame()
        {
            tcp.Listen();
        }

        private void TestConnect(object sender, RoutedEventArgs e)
        {
            TcpCommand command = new TcpCommand(1, ASCIIEncoding.ASCII.GetBytes("vukovlevi"));
            tcp.Send(command.EncodeToBytes());
        }

        private void TestMessageBox(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("mukszik");
        }
    }
}
