using Client.Core;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;

namespace Client.MVVM.Model
{
    class Ship
    {
        public int Id { get; private set;}
        public bool IsPlaced { get; set;}
        public int StartRow { get; private set; } = 0;
        public int StartColumn { get; private set; } = 0;
        public Core.Orientation orientation { get; set; }
        public int Length { get; private set; }
        public int Health { get; set; }
        private Button? Cell { get; set; } = null;

        public Ship(int id, int length, Core.Orientation orientation)
        {
            Id = id;
            Length = length;
            Health = length;
            this.orientation = orientation;
        }

        public bool SetRowAndColumn(int row, int column)
        {
            switch (orientation)
            {
                case Core.Orientation.Horizontal:
                    if (column + Length > GlobalData.Instance.BoardSize)
                    {
                        MessageBox.Show("Ide nem fér a hajó!");
                        return false;
                    }
                    break;
                case Core.Orientation.Vertical:
                    if (row + Length > GlobalData.Instance.BoardSize)
                    {
                        MessageBox.Show("Ide nem fér a hajó!");
                        return false;
                    }
                    break;
                default:
                    Asserter.Assert(false, "orientation can not be other than specified in the enum", "got orientation", orientation.ToString());
                    return false;
            }

            StartRow = row;
            StartColumn = column;

            SetCell();
            return true;
        }

        void SetCell()
        {
            Asserter.Assert(StartRow != 0 && StartColumn != 0, "both the starting row and column should be set before calling GetCells");
            Button button = new Button();
            button.Style = (Style)button.FindResource("ShipCell");
            Grid.SetRow(button, StartRow);
            Grid.SetColumn(button, StartColumn);
            switch (orientation)
            {
                case Core.Orientation.Horizontal:
                    Grid.SetColumnSpan(button, Length);
                    break;
                case Core.Orientation.Vertical:
                    Grid.SetRowSpan(button, Length);
                    break;
                default:
                    Asserter.Assert(false, "orientation can not be other than specified in the enum", "got orientation", orientation.ToString());
                    break;
            }
            Cell = button;
        }

        public Button GetCell()
        {
            Asserter.Assert(Cell != null, "cell should not be null when program is trying to access it");
            return Cell;
        }

        public void DeleteShip(Grid grid)
        {
            grid.Children.Remove(Cell);
            IsPlaced = false;

            switch (orientation)
            {
                case Core.Orientation.Horizontal:
                    for (int i = 0; i < Length; i++)
                    {
                        Button button = new Button();
                        button.Style = (Style)button.FindResource("GridCell");
                        button.Command = Ship.PlaceShipCommand(button, grid);
                        Grid.SetRow(button, StartRow);
                        Grid.SetColumn(button, StartColumn + i);
                        grid.Children.Add(button);
                    }
                    break;
                case Core.Orientation.Vertical:
                    for (int i = 0; i < Length; i++)
                    {
                        Button button = new Button();
                        button.Style = (Style)button.FindResource("GridCell");
                        button.Command = Ship.PlaceShipCommand(button, grid);
                        Grid.SetRow(button, StartRow + i);
                        Grid.SetColumn(button, StartColumn);
                        grid.Children.Add(button);
                    }
                    break;
                default:
                    Asserter.Assert(false, "orientation can not be other than specified in the enum", "got orientation", orientation.ToString());
                    break;
            }
        }

        public bool ContainsSpot(int row, int column)
        {
            for (int i = 0; i < Length; i++)
            {
                if (this.orientation == Core.Orientation.Horizontal)
                {
                    if (StartRow == row && StartColumn + i == column) return true;
                } else
                {
                    if (StartRow + i == row && StartColumn == column) return true;
                }
            }

            return false;
        }

        public byte[] GetBytes()
        {
            byte[] bytes = new byte[1];
            bytes[0] = Convert.ToByte(this.Length * 2);
            for (int i = 0; i < Length; i++)
            {
                int y = this.StartRow;
                int x = this.StartColumn;
                if (this.orientation == Core.Orientation.Horizontal)
                {
                    x += i;
                } else
                {
                    y += i;
                }
                byte[] spot = new byte[2];
                Tcp.PutUint16(spot, x * 1000 + y);
                bytes = bytes.Concat(spot).ToArray();
            }
            return bytes;
        }

        public static RelayCommand PlaceShipCommand(Button button, Grid grid)
        {
            return new RelayCommand(o =>
                        {
                            if (GameState.CurrentShip == null)
                            {
                                return;
                            }
                            GlobalData.Instance.GameBoardVM.PlaceShip(button, grid);
                        });

        }

        public static RelayCommand GuessSpotCommand(Button button, Grid grid)
        {
            return new RelayCommand(o =>
            {
                Ship.DeletePreviousGuess(grid);
                Button newButton = new Button();
                newButton.Style = (Style)newButton.FindResource("GuessedSpot");
                newButton.Command = new RelayCommand(o => { Ship.DeletePreviousGuess(grid); });
                Grid.SetRow(newButton, Grid.GetRow(button));
                Grid.SetColumn(newButton, Grid.GetColumn(button));
                grid.Children.Remove(button);
                grid.Children.Add(newButton);
                GameState.GuessedPlace = newButton;
            });
        }

        public static void DeletePreviousGuess(Grid grid)
        {
            if (GameState.GuessedPlace == null) return;

            Button newButton = new Button();
            newButton.Style = (Style)newButton.FindResource("GridCell");
            newButton.Command = Ship.GuessSpotCommand(newButton, grid);
            Grid.SetRow(newButton, Grid.GetRow(GameState.GuessedPlace));
            Grid.SetColumn(newButton, Grid.GetColumn(GameState.GuessedPlace));
            grid.Children.Remove(GameState.GuessedPlace);
            grid.Children.Add(newButton);
            GameState.GuessedPlace = null;
        }
    }
}
