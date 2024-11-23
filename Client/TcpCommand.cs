using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Client
{
    internal class TcpCommand
    {
        public byte commandType;
        public byte[] data;

        public TcpCommand(byte commandType, byte[] data)
        {
            this.commandType = commandType;
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
