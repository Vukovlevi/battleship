package e2e_test

import (
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/game"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
	PORT uint16 = 42426
)

func createLogger() (io.WriteCloser, logger.Logger) {
	outFile, err := os.Create("debug.txt")
	if err != nil {
		panic("debug file could not be deleted")
	}
	outFile.Write([]byte("--- NEW TEST ---\n"))

	return outFile, logger.CreateLogger(outFile, outFile, outFile, true)
}

func createTcpConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", PORT))
	assert.Nil(err, "connection should always be open", "err", err)
	return conn
}

func TestEndToEnd(t *testing.T) {
	outFile, log := createLogger()
	defer outFile.Close()
	assert.SetLogger(&log)

	testGameServer := game.NewGameServer(&log)
	testGameServer.Start()

	testTcpServer := tcp.NewTcpServer(PORT, &log)
	go func() {
		testTcpServer.Start(testGameServer.IncomingRequestChan)
	}()

	conn1 := createTcpConnection()
	conn2 := createTcpConnection()
	time.Sleep(time.Millisecond * 10)

	//testing connection phase
	assert.Assert(len(testTcpServer.Connections) == 2, "there should be 2 connections", "connections", testTcpServer.Connections)

	//testing version mismatch
	n, err := conn1.Write([]byte{0,1,0,0}) //version mismatch command
	time.Sleep(time.Millisecond * 10)
	assert.Nil(err, "writing should not return error", "err", err)
	assert.Assert(n == 4, "written bytes should be 4", "written bytes", n)
	buf := make([]byte, 256)
	n, err = conn1.Read(buf)
	assert.Nil(err, "reading connection should not return error on version mismatch", "err", err)
	assert.Assert(n > 0, "read bytes should be more than 0 on version mismatch")
	log.Debug("client got message", "bytes", buf[:n])

	//testing length mismatch
	n, err = conn2.Write([]byte{1,1,0,1}) //length mismatch command
	time.Sleep(time.Millisecond * 10)
	assert.Nil(err, "writing should not return error", "err", err)
	assert.Assert(n == 4, "written bytes should be 4", "written bytes", n)
	buf = make([]byte, 256)
	n, err = conn2.Read(buf)
	assert.Nil(err, "reading connection should not return error on length mismatch", "err", err)
	assert.Assert(n > 0, "read bytes should be more than 0 on length mismatch")
	log.Debug("client got message", "bytes", buf[:n])

	//testing command type mismatch
	badTcpCommand := tcp.TcpCommand{
		Type: tcp.CommandType.DuplicateUsername,
		Data: make([]byte, 0),
	}
	n, err = conn1.Write(badTcpCommand.EncodeToBytes())
	time.Sleep(time.Millisecond * 10)
	assert.Nil(err, "writing should not return error", "err", err)
	assert.Assert(n == 4, "written bytes should be 4", "written bytes", n)
	buf = make([]byte, 256)
	n, err = conn1.Read(buf)
	assert.Nil(err, "reading connection should not return error on command type mismatch", "err", err)
	assert.Assert(n > 0, "read bytes should be more than 0 on command type mismatch")
	log.Debug("client got message", "bytes", buf[:n])

	//testing first player joining
	joinPlayer1Command := tcp.TcpCommand{
		Type: tcp.CommandType.JoinRequest,
		Data: []byte("vukovlevi"),
	}
	n, err = conn1.Write(joinPlayer1Command.EncodeToBytes())
	time.Sleep(time.Millisecond * 10)
	assert.Nil(err, "writing should not return error", "err", err)
	assert.Assert(n == 13, "written bytes should be 13", "written bytes", n)
	assert.Assert(len(testGameServer.MatchMaking.Players) == 1, "there should be 1 player in mm", "found players", len(testGameServer.MatchMaking.Players))
	assert.Assert(!testGameServer.MatchMaking.CanStartGame(), "starting game shoud not be possible")
	assert.Assert(testGameServer.MatchMaking.HasPlayer(string(joinPlayer1Command.Data)), "user should exist in mm", "user", string(joinPlayer1Command.Data))
	assert.Assert(len(testGameServer.Rooms) == 0, "there should not be any gameroom")

	//testing duplicate username
	joinPlayer2Command := tcp.TcpCommand{
		Type: tcp.CommandType.JoinRequest,
		Data: []byte("vukovlevi"),
	}
	n, err = conn2.Write(joinPlayer2Command.EncodeToBytes())
	time.Sleep(time.Millisecond * 10)
	assert.Nil(err, "writing should not return error", "err", err)
	assert.Assert(n == 13, "written bytes should be 13", "written bytes", n)
	buf = make([]byte, 256)
	n, err = conn2.Read(buf)
	assert.Nil(err, "reading from conn should not return error on duplicate username")
	assert.Assert(n > 0, "read bytes should be more than 0 on duplicate username")
	log.Debug("client got message", "bytes", buf[:n])

	//testing second player joining
	joinPlayer2Command = tcp.TcpCommand{
		Type: tcp.CommandType.JoinRequest,
		Data: []byte("jozsi"),
	}
	n, err = conn2.Write(joinPlayer2Command.EncodeToBytes())
	time.Sleep(time.Millisecond * 10)
	assert.Nil(err, "writing should not return error", "err", err)
	assert.Assert(n == 9, "written bytes should be 9", "written bytes", n)
	assert.Assert(len(testGameServer.MatchMaking.Players) == 0, "there should be 0 player in mm", "found players", len(testGameServer.MatchMaking.Players))
	assert.Assert(!testGameServer.MatchMaking.CanStartGame(), "starting game shoud not be possible")
	assert.Assert(!testGameServer.MatchMaking.HasPlayer(string(joinPlayer1Command.Data)), "user should not exist in mm", "user", string(joinPlayer1Command.Data))
	assert.Assert(!testGameServer.MatchMaking.HasPlayer(string(joinPlayer2Command.Data)), "user should not exist in mm", "user", string(joinPlayer2Command.Data))
	assert.Assert(len(testGameServer.Rooms) == 1, "there should be 1 gameroom", "found gamerooms", len(testGameServer.Rooms))

	//testing one player disconnecting
	conn1.Close()
	buf = make([]byte, 256)
	time.Sleep(time.Millisecond * 10)

	assert.Assert(len(testTcpServer.Connections) == 1, "there should be only 1 connection left", "con len", len(testTcpServer.Connections))
	assert.Assert(len(testGameServer.Rooms) == 0, "all rooms should be deleted", "rooms", len(testGameServer.Rooms))
	assert.Assert(len(testGameServer.MatchMaking.Players) == 0, "there should not be any player in mm", "mm len", len(testGameServer.MatchMaking.Players))

	t.Log("anyad")
	n, err = conn2.Read(buf)
	t.Log("anyad2")
	assert.Nil(err, "reading from second connection should not return error", "err", err)
	assert.Assert(n > 0, "read bytes should be more than 0 when game over")
	log.Debug("client got message", "buf", buf[:n])
	assert.Assert(buf[tcp.MESSAGE_TYPE_OFFSET] == tcp.CommandType.GameOver, "server should send game over message", "got messagetype", buf[tcp.MESSAGE_TYPE_OFFSET])

	conn2.Close()
	time.Sleep(time.Millisecond * 10)
	assert.Assert(len(testTcpServer.Connections) == 0, "there should be no connections left", "con len", len(testTcpServer.Connections))
}
