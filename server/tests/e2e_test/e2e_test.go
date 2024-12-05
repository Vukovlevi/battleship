package e2e_test

import (
	"encoding/binary"
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

	return outFile, logger.CreateLogger(outFile, outFile, true)
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

    //receiving game found msg
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing match found command on conn1", "n", n)
    assert.Nil(err, "there should be no error reading match found command on conn1", "err", err)
    username := string(buf[tcp.HEADER_OFFSET:n])
    messageType := buf[tcp.MESSAGE_TYPE_OFFSET:tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(messageType == 3, "command type should 3 on match found event on conn1", "got type", messageType)
    assert.Assert(username == "jozsi", "client should receive the other player's name", "received info", username)
    log.Debug("client got message on conn1", "msg", buf[:n])
	time.Sleep(time.Millisecond * 10)

    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing match found command on conn2", "n", n)
    assert.Nil(err, "there should be no error reading match found command on conn2", "err", err)
    username = string(buf[tcp.HEADER_OFFSET:n])
    messageType = buf[tcp.MESSAGE_TYPE_OFFSET:tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(messageType == 3, "command type should 3 on match found event on conn2", "got type", messageType)
    assert.Assert(username == "vukovlevi", "client should receive the other player's name", "received info", username)
    log.Debug("client got message on conn2", "msg", buf[:n])
    time.Sleep(time.Millisecond * 10)

    //testing ship parsing with command type mismatch error
    cmd := tcp.TcpCommand{
        Connection: nil,
        Type: tcp.CommandType.DuplicateUsername,
        Data: make([]byte, 0),
    }

    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing cmd mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on cmd mismatch while parsing ship", "err", err)
    cmdType := buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a cmd type mismatch", "got cmd type", cmdType)
    mismatchType := buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 2, "there should be a cmd type mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - spot len mismatch
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 5)
    cmd.Data[0] = 3
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1002)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - spot len mismatch
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 6)
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1002)
    cmd.Data[5] = 8
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - spot max value mismatch
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 5)
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1011)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1012)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - spot min value mismatch
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 5)
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1000)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1001)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - spots not beside each other horizontally
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 5)
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1003)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - spots not beside each other vertically
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 5)
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 3001)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - not enough ship
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 5)
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1002)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - too many 2-len ships
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 25)
    //putting 5 ships to pass the enough ship check before
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1002)
    cmd.Data[5] = 4
    binary.BigEndian.PutUint16(cmd.Data[6:], 2001)
    binary.BigEndian.PutUint16(cmd.Data[8:], 2002)
    cmd.Data[10] = 4
    binary.BigEndian.PutUint16(cmd.Data[11:], 3001)
    binary.BigEndian.PutUint16(cmd.Data[13:], 3002)
    cmd.Data[15] = 4
    binary.BigEndian.PutUint16(cmd.Data[16:], 4001)
    binary.BigEndian.PutUint16(cmd.Data[18:], 4002)
    cmd.Data[20] = 4
    binary.BigEndian.PutUint16(cmd.Data[21:], 5001)
    binary.BigEndian.PutUint16(cmd.Data[23:], 5002)
    //checking results
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with data mismatch - overlapping ships
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = make([]byte, 25)
    //putting 5 ships to pass the enough ship check before
    cmd.Data[0] = 4
    binary.BigEndian.PutUint16(cmd.Data[1:], 1001)
    binary.BigEndian.PutUint16(cmd.Data[3:], 1002)
    cmd.Data[5] = 4
    binary.BigEndian.PutUint16(cmd.Data[6:], 1002)
    binary.BigEndian.PutUint16(cmd.Data[8:], 1003)
    cmd.Data[10] = 4
    binary.BigEndian.PutUint16(cmd.Data[11:], 3001)
    binary.BigEndian.PutUint16(cmd.Data[13:], 3002)
    cmd.Data[15] = 4
    binary.BigEndian.PutUint16(cmd.Data[16:], 4001)
    binary.BigEndian.PutUint16(cmd.Data[18:], 4002)
    cmd.Data[20] = 4
    binary.BigEndian.PutUint16(cmd.Data[21:], 5001)
    binary.BigEndian.PutUint16(cmd.Data[23:], 5002)
    //checking results
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing data mismatch while parsing ship", "n", n)
    assert.Nil(err, "there should be no error on data mismatch while parsing ship", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a data mismatch", "got cmd type", cmdType)
    mismatchType = buf[tcp.HEADER_OFFSET]
    assert.Assert(mismatchType == 3, "there should be a data mismatch", "got mismatch type", mismatchType)

    //testing ship parsing with correct info - player1
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = getCorrectShips()
    //checking results
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing ship parsing on conn1", "n", n)
    assert.Nil(err, "there should be no error on ship parsing on conn1", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.PlayerReady, "there should be a player ready command on conn2", "got cmd type", cmdType)

    //testing ship parsing with correct info - player2
    cmd.Type = tcp.CommandType.ShipsReady
    cmd.Data = getCorrectShips()
    //checking results
    conn2.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "there should be a match start command after both players sent their ships", "n", n)
    assert.Nil(err, "there should be no error on receiving match start", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    data := buf[tcp.HEADER_OFFSET]
    assert.Assert(cmdType == tcp.CommandType.MatchStart, "there should be a match start command conn1", "got cmd type", cmdType)
    assert.Assert(data == 0, "player1 should be the starting one")

    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Assert(n > 0, "there should be a match start command after both players sent their ships", "n", n)
    assert.Nil(err, "there should be no error on receiving match start", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    data = buf[tcp.HEADER_OFFSET]
    assert.Assert(cmdType == tcp.CommandType.MatchStart, "there should be a match start command conn2", "got cmd type", cmdType)
    assert.Assert(data == 1, "player2 should not be the starting one")

    //testing unexpected ships sending
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing unexpected ships being sent", "n", n)
    assert.Nil(err, "there should be no error on testing unexpected ships being sent", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.Mismatch, "there should be a mismatch error on testing unexpected ships being sent", "got cmd type", cmdType)
    data = buf[tcp.HEADER_OFFSET]
    assert.Assert(data == tcp.CommandTypeMismatchCommand.Data[0], "the mismatch type should be command type mismatch", "got mismatch type", data)

    //testing spot sending from wrong connection
    cmd = tcp.TcpCommand{
        Type: tcp.CommandType.PlayerGuess,
        Data: make([]byte, 2),
    }
    binary.BigEndian.PutUint16(cmd.Data, 1001)
    conn2.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing wrong player sending guess", "n", n)
    assert.Nil(err, "there should be no error on testing wrong player sending guess", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.GuessConfirm, "there should be guess confirm command when wrong player sends guess", "got cmd type", cmdType)
    data = buf[tcp.HEADER_OFFSET]
    assert.Assert((data >> 6) == 0, "wrong player guessing should return notYourTurn", "returned bytes", buf[:n])

    //testing invalid spot sending
    binary.BigEndian.PutUint16(cmd.Data, 1011)
    log.Debug("cmd data on invalid spot testing", "data", cmd.Data)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing out of bound player guess", "n", n)
    assert.Nil(err, "there should be no error on testing out of bounds player guess", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.GuessConfirm, "there should be guess confirm command when guess is out of bounds", "got cmd type", cmdType)
    data = buf[tcp.HEADER_OFFSET]
    assert.Assert((data >> 6) == 1, "out of bounds guess should return invalidSpot", "returned bytes", buf[:n])

    //testing miss
    binary.BigEndian.PutUint16(cmd.Data, 10001)
    conn1.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing miss guess", "n", n)
    assert.Nil(err, "there should be no error on testing miss guess", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.GuessConfirm, "there should be guess confirm command when guess is missed", "got cmd type", cmdType)
    data = buf[tcp.HEADER_OFFSET]
    assert.Assert((data >> 6) == 2, "missed guess should return miss", "returned bytes", buf[:n])

    //the other connection should receive the guess as well for displaying reasons
    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing miss guess on the other user", "n", n)
    assert.Nil(err, "there should be no error on testing miss guess on the other user", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.PlayerGuess, "there should be player guess command when guess is missed on by the other user", "got cmd type", cmdType)
    spot := binary.BigEndian.Uint16(buf[tcp.HEADER_OFFSET:n])
    assert.Assert(binary.BigEndian.Uint16(cmd.Data) == spot, "the other players should receive the exact same spot", "returned bytes", buf[:n])
    time.Sleep(time.Millisecond * 10)

    //testing hit
    binary.BigEndian.PutUint16(cmd.Data, 10010)
    conn2.Write(cmd.EncodeToBytes())
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing hit guess", "n", n)
    assert.Nil(err, "there should be no error on testing hit guess", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.GuessConfirm, "there should be guess confirm command when guess is hit", "got cmd type", cmdType)
    data = buf[tcp.HEADER_OFFSET]
    assert.Assert((data >> 6) == 3, "missed guess should return hit", "returned bytes", buf[:n])

    //the other connection should receive the guess as well for displaying reasons
    time.Sleep(time.Millisecond * 10)
    buf = make([]byte, 256)
    n, err = conn1.Read(buf)
    assert.Assert(n > 0, "message should be longer than 0 byte when testing miss guess on the other user", "n", n)
    assert.Nil(err, "there should be no error on testing miss guess on the other user", "err", err)
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.PlayerGuess, "there should be player guess command when guess is missed on by the other user", "got cmd type", cmdType)
    spot = binary.BigEndian.Uint16(buf[tcp.HEADER_OFFSET:n])
    assert.Assert(binary.BigEndian.Uint16(cmd.Data) == spot, "the other players should receive the exact same spot", "returned bytes", buf[:n])
    time.Sleep(time.Millisecond * 10)

	//testing one player disconnecting
	conn1.Close()
	buf = make([]byte, 256)
	time.Sleep(time.Millisecond * 10)

	assert.Assert(len(testTcpServer.Connections) == 1, "there should be only 1 connection left", "con len", len(testTcpServer.Connections))
	assert.Assert(len(testGameServer.Rooms) == 0, "all rooms should be deleted", "rooms", len(testGameServer.Rooms))
	assert.Assert(len(testGameServer.MatchMaking.Players) == 0, "there should not be any player in mm", "mm len", len(testGameServer.MatchMaking.Players))

	n, err = conn2.Read(buf)
	assert.Nil(err, "reading from second connection should not return error", "err", err)
	assert.Assert(n > 0, "read bytes should be more than 0 when game over")
	log.Debug("client got message", "buf", buf[:n])
	assert.Assert(buf[tcp.MESSAGE_TYPE_OFFSET] == tcp.CommandType.GameOver, "server should send game over message", "got messagetype", buf[tcp.MESSAGE_TYPE_OFFSET])

	conn2.Close()
	time.Sleep(time.Millisecond * 10)
	assert.Assert(len(testTcpServer.Connections) == 0, "there should be no connections left", "con len", len(testTcpServer.Connections))
}

func getCorrectShips() []byte {
    data := make([]byte, 39)

    //ship1 - len: 2
    data[0] = 4
    binary.BigEndian.PutUint16(data[1:], 1001)
    binary.BigEndian.PutUint16(data[3:], 1002)

    //ship2 - len: 3
    data[5] = 6
    binary.BigEndian.PutUint16(data[6:], 2001)
    binary.BigEndian.PutUint16(data[8:], 3001)
    binary.BigEndian.PutUint16(data[10:], 4001)

    //ship3 - len: 3
    data[12] = 6
    binary.BigEndian.PutUint16(data[13:], 9007)
    binary.BigEndian.PutUint16(data[15:], 9008)
    binary.BigEndian.PutUint16(data[17:], 9009)

    //ship4 - len: 4
    data[19] = 8
    binary.BigEndian.PutUint16(data[20:], 7010)
    binary.BigEndian.PutUint16(data[22:], 8010)
    binary.BigEndian.PutUint16(data[24:], 9010)
    binary.BigEndian.PutUint16(data[26:], 10010)

    //ship5 - len: 5
    data[28] = 10
    binary.BigEndian.PutUint16(data[29:], 5004)
    binary.BigEndian.PutUint16(data[31:], 5005)
    binary.BigEndian.PutUint16(data[33:], 5006)
    binary.BigEndian.PutUint16(data[35:], 5007)
    binary.BigEndian.PutUint16(data[37:], 5008)

    return data
}
