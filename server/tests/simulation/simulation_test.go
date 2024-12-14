package simulation_test

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
	PORT uint16 = 42427
)

var (
    positions = []int{1001, 1002, 2001, 3001, 4001, 9007, 9008, 9009, 7010, 8010, 9010, 10010, 5004, 5005, 5006, 5007, 5008}
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

func connectPlayers(conn1, conn2 net.Conn) {
	joinPlayerCommand := tcp.TcpCommand{
		Type: tcp.CommandType.JoinRequest,
		Data: []byte("vukovlevi"),
	}
	conn1.Write(joinPlayerCommand.EncodeToBytes())

    joinPlayerCommand.Data = []byte("joska")
	conn2.Write(joinPlayerCommand.EncodeToBytes())

	time.Sleep(time.Millisecond * 10)
}

func sendShips(conn1, conn2 net.Conn) {
    data := getCorrectShips()
    sendShipsCommand := tcp.TcpCommand{
        Type: tcp.CommandType.ShipsReady,
        Data: data,
    }
    conn1.Write(sendShipsCommand.EncodeToBytes())
    conn2.Write(sendShipsCommand.EncodeToBytes())

	time.Sleep(time.Millisecond * 10)
}

func TestGame(t *testing.T) {
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

    connectPlayers(conn1, conn2)
    sendShips(conn1, conn2)
	time.Sleep(time.Millisecond * 10)

    buf1 := make([]byte, 256)
    buf2 := make([]byte, 256)
    data := make([]byte, 2)
    for _, pos := range positions {
        binary.BigEndian.PutUint16(data, uint16(pos))
        cmd := tcp.TcpCommand{
            Type: tcp.CommandType.PlayerGuess,
            Data: data,
        }

        //testing conn1 guessing
        conn1.Write(cmd.EncodeToBytes())
        time.Sleep(time.Millisecond * 50)
        n2, err2 := conn2.Read(buf2)

        n1, err1 := conn1.Read(buf1)
        assert.Nil(err1, "there should be no error on reading guess confirm from conn1")
        assert.Assert(n1 > 0, "reading message from conn1 should be more than 0 bytes")
        over1 := testPositionResponse(pos, buf1[:n1], true, log)
        if over1 {
            over2 := testPositionResponse(pos, buf2[:n2], false, log)
            assert.Assert(over2, "conn2 should get the game over event as well")
            continue
        }

        //testing conn2 guessing
        conn2.Write(cmd.EncodeToBytes())
        time.Sleep(time.Millisecond * 50)
        n1, err1 = conn1.Read(buf1)

        n2, err2 = conn2.Read(buf2)
        assert.Nil(err2, "there should be no error on reading guess confirm from conn1")
        assert.Assert(n2 > 0, "reading message from conn1 should be more than 0 bytes")
        over2 := testPositionResponse(pos, buf2[:n2], false, log)
        assert.Assert(!over2, "conn2 should never be the reason for game over (the winner)")
    }

    conn1.Close()
    conn2.Close()
    time.Sleep(time.Millisecond * 10)
}

func testPositionResponse(pos int, resp []byte, won bool, log logger.Logger) bool {
    switch pos {
    case 1002:
    case 4001:
    case 9009:
    case 10010:
        commandType := resp[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
        assert.Assert(commandType == tcp.CommandType.GuessConfirm, "command type should be guess confirm", "got cmd type", commandType)

        hit := resp[tcp.HEADER_OFFSET]
        assert.Assert((hit >> 6) & 3 == 3, "every guess should be a hit", "got byte", hit)
        assert.Assert((hit >> 5) & 1 == 1, "these guesses should sink a ship", "got byte", hit)
        assert.Assert(len(resp[tcp.HEADER_OFFSET:]) > 1, "the length of the received message should be more than 1, because it should contain the ships positions", "got data", resp[tcp.HEADER_OFFSET:])

        return false
    case 5008:
        commandType := resp[tcp.MESSAGE_TYPE_OFFSET: tcp.MESSAGE_TYPE_OFFSET + tcp.MESSAGE_TYPE_SIZE][0]
        assert.Assert(commandType == tcp.CommandType.GameOver, "command type should be game over", "got cmd type", commandType)

        firstByte := resp[tcp.HEADER_OFFSET]
        remainingHealth := resp[tcp.HEADER_OFFSET + 1]
        assert.Assert(firstByte >> 7 == 0, "game should be over correctly", "got first byte", firstByte)
        if won {
            assert.Assert((firstByte >> 6) & 1 == 0, "this player should have won", "got first byte", firstByte)
            assert.Assert((firstByte >> 3) & 7 == 0, "there should be no remaining ships if this player has won", "got first byte", firstByte)
            assert.Assert(remainingHealth == 0, "there should be no remaining health if this player has won", "got remaining health", remainingHealth)
        } else {
            assert.Assert((firstByte >> 6) & 1 == 1, "this player should have lost", "got first byte", firstByte)
            assert.Assert((firstByte >> 3) & 7 == 1, "there should be 1 remaining ship if this player has lost", "got first byte", firstByte)
            assert.Assert(remainingHealth == 1, "there should be 1 remaining health if this player has lost", "got remaining health", remainingHealth)

            remainingSpots := resp[tcp.HEADER_OFFSET + 2:]
            assert.Assert(len(remainingSpots) == 2, "there should be 2 bytes for the last remaining spot", "got bytes", len(remainingSpots))
            pos := binary.BigEndian.Uint16(remainingSpots)
            assert.Assert(pos == 5008, "the last remaining positions should be 5008", "got last remaining positions", pos)
        }

        return true
    default:
        return false
    }
    return false
}
