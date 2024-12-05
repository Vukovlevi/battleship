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

		private bool _isCodeRoom;
		public bool IsCodeRoom
		{
			get { return _isCodeRoom; }
			set
			{
				_isCodeRoom = value;

				if (value) GlobalData.Instance.GameCode.Visibility = Visibility.Visible;
				else GlobalData.Instance.GameCode.Visibility = Visibility.Collapsed;

				OnPropertyChanged();
			}
		}

		private string _joinButtonText;

		public string JoinButtonText
		{
			get { return _joinButtonText; }
			set
			{
				_joinButtonText = value;
				OnPropertyChanged();
			}
		}


		private string _gameCode;
		public string GameCode
		{
			get { return _gameCode; }
			set
			{
				_gameCode = value;
				OnPropertyChanged();

				if (value == "") JoinButtonText = "Meccskeresés indítása";
				else JoinButtonText = "Csatlakozás a játékhoz";
			}
		}

		public RelayCommand SetUsername { get; set; }

		public bool IsSearching { get; set; } = false;

        public LoginViewModel()
        {
			_isCodeRoom = false;
            JoinButtonText = "Meccskeresés indítása";
			GameCode = "";

            SetUsername = new RelayCommand(o =>
			{
				if (IsSearching) return;

				if (String.IsNullOrEmpty(Username) || String.IsNullOrWhiteSpace(Username))
				{
					MMState = "Érvénytelen felhasználónév";
					MessageBox.Show("Adj meg egy felhasználónevet!");
					return;
				}

                if (Username.Length > 255)
                {
                    MessageBox.Show("A név túl hosszú!");
                    return;
                }

                GlobalData.Instance.Username = Username;
                GameState.state = State.WaitingForMatch;

                if (GameCode == "")
				{
                    MMState = "Meccs keresése folyamatban...";

                    TcpCommand command = new TcpCommand(CommandType.JoinRequest, ASCIIEncoding.ASCII.GetBytes(Username));
                    GlobalData.Instance.Tcp.Send(command.EncodeToBytes());

                    IsSearching = true;
                } else
				{
					MMState = "Csatlakozás a játékhoz";

					var username = ASCIIEncoding.ASCII.GetBytes(Username);
					byte[] data = { (byte)username.Length };
					data = data.Concat(username).Concat(ASCIIEncoding.ASCII.GetBytes(GameCode)).ToArray();

					TcpCommand command = new TcpCommand(CommandType.CodeJoin, data);
					GlobalData.Instance.Tcp.Send(command.EncodeToBytes());

					IsSearching = true;
				}
            });

			MMState = "";
        }

		public void Clear()
		{
			GameState.state = State.SetUsername;
			Username = "";
			MMState = "";
			IsSearching = false;
			IsCodeRoom = false;
			GameCode = "";
		}

		public void DuplicateUsername()
		{
            GlobalData.Instance.MainVM.SetCurrentView(GlobalData.Instance.LoginVM);
            GameState.state = State.SetUsername;
            Username = "";
            MMState = "Foglalt felhasználónév";
            MessageBox.Show($"A {GlobalData.Instance.Username} felhasználónév már foglalt.\nVálassz másikat!");
        }
    }
}
