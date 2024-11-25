using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Client.Core;
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

        public LoginViewModel LoginVM { get; set; }

        public MainViewModel()
        {
			LoginVM = new LoginViewModel();
			CurrentView = LoginVM;
        }
    }
}
