package tcp

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
)

type Server struct {
    listener net.Listener
    connections map[int]*Connection
    mutex sync.RWMutex
    log *logger.Logger
    port uint16
}

func NewTcpServer(port uint16, log *logger.Logger) *Server {
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    assert.Nil(err, "listener returned error", "err", err)

    return &Server{
        listener: listener,
        connections: make(map[int]*Connection),
        mutex: sync.RWMutex{},
        log: log,
        port: port,
    }
}

func readConnection(server *Server, connection Connection) {
    for {
        msg, err := connection.NextMsg()

        if err != nil {
            if err == io.EOF {
                server.log.Info("closing connection", "connectionId", connection.Id)
                //TODO: close this connection correctly -> connection closed by client -- use the channel
            } else {
                server.log.Warning("unknown error occured while reading connection", "err", err)
            }

            server.mutex.Lock()
            delete(server.connections, connection.Id)
            server.mutex.Unlock()

            server.log.Info("closing connection", "id", connection.Id)
            server.log.Debug("connections info", "len", len(server.connections))
            break
        }

        server.log.Debug("got message", "connId", connection.Id, "msg", msg)
        connection.SendToChan <- *msg
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

        connection := Connection{Id: id, conn: conn, SendToChan: sendToChan}
        s.mutex.Lock()
        s.connections[id] = &connection
        s.mutex.Unlock()

        s.log.Info("accepting new connection", "id", id, "addr", conn.RemoteAddr())
        s.log.Debug("connections info", "len", len(s.connections))
        id++

        go readConnection(s, connection)
    }
}
