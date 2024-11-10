package tcp

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
)

const (
    VERSION byte = 1
    VERSION_SIZE = 1
    VERSION_OFFSET = 0
    MESSAGE_TYPE_SIZE = 1
    MESSAGE_TYPE_OFFSET = 1
    DATA_LENGTH_SIZE = 2
    DATA_LENGTH_OFFSET = 2
    HEADER_OFFSET = VERSION_SIZE + MESSAGE_TYPE_SIZE + DATA_LENGTH_SIZE
)

type Server struct {
    listener net.Listener
    Connections map[int]*Connection
    mutex sync.RWMutex
    log *logger.Logger
    port uint16
}

func NewTcpServer(port uint16, log *logger.Logger) *Server {
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    assert.Nil(err, "listener returned error", "err", err)

    return &Server{
        listener: listener,
        Connections: make(map[int]*Connection),
        mutex: sync.RWMutex{},
        log: log,
        port: port,
    }
}

func sendCloseCommandToGameServer(connection *Connection) {
    command := CloseCommand
    command.Connection = connection

    connection.SendToChan <- command
}

func readConnection(server *Server, connection Connection) {
    defer connection.Close()
    for {
        command, err := connection.NextMsg()
        close := false

        if err != nil {
            switch err {
            case io.EOF: //case: the connection was closed correctly
                msg := ""
                if connection.GameOver { //check if the connection close was initiated by the server or client
                    msg = "client closed connection on server command"
                } else {
                    msg = "closing connection on client event"
                    sendCloseCommandToGameServer(&connection)
                }
                server.log.Info(msg, "connectionId", connection.Id)
                close = true
            default: //case: there is another error
                if tcpError, ok := err.(TcpError); ok { //if the error was with message we send it back to inform client about bad command
                    server.log.Warning(err.Error(), "connectionId", connection.Id)
                    connection.Send(tcpError.command.EncodeToBytes())
                } else { //if the client closed connection by force
                    server.log.Warning("unknown error occured while reading connection", "connectionId", connection.Id, "err", err)
                    sendCloseCommandToGameServer(&connection) //handle all closing of gamerooms, mm and this tcp connection
                    close = true
                }
            }

            if close {
                server.mutex.Lock()
                delete(server.Connections, connection.Id)
                server.mutex.Unlock()

                server.log.Debug("connections info", "len", len(server.Connections))
                break
            }
        }

        server.log.Debug("got message", "connId", connection.Id, "msg", command)
        if command != nil {
            connection.SendToChan <- *command //if there is a command, send it to the gameserver or the room
        }
    }
}

func (s *Server) Start(sendToChan chan TcpCommand) {
    s.log.Info("starting tcp server", "port", s.port)
    id := 0
    for {
        conn, err := s.listener.Accept()

        if err != nil {
            s.log.Warning("error while acceptin new connection", "err", err)
        }

        connection := CreateConnection(id, conn, sendToChan)
        s.mutex.Lock()
        s.Connections[id] = &connection
        s.mutex.Unlock()

        s.log.Info("accepting new connection", "id", id, "addr", conn.RemoteAddr())
        s.log.Debug("connections info", "len", len(s.Connections))
        id++

        go readConnection(s, connection)
    }
}
