﻿using Client.Core;
using Client.MVVM.ViewModel;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.RightsManagement;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Controls;

namespace Client.MVVM.Model
{
    class GlobalData
    {
        public static GlobalData Instance = new GlobalData();
        public Tcp Tcp { get; set; }
        public MainViewModel MainVM { get; set; }
        public LoginViewModel LoginVM { get; set; }
        public GameBoardViewModel GameBoardVM { get; set; }
        public string Username { get; set; }
        public string EnemyUsername { get; set; }
        public int BoardSize = 11;
        public Grid EnemyGrid { get; set; }
        public Grid YourGrid { get; set; }
        public TextBox GameCode { get; set; }
        public List<int> YourNotGuessedShipSpots = new List<int>(); //the spots that hold your ships and have not been guessed by the enemy
    }
}
