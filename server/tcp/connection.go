package tcp

import (
	"net"
)

type Connection struct {
    Id int
    conn net.Conn
    SendToChan chan TcpCommand
}

func CreateConnection(id int, conn net.Conn, sendToChan chan TcpCommand) Connection {
    return Connection{
        Id: id,
        conn: conn,
        SendToChan: sendToChan,
    }
}

func (c *Connection) NextMsg() (*TcpCommand, error) { //TODO: return value (tcp command)
    buf := make([]byte, 1024)
    n, err := c.conn.Read(buf)
    if err != nil {
        return nil, err
    }

    command, err := parseTcpCommand(buf[:n])

    if err != nil {
        return nil, err
    }

    command.Connection = c

    return command, nil
}

func (c *Connection) Close() {
    c.conn.Close()
}

func (c *Connection) Send(b []byte) (int, error) {
    return c.conn.Write(b)
}
