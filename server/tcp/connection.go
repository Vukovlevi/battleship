package tcp

import (
	"net"
)

type Connection struct {
    Id int
    conn net.Conn
    SendToChan chan TcpCommand
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
