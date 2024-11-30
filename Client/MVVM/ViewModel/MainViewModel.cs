using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using System.Windows;
using Client.Core;
using Client.MVVM.Model;
using Client.MVVM.View;

namespace Client.MVVM.ViewModel
{
    internal class MainViewModel : ObservableObject
    {
		private object _currentView;
		public object CurrentView
		{
			get { return _currentView; }
			set 
			{
				_currentView = value;
				OnPropertyChanged();
			}
		}
        public RelayCommand CloseCommand { get; set; }
        public RelayCommand MinimizeCommand { get; set; }
        public RelayCommand MaximizeCommand { get; set; }
        public RelayCommand MoveWindowCommand { get; set; }

        public LoginViewModel LoginVM { get; set; }
        public GameBoardViewModel GameBoardVM { get; set; }

        public MainViewModel()
        {
            StartGame();

            GlobalData.Instance.MainVM = this;

			LoginVM = new LoginViewModel();
            CurrentView = LoginVM;
            GlobalData.Instance.LoginVM = LoginVM;

			GameBoardVM = new GameBoardViewModel();
            //CurrentView = GameBoardVM;
            GlobalData.Instance.GameBoardVM = GameBoardVM;

            Application.Current.MainWindow.MaxHeight = SystemParameters.MaximizedPrimaryScreenHeight;

            CloseCommand = new RelayCommand(o =>
            {
                GlobalData.Instance.Tcp.Close();
                Application.Current.Shutdown();
            });

            MinimizeCommand = new RelayCommand(o =>
            {
                Application.Current.MainWindow.WindowState = WindowState.Minimized;
            });

            MaximizeCommand = new RelayCommand(o =>
            {
                Application.Current.MainWindow.WindowState = Application.Current.MainWindow.WindowState == WindowState.Maximized ? WindowState.Normal : WindowState.Maximized;
            });

            MoveWindowCommand = new RelayCommand(o =>
            {
                Application.Current.MainWindow.DragMove();
            });
        }

        public void StartGame()
        {
            GlobalData.Instance.Tcp = new Tcp();
            Tcp tcp = GlobalData.Instance.Tcp;
            Asserter.SetTcp();
            GameState.SetTcp();
            GameState.SetShips();
            tcp.Connect();
            Thread listeningThread = new Thread(tcp.Listen);
            listeningThread.SetApartmentState(ApartmentState.STA);
            listeningThread.Start();
            GameState.state = State.SetUsername;
            GameState.orientation = Core.Orientation.Vertical;
        }

        public void RestartGame()
        {
            GlobalData.Instance.Tcp.Close();
            StartGame();
            CurrentView = LoginVM;
            LoginVM.Clear();
        }

        public void SetCurrentView(object view)
        {
            CurrentView = view;
        }
    }
}
