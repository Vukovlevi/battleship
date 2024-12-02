using Client.MVVM.Model;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;

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
                case CommandType.PlayerReady:
                    HandlePlayerReady();
                    break;
                case CommandType.MatchStart:
                    HandleMatchStart(command);
                    break;
                case CommandType.PlayerGuess:
                    HandlePlayerGuess(command);
                    break;
                case CommandType.GuessConfirm:
                    HandleGuessConfirm(command);
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

        static void HandlePlayerReady()
        {
            GlobalData.Instance.GameBoardVM.Status = "Az ellenfél rögzítette hajóit";
        }

        static void HandleMatchStart(TcpCommand command)
        {
            Asserter.Assert(GameState.state == State.EnemyTurn, "the state should be enemyturn when receiving match start command", "state", GameState.state.ToString());
            Asserter.Assert(command.data.Length == 1, "the data length of the match start command should be 1", "data len", command.data.Length.ToString());

            switch (command.data[0])
            {
                case 0:
                    GameState.state = State.YourTurn;
                    GlobalData.Instance.GameBoardVM.Status = "Te jössz";
                    break;
                case 1:
                    GameState.state = State.EnemyTurn;
                    GlobalData.Instance.GameBoardVM.Status = $"{GlobalData.Instance.EnemyUsername} jön";
                    break;
                default:
                    Asserter.Assert(false, "the data of match start command should always be 0 or 1", "got data", Tcp.GetByteAsString(command.data[0]));
                    break;
            }
        }

        static void HandlePlayerGuess(TcpCommand command)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                int spot = Tcp.GetUint16(command.data);
                int x = spot / 1000;
                int y = spot % 1000;

                Button button = new Button();
                button.Style = (Style)button.FindResource("ConfirmedSpot");
                Grid.SetRow(button, y);
                Grid.SetColumn(button, x);
                Panel.SetZIndex(button, 1);
                GlobalData.Instance.YourGrid.Children.Add(button);

                int shipIndex = GameState.Ships.FindIndex(s => s.ContainsSpot(y, x));
                if (shipIndex != -1)
                {
                    button.Background = new SolidColorBrush(Colors.Red);

                    Ship ship = GameState.Ships.Find(s => s.ContainsSpot(y, x));
                    ship.Health--;
                    if (ship.Health == 0)
                    {
                        GlobalData.Instance.GameBoardVM.YourRemainingShips--;
                    }
                }
                else
                {
                    button.Background = new SolidColorBrush(Colors.White);
                }

                GameState.state = State.YourTurn;
                GlobalData.Instance.GameBoardVM.Status = "Te jössz";
            });
        }

        static void HandleGuessConfirm(TcpCommand command)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                Asserter.Assert(GameState.state == State.EnemyTurn, "state should be enemyturn when receiving guess confirm", "got state", GameState.state.ToString());
                Asserter.Assert(command.data.Length == 1, "data length should be 1 when receiving guess confirm", "got len", command.data.Length.ToString());

                Button button = new Button();
                Grid.SetRow(button, Grid.GetRow(GameState.GuessedPlace));
                Grid.SetColumn(button, Grid.GetColumn(GameState.GuessedPlace));
                button.Style = (Style)button.FindResource("ConfirmedSpot");
                GlobalData.Instance.EnemyGrid.Children.Add(button);

                GlobalData.Instance.EnemyGrid.Children.Remove(GameState.GuessedPlace);
                GameState.GuessedPlace = null;

                byte feedback = Convert.ToByte((command.data[0] >> 6) & 0x3);
                switch (feedback)
                {
                    case 2:
                        button.Background = new SolidColorBrush(Colors.White);
                        GlobalData.Instance.GameBoardVM.Status = "Nem talált\n" + GlobalData.Instance.GameBoardVM.Status;
                        break;
                    case 3:
                        button.Background = new SolidColorBrush(Colors.Red);

                        byte didSink = Convert.ToByte((command.data[0] >> 5) & 0x1);
                        if (didSink == 1)
                        {
                            GlobalData.Instance.GameBoardVM.EnemyRemainingShips--;
                            GlobalData.Instance.GameBoardVM.Status = "Talált, süllyedt\n" + GlobalData.Instance.GameBoardVM.Status;
                        }
                        else
                        {
                            GlobalData.Instance.GameBoardVM.Status = "Talált\n" + GlobalData.Instance.GameBoardVM.Status;
                        }
                        break;
                    default:
                        Asserter.Assert(false, "feedback on guess confirm should only be 2 or 3", "got feedback", Tcp.GetByteAsString(feedback));
                        break;
                }
            });
        }

        static void HandleGameOver(TcpCommand command)
        {
            Asserter.Assert(command.data.Length == 2, "the length of the data should always be 2 when receiving game over command", "got len", command.data.Length.ToString());

            string message = "A játéknak vége!\n";
            byte reason = Convert.ToByte((command.data[0] >> 7) & 0x1);
            if (reason == 1)
            {
                message += "Az ellenfeled kilépett.";
                MessageBox.Show(message);
            } else
            {
                byte winner = Convert.ToByte((command.data[0] >> 6) & 0x1);
                if (winner == 0) message += "Nyertél!";
                else
                {
                    message += "Vesztettél!\n";
                    byte remainingShips = Convert.ToByte((command.data[0] >> 3) & 0x7);
                    message += $"Az ellenfelednek {remainingShips} hajója maradt, amit {Tcp.GetByteAsString(command.data[1])} lövésből tudtál volna elsüllyeszteni.";
                }
                MessageBox.Show(message);
            }
            GlobalData.Instance.MainVM.RestartGame();
        }
    }
}
