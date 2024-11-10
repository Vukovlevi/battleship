package tcp

import (
	"net"
)

type Connection struct {
    Id int
    conn net.Conn
    SendToChan chan TcpCommand
    GameOver bool
}

func CreateConnection(id int, conn net.Conn, sendToChan chan TcpCommand) Connection {
    return Connection{
        Id: id,
        conn: conn,
        SendToChan: sendToChan,
        GameOver: false, //set to true, if the server sent game over event to client, therefore the client can close the connection -> we don't handle it like an unexpected close event
    }
}

func (c *Connection) NextMsg() (*TcpCommand, error) {
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
