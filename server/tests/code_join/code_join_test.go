package codejoin_test

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
    PORT = 42020
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

func TestCodeJoin(t *testing.T) {
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
    conn3 := createTcpConnection()
	time.Sleep(time.Millisecond * 10)

    username1 := []byte("vukovlevi")
    username2 := []byte("joska")

    cmd := tcp.TcpCommand{
        Type: tcp.CommandType.CodeJoin,
        Data: []byte{byte(len(username1))},
    }
    cmd.Data = append(cmd.Data, username1...)
    cmd.Data = append(cmd.Data, []byte("asd")...)

    _, err := conn1.Write(cmd.EncodeToBytes())
    assert.Nil(err, "writing code join command should not return error")
    time.Sleep(time.Millisecond * 10)

    cmd.Data = []byte{byte(len(username2))}
    cmd.Data = append(cmd.Data, username2...)
    cmd.Data = append(cmd.Data, []byte("asd")...)

    _, err = conn2.Write(cmd.EncodeToBytes())
    assert.Nil(err, "writing code join command should not return error")
    time.Sleep(time.Millisecond * 10)

    buf := make([]byte, 256)
    n, err := conn1.Read(buf)
    assert.Nil(err, "reading from conn1 should not return error")
    assert.Assert(n > 0, "reading from conn1 should return some bytes")
    cmdType := buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.MatchFound, "reading from conn1 should return matchfound command", "got command", cmdType)
    username := string(buf[tcp.HEADER_OFFSET:n])
    assert.Assert(username == string(username2), "the match found command on conn1 should return username2", "expected username", string(username2), "got username", username)
    time.Sleep(time.Millisecond * 10)

    buf = make([]byte, 256)
    n, err = conn2.Read(buf)
    assert.Nil(err, "reading from conn2 should not return error")
    assert.Assert(n > 0, "reading from conn2 should return some bytes")
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.MatchFound, "reading from conn2 should return matchfound command", "got command", cmdType)
    username = string(buf[tcp.HEADER_OFFSET:n])
    assert.Assert(username == string(username1), "the match found command on conn2 should return username1", "expected username", string(username1), "got username", username)

    _, err = conn3.Write(cmd.EncodeToBytes())
    assert.Nil(err, "writing code join command should not return error")
    time.Sleep(time.Millisecond * 10)

    buf = make([]byte, 256)
    n, err = conn3.Read(buf)
    assert.Nil(err, "reading from conn3 should not return error")
    assert.Assert(n > 0, "reading from conn3 should return some bytes")
    cmdType = buf[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
    assert.Assert(cmdType == tcp.CommandType.CodeJoinRejected, "reading from conn3 should return code join rejected command", "got command", cmdType)
}
