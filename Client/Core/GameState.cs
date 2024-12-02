using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

namespace Client.Core
{
    enum State
    {
        SetUsername = 1,
        WaitingForMatch = 2,
        PlacingShips = 3,
        YourTurn = 4,
        EnemyTurn = 5,
        GameOver = 6,
    }

    enum Orientation
    {
        Horizontal = 1,
        Vertical = 2,
    }

    internal static class GameState
    {
        public static State state;
        static MainWindow window;
        static Tcp tcp;
        public static readonly int DefaultShipCount = 5;
        public static Core.Orientation orientation;
        public static Ship? CurrentShip = null;
        public static List<Ship> Ships = new List<Ship>();
        public static Button? GuessedPlace = null;

        public static void SetWindow(MainWindow window)
        {
            GameState.window = window;
        }

        public static void SetTcp()
        {
            tcp = GlobalData.Instance.Tcp;
        }

        public static void SetShips()
        {
            Ships.Clear();
            int id = 0;
            for (int i = 2; i <= 5; i++)
            {
                Ship ship = new Ship(id, i, GameState.orientation);
                Ships.Add(ship);
                if (i == 3)
                {
                    id++;
                    ship = new Ship(id, i, GameState.orientation);
                    Ships.Add(ship);
                }
                id++;
            }
        }

        public static void HandleCommand(TcpCommand command)
        {
            switch (command.commandType)
            {
                case CommandType.DuplicateUsername:
                    HandleDuplicateUsername();
                    break;
                case CommandType.MatchFound:
                    HandleMatchFound(command);
                    break;
                case CommandType.GameOver:
                    HandleGameOver(command);
                    break;
                default:
                    Asserter.Assert(false, "got unexpected command type from server", "command type", command.commandType.ToString());
                    break;
            }
        }

        static void HandleDuplicateUsername()
        {
            Asserter.Assert(state == State.WaitingForMatch, "getting duplicate username should only occur during waitingForMatch state", "got state", state.ToString());

            GlobalData.Instance.LoginVM.DuplicateUsername();
        }

        static void HandleMatchFound(TcpCommand command)
        {
            Asserter.Assert(state == State.WaitingForMatch, "state should be waiting for match when receiving match found command", "state", state.ToString());

            GlobalData.Instance.GameBoardVM.SetUsernames(GlobalData.Instance.Username, ASCIIEncoding.ASCII.GetString(command.data));
        }

        static void HandleGameOver(TcpCommand command)
        {
            MessageBox.Show("TODO: change this -- GAME OVER");
            GlobalData.Instance.MainVM.RestartGame();
        }
    }
}
