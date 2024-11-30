using Client.Core;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

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
        private Button? Cell { get; set; } = null;

        public Ship(int id, int length, Core.Orientation orientation)
        {
            Id = id;
            Length = length;
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
    }
}
