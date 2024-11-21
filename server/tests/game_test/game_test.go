package game_test

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/game"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

func TestConnectingToGameServer(t *testing.T) {
	outFile, err := os.Create("debug.txt")
	if err != nil {
		panic("debug file could not be deleted")
	}
	defer outFile.Close()
	outFile.Write([]byte("--- NEW TEST ---\n"))

	log := logger.CreateLogger(outFile, outFile, true)
	assert.SetLogger(&log)

	testGameServer := game.NewGameServer(&log)
	testGameServer.Start()

	r, w := net.Pipe()
	connection := tcp.CreateConnection(0, w, testGameServer.IncomingRequestChan)
	connection.SendToChan = testGameServer.IncomingRequestChan
	command := tcp.TcpCommand{
		Connection: &connection,
		Type: tcp.CommandType.DuplicateUsername,
		Data: []byte("vukovlevi"),
	}

	connection.SendToChan <- command
	time.Sleep(time.Millisecond * 10)

	//testing command type mismatch
	buf := make([]byte, 256)
	n, err := r.Read(buf)
	assert.Nil(err, "reading from connectiond should not return error")
	assert.Assert(n > 0, "returned message should not be of length 0")
	log.Debug("client got message", "buf", buf[:n])

	//testing first player connection
	r, w = net.Pipe()
	connection = tcp.CreateConnection(0, w, testGameServer.IncomingRequestChan)
	command.Connection = &connection
	command.Type = tcp.CommandType.JoinRequest

	connection.SendToChan <- command
	time.Sleep(time.Millisecond * 10)

	assert.Assert(len(testGameServer.MatchMaking.Players) == 1, "there should be 1 players in mm", "player count", len(testGameServer.MatchMaking.Players), "sent command", command, "players", testGameServer.MatchMaking.Players)
	assert.Assert(!testGameServer.MatchMaking.CanStartGame(), "starting the game should not be possible")
	assert.Assert(testGameServer.MatchMaking.HasPlayer(string(command.Data)), "user with that name should exist", "username", string(command.Data))
	game := testGameServer.MatchMaking.SetupGame()
	if game != nil {
		log.Error("setting up a game should return nil", "got value", game)
		panic("assert triggered")
	}
	assert.Assert(len(testGameServer.Rooms) == 0, "there should be 0 gameroom", "gameroom count", len(testGameServer.Rooms))

	//testing duplicate username
	connection.SendToChan <- command
	time.Sleep(time.Millisecond * 10)

	buf = make([]byte, 256)
	n, err = r.Read(buf)
	assert.Nil(err, "reading from connectiond should not return error")
	assert.Assert(n > 0, "returned message should no be of length 0")
	log.Debug("client got message", "buf", buf[:n])

	//testing second connection
	_, w = net.Pipe()
	connection = tcp.CreateConnection(1, w, testGameServer.IncomingRequestChan)
	command.Connection = &connection

	command.Data = []byte("joska")
	connection.SendToChan <- command
	time.Sleep(time.Millisecond * 10)

	assert.Assert(len(testGameServer.MatchMaking.Players) == 0, "there should be 0 players in mm", "player count", len(testGameServer.MatchMaking.Players), "sent command", command, "players", testGameServer.MatchMaking.Players)
	assert.Assert(!testGameServer.MatchMaking.CanStartGame(), "starting the game should not be possible")
	assert.Assert(!testGameServer.MatchMaking.HasPlayer(string(command.Data)), "user with that name should not exist", "username", string(command.Data))
	game = testGameServer.MatchMaking.SetupGame()
	if game != nil {
		log.Error("setting up a game should return nil", "got value", game)
		panic("assert triggered")
	}
	assert.Assert(len(testGameServer.Rooms) == 1, "there should be 1 gameroom", "gameroom count", len(testGameServer.Rooms))
}
