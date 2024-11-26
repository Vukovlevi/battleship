using Client.Core;
using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Client.MVVM.ViewModel
{
    class GameBoardViewModel : ObservableObject
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

		private string _enemyUsername;
		public string EnemyUsername
		{
			get { return _enemyUsername; }
			set
			{
				_enemyUsername = value;
				OnPropertyChanged();
			}
		}

		public void SetUsernames(string username, string enemyUsername)
		{
			GlobalData.Instance.MainVM.SetCurrentView(GlobalData.Instance.LoginVM);
			GameState.state = State.PlacingShips;
			Username = username;
			EnemyUsername = enemyUsername;
		}
	}
}
