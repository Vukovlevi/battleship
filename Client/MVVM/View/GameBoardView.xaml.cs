﻿using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Interop;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;

namespace Client.MVVM.View
{
    /// <summary>
    /// Interaction logic for GameBoardView.xaml
    /// </summary>
    public partial class GameBoardView : UserControl
    {
        private readonly int boardSize = 11;
        private readonly int gridCellSize = 40; //also modify styles (GridCell, GridHeaderElement) to match this size
        private readonly string[] letters = { "a", "b", "c", "d", "e", "f", "g", "h", "i", "j" };
        public GameBoardView()
        {
            InitializeComponent();
            this.DataContext = GlobalData.Instance.GameBoardVM;

            GenerateBoard(EnemyBoard);
            GenerateBoard(YourBoard);
        }

        void GenerateBoard(Grid grid)
        {
            grid.RowDefinitions.Clear();
            grid.ColumnDefinitions.Clear();
            grid.Children.Clear();

            for (int i = 0; i < boardSize; i++)
            {
                RowDefinition rowDefinition = new RowDefinition();
                rowDefinition.Height = new GridLength(gridCellSize);
                grid.RowDefinitions.Add(rowDefinition); 

                ColumnDefinition columnDefinition = new ColumnDefinition();
                columnDefinition.Width = new GridLength(gridCellSize);
                grid.ColumnDefinitions.Add(new ColumnDefinition()); 
            }

            for (int i = 1; i < boardSize; i++)
            {
                Label numberLabel = new Label();
                numberLabel.Style = (Style)FindResource("GridHeaderElement");
                numberLabel.Content = i.ToString();
                grid.Children.Add(numberLabel);
                Grid.SetRow(numberLabel, 0);
                Grid.SetColumn(numberLabel, i);

                Label letterLabel = new Label();
                letterLabel.Style = (Style)FindResource("GridHeaderElement");
                letterLabel.Content = letters[i - 1].ToUpper();
                grid.Children.Add(letterLabel);
                Grid.SetRow(letterLabel, i);
                Grid.SetColumn(letterLabel, 0);
            }

            for (int i = 1; i < boardSize; i++)
            {
                for (int j = 1; j < boardSize; j++)
                {
                    Button button = new Button();
                    button.Style = (Style)FindResource("GridCell");
                    grid.Children.Add(button);
                    Grid.SetRow(button, i);
                    Grid.SetColumn(button, j);
                }
            }
        }
    }
}
