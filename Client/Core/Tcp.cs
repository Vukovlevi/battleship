using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Net;
using System.Net.Sockets;
using System.Data;
using System.Threading;
using System.Windows;

namespace Client.Core
{
    internal class Tcp
    {
        public static readonly byte VERSION = 2;
        public static readonly int VERSION_OFFSET = 0;
        public static readonly int VERSION_SIZE = 1;

        public static readonly int MESSAGE_TYPE_OFFSET = 1;
        public static readonly int MESSAGE_TYPE_SIZE = 1;

        public static readonly int DATA_LENGTH_OFFSET = 2;
        public static readonly int DATA_LENGTH_SIZE = 2;

        public static readonly int HEADER_OFFSET = 4;

        string address { get; set; }
        int port { get; set; }
        TcpClient? client = null;
        bool listen = false;

        public Tcp(string address = "vukovlevi.dev", int port = 42069)
        {
            Asserter.Assert(client == null, "tcp client should be only initialized once");

            this.address = address;
            this.port = port;
        }

        public void Connect()
        {
            try
            {
                client = new TcpClient(address, port);
            }
            catch (Exception ex)
            {
                Asserter.Assert(false, "an error occured while connecting client to server", "err", ex.Message);
            }
        }

        public NetworkStream GetNetworkStream()
        {
            try
            {
                var stream = client.GetStream();
                Asserter.Assert(stream != null, "getting the network stream should never return null");
                return stream;
            }
            catch (Exception ex)
            {
                Asserter.Assert(false, "an error occured while getting networkstream", "err", ex.Message);
                return null;
            }
        }

        public void Close()
        {
            listen = false;
            client.Close();
            client = null;
        }

        public void Send(byte[] data)
        {
            Asserter.Assert(client != null, "client should not be null when sending tcpcommand");
            Asserter.Assert(data.Length >= 4, "the length of tcpcommand should be at least 4", "got len", data.Length.ToString());
            try
            {
                var stream = GetNetworkStream();
                stream.Write(data, 0, data.Length);
            }
            catch (Exception ex)
            {
                Asserter.Assert(false, "an error occured while sending tcpcommand", "err", ex.Message);
            }
        }

        public void Listen()
        {
            listen = true;
            try
            {
                var stream = GetNetworkStream();
                byte[] buffer = new byte[1024];
                while (listen)
                {
                    if (stream.DataAvailable && listen)
                    {
                        int n = stream.Read(buffer, 0, buffer.Length);
                        Asserter.Assert(n <= buffer.Length, $"there should never be a message that reaches {buffer.Length} bytes", "read bytes", n.ToString());

                        HandleTcpCommand(buffer.Take(n).ToArray());
                    }
                    else
                    {
                        Thread.Sleep(1);
                    }
                }
            }
            catch (Exception ex)
            {
                Asserter.Assert(false, "an error occured while reading connection", "err", ex.Message);
            }
        }

        public static string GetByteAsString(byte x)
        {
            return Convert.ToInt32(x).ToString();
        }

        public static int GetUint16(byte[] data)
        {
            Asserter.Assert(data.Length >= 2, "getting a uint16 from this array is impossible because of its length", "length", data.Length.ToString());
            return data[0] << 8 | data[1];
        }

        public static void PutUint16(byte[] data, int value)
        {
            Asserter.Assert(data.Length >= 2, "putting a uint16 to this array is impossible because of its length", "length", data.Length.ToString());
            byte firstByte = (byte)((value & 0xFF00) >> 8);
            byte secondByte = (byte)(value & 0xFF);
            data[0] = firstByte;
            data[1] = secondByte;
        }

        void HandleTcpCommand(byte[] data)
        {
            byte version = data.Skip(VERSION_OFFSET).Take(VERSION_SIZE).ToArray()[0];
            Asserter.Assert(version == VERSION, "the version of the client does not match the version of the server", "client version", GetByteAsString(VERSION), "server version", GetByteAsString(version));

            int len = GetUint16(data.Skip(DATA_LENGTH_OFFSET).Take(DATA_LENGTH_SIZE).ToArray());
            string display = "";
            foreach (var elem in data)
            {
                display += GetByteAsString(elem) + " ";
            }
            Asserter.Assert(HEADER_OFFSET + len == data.Length, "the length of the message doesnt equals to the said length", "said length", (len + 4).ToString(), "real length", data.Length.ToString(), "data", display);

            byte type = data.Skip(MESSAGE_TYPE_OFFSET).Take(MESSAGE_TYPE_SIZE).ToArray()[0];
            TcpCommand command = new TcpCommand(type, data.Skip(HEADER_OFFSET).ToArray());

            GameState.HandleCommand(command);
        }
    }
}
