using Client.Core;
using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;

namespace Client.MVVM.ViewModel
{
    internal class LoginViewModel : ObservableObject
    {
		private string _username;
		public string Username
		{
			get { return _username; }
			set 
			{ 
				_username = value;
				OnPropertyChanged();
			}
		}

		private string _mmState;

		public string MMState
		{
			get { return _mmState; }
			set
			{
				_mmState = value;
				OnPropertyChanged();
			}
		}

		public RelayCommand SetUsername { get; set; }

        public LoginViewModel()
        {
			SetUsername = new RelayCommand(o =>
			{
				if (String.IsNullOrEmpty(Username) || String.IsNullOrWhiteSpace(Username))
				{
					MMState = "Érvénytelen felhasználónév";
					MessageBox.Show("Adj meg egy felhasználónevet!");
					return;
				}

				GlobalData.Instance.Username = Username;
				MMState = "Meccs keresése folyamatban...";

				GameState.state = State.WaitingForMatch;
				TcpCommand command = new TcpCommand(CommandType.JoinRequest, ASCIIEncoding.ASCII.GetBytes(Username));
				GlobalData.Instance.Tcp.Send(command.EncodeToBytes());
			});

			MMState = "";
        }

		public void Clear()
		{
			GameState.state = State.SetUsername;
			Username = "";
			MMState = "";
		}

		public void DuplicateUsername()
		{
			Application.Current.Dispatcher.Invoke(() =>
			{
                GlobalData.Instance.MainVM.SetCurrentView(GlobalData.Instance.LoginVM);
                GameState.state = State.SetUsername;
                Username = "";
                MMState = "Foglalt felhasználónév";
                MessageBox.Show($"A {GlobalData.Instance.Username} felhasználónév már foglalt.\nVálassz másikat!");
            }, System.Windows.Threading.DispatcherPriority.Normal);
		}
    }
}
