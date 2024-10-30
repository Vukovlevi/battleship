package tcp

import (
	"net"
)

type Connection struct {
    id int
    conn net.Conn
}

func (c *Connection) NextMsg() (string, error) { //TODO: return value (tcp command)
    buf := make([]byte, 1024, 1024)
    n, err := c.conn.Read(buf)

    msg := buf[:n]

    return string(msg), err
}
