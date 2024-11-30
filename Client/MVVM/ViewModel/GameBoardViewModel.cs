using Client.Core;
using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;

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

		private int _enemyRemainingShips;
		public int EnemyRemainingShips
		{
			get { return _enemyRemainingShips; }
			set
			{
				_enemyRemainingShips = value;
				OnPropertyChanged();
			}
		}

		private int _yourRemainingShips;
		public int YourRemainingShips
		{
			get { return _yourRemainingShips; }
			set
			{
				_yourRemainingShips = value;
				OnPropertyChanged();
			}
		}

		private string _orientationStatus;

		public string OrientationStatus
		{
			get { return _orientationStatus; }
			set
			{
				_orientationStatus = value;
				OnPropertyChanged();
			}
		}


		public RelayCommand ChangeOrientation { get; set; }

        public GameBoardViewModel()
        {
			ChangeOrientation = new RelayCommand(o =>
			{
				if (GameState.CurrentShip != null)
				{
					MessageBox.Show("Nem változtathatod meg az irányt, ameddig egy hajó felrakás alatt áll!\n(Ha mégsem akarod felrakni a hajót, kattints rá mégegyszer a felrakás visszavonásához!)");
					return;
				}

                switch (GameState.orientation)
                {
                    case Core.Orientation.Horizontal:
						GameState.orientation = Core.Orientation.Vertical;
						OrientationStatus = "A jelenlegi irány: függőleges";
                        break;
                    case Core.Orientation.Vertical:
						GameState.orientation = Core.Orientation.Horizontal;
						OrientationStatus = "A jelenlegi irány: vízszintes";
                        break;
                    default:
                        Asserter.Assert(false, "orientation can not be other than specified in the enum", "got orientation", GameState.orientation.ToString());
                        break;
                }
            });

			OrientationStatus =  "A jelenlegi irány: függőleges";
        }

        public void SetUsernames(string username, string enemyUsername)
		{
			GlobalData.Instance.MainVM.SetCurrentView(GlobalData.Instance.GameBoardVM);
			GameState.state = State.PlacingShips;
			Username = username;
			EnemyUsername = enemyUsername;
			EnemyRemainingShips = GameState.DefaultShipCount;
			YourRemainingShips = GameState.DefaultShipCount;
		}

		public void PlaceShip(Button cell, Grid grid)
		{
			int y = Grid.GetRow(cell);
			int x = Grid.GetColumn(cell);

			bool success = GameState.CurrentShip.SetRowAndColumn(y, x);
			if (!success)
			{
				return;
			}

			List<Button> toRemoveCells = new List<Button>();
			toRemoveCells.Add(cell);
			switch (GameState.orientation)
			{
				case Core.Orientation.Horizontal:
					for (int i = 1; i < GameState.CurrentShip.Length; i++)
					{
						int shipIndex = GameState.Ships.FindIndex(s => s.StartRow == y && s.StartColumn == x + i);
						if (shipIndex != -1)
						{
							MessageBox.Show("A hajó egy másik hajóba ütközne, ezért nem rakhatod ide!");
							return;
						}

						cell = GetButtonAt(y, x + i, grid);
						if (cell == null)
						{
							MessageBox.Show("A hajó egy másik hajóba ütközne, ezért nem rakhatod ide!");
							return;
						}
						toRemoveCells.Add(cell);
					}
					break;
				case Core.Orientation.Vertical:
					for (int i = 1; i < GameState.CurrentShip.Length; i++)
					{
						int shipIndex = GameState.Ships.FindIndex(s => s.StartRow == y + i && s.StartColumn == x);
						if (shipIndex != -1)
						{
							MessageBox.Show("A hajó egy másik hajóba ütközne, ezért nem rakhatod ide!");
							return;
						}

						cell = GetButtonAt(y + i, x, grid);
						if (cell == null)
						{
							MessageBox.Show("A hajó egy másik hajóba ütközne, ezért nem rakhatod ide!");
							return;
						}
						toRemoveCells.Add(cell);
					}
					break;
				default:
                    Asserter.Assert(false, "orientation can not be other than specified in the enum", "got orientation", GameState.orientation.ToString());
					break;
			}

			foreach (Button currCell in toRemoveCells)
			{
				grid.Children.Remove(currCell);
			}

			Button button = GameState.CurrentShip.GetCell();
            Ship ship = GameState.CurrentShip;
            button.Command = new RelayCommand(o =>
			{
				ship.DeleteShip(grid);
			});
			grid.Children.Add(button);

			GameState.CurrentShip.IsPlaced = true;
			GameState.CurrentShip = null;
		}

		Button? GetButtonAt(int row, int column, Grid grid)
		{
			try
			{
                return (Button)grid.Children.Cast<UIElement>().First(c => Grid.GetRow(c) == row && Grid.GetColumn(c) == column);
            } catch (Exception e)
			{
				return null;
			}
        }
    }
}
