using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Client
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
    internal static class GameState
    {
        public static State state;
        static MainWindow window;

        public static void SetWindow(MainWindow window)
        {
            GameState.window = window;
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
                default:
                    Asserter.Assert(false, "got unexpected command type from server", "command type", command.commandType.ToString());
                    break;
            }
        }

        static void HandleDuplicateUsername()
        {
            Asserter.Assert(state == State.WaitingForMatch, "getting duplicate username should only occur during waitingForMatch state", "got state", state.ToString());

            window.ShowSetUsername();
            MessageBox.Show($"A {MainWindow.username} felhasználónév már foglalt.\nKérem válasszon másikat!");
        }

        static void HandleMatchFound(TcpCommand command)
        {
            Asserter.Assert(state == State.WaitingForMatch, "state should be waiting for match when receiving match found command", "state", state.ToString());

        }
    }
}
