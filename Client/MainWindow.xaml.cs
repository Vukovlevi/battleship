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
        public SetUsernamePage setUsernamePage;
        public GamePage gamePage;
        Thread listeningThread;
        
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


        public void StartGame()
        {
            tcp = new Tcp();
            Asserter.SetTcp(tcp);
            GameState.SetTcp(tcp);
            tcp.Connect();
            listeningThread = new Thread(tcp.Listen);
            listeningThread.SetApartmentState(ApartmentState.STA);
            listeningThread.Start();
            GameState.state = State.SetUsername;

            setUsernamePage = new SetUsernamePage();
            setUsernamePage.SetTcp(tcp);
            Frame.Content = setUsernamePage;

            gamePage = new GamePage();
            gamePage.SetTcp(tcp);
        }
    }
}
