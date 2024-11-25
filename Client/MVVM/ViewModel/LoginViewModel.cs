using Client.Core;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

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

        public RelayCommand SetUsername { get; set; }

        public LoginViewModel()
        {
			SetUsername = new RelayCommand(o =>
			{
				MessageBox.Show(Username);
			});
        }
    }
}
