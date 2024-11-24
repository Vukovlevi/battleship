using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Client
{
    enum CommandType
    {
        JoinRequest = 1,
        DuplicateUsername = 2,
        MatchFound = 3,
        ShipsReady = 4,
        PlayerReady = 5,
        MatchStart = 6,
        PlayerGuess = 7,
        GuessConfirm = 8,
        GameOver = 9,
        CloseEvent = 10,
        Mismatch = 11,
    }

    enum MismatchType
    {
        VersionMismatch = 0,
        LengthMismatch = 1,
        CommandTypeMismatch = 2,
        DataMismatch = 3,
    }

    internal class TcpCommand
    {
        public CommandType commandType;
        public byte[] data;

        void SetCommandType(byte commandType)
        {
            switch (commandType)
            {
                case (byte)1:
                    this.commandType = CommandType.JoinRequest;
                    break;
                case (byte)2:
                    this.commandType = CommandType.DuplicateUsername;
                    break;
                case (byte)3:
                    this.commandType = CommandType.MatchFound;
                    break;
                case (byte)4:
                    this.commandType = CommandType.ShipsReady;
                    break;
                case (byte)5:
                    this.commandType = CommandType.PlayerReady;
                    break;
                case (byte)6:
                    this.commandType = CommandType.MatchStart;
                    break;
                case (byte)7:
                    this.commandType = CommandType.PlayerGuess;
                    break;
                case (byte)8:
                    this.commandType = CommandType.GuessConfirm;
                    break;
                case (byte)9:
                    this.commandType = CommandType.GameOver;
                    break;
                case (byte)10:
                    this.commandType = CommandType.CloseEvent;
                    break;
                case (byte)11:
                    this.commandType = CommandType.Mismatch;
                    break;
                default:
                    Asserter.Assert(false, "got unknown command type from server", "command type", Tcp.GetByteAsString(commandType));
                    break;
            }

        }

        public TcpCommand(byte commandType, byte[] data)
        {
            SetCommandType(commandType);
            this.data = data;
        }

        public byte[] EncodeToBytes()
        {
            byte[] bytes = new byte[Tcp.VERSION_SIZE + Tcp.MESSAGE_TYPE_SIZE];
            bytes[Tcp.VERSION_OFFSET] = Tcp.VERSION;
            bytes[Tcp.MESSAGE_TYPE_OFFSET] = 1;

            byte[] len = new byte[Tcp.DATA_LENGTH_SIZE];
            Tcp.PutUint16(len, data.Length);

            byte[] fullBytes = bytes.Concat(len).Concat(data).ToArray();

            return fullBytes;
        }
    }
}
