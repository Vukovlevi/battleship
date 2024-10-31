package tcp_test

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
	PORT uint16 = 42690
)

func TestCommandEncoding(t *testing.T) {
	command1 := tcp.TcpCommand{
		Connection: nil,
		Type: 1,
		Data: []byte("vukovlevi"),
	}

	command2 := tcp.TcpCommand{
		Connection: nil,
		Type: 2,
		Data: make([]byte, 0),
	}

	bytes1 := command1.EncodeToBytes()
	bytes2 := command2.EncodeToBytes()

	name := []byte("vukovlevi")
	shouldbe1 := make([]byte, tcp.HEADER_OFFSET)
	shouldbe1[tcp.VERSION_OFFSET] = 1
	shouldbe1[tcp.MESSAGE_TYPE_OFFSET] = 1
	binary.BigEndian.PutUint16(shouldbe1[tcp.DATA_LENGTH_OFFSET:], uint16(len(name)))
	shouldbe1 = append(shouldbe1, name...)

	shouldbe2 := make([]byte, tcp.HEADER_OFFSET)
	shouldbe2[tcp.VERSION_OFFSET] = 1
	shouldbe2[tcp.MESSAGE_TYPE_OFFSET] = 2
	binary.BigEndian.PutUint16(shouldbe2[tcp.DATA_LENGTH_OFFSET:], uint16(0))

	t.Logf("bytes1: %v\nshould be: %v", bytes1, shouldbe1)
	t.Logf("bytes2: %v\nshould be: %v", bytes2, shouldbe2)

	if len(bytes1) != len(shouldbe1) {
		t.Fatalf("bytes array 1 should be equal, expected(len):%d, got(len):%d", len(shouldbe1), len(bytes1))
	}

	for i := range bytes1 {
		if bytes1[i] != shouldbe1[i] {
			t.Fatalf("every byte should be equal in array 1, expected(byte):%v, got(byte):%v, at(index):%d", shouldbe1[i], bytes1[i], i)
		}
	}

	if len(bytes2) != len(shouldbe2) {
		t.Fatalf("bytes array 2 should be equal, expected(len):%d, got(len):%d", len(shouldbe2), len(bytes2))
	}

	for i := range bytes2 {
		if bytes2[i] != shouldbe2[i] {
			t.Fatalf("every byte should be equal in array 2, expected(byte):%v, got(byte):%v, at(index):%d", shouldbe2[i], bytes2[i], i)
		}
	}
}

func TestTcpServer(t *testing.T) {
	wg := sync.WaitGroup{}

	name := []byte("vukovlevi")
	shouldbe1 := make([]byte, tcp.HEADER_OFFSET)
	shouldbe1[tcp.VERSION_OFFSET] = 1
	shouldbe1[tcp.MESSAGE_TYPE_OFFSET] = 1
	binary.BigEndian.PutUint16(shouldbe1[tcp.DATA_LENGTH_OFFSET:], uint16(len(name)))
	shouldbe1 = append(shouldbe1, name...)

	shouldbe2 := make([]byte, tcp.HEADER_OFFSET)
	shouldbe2[tcp.VERSION_OFFSET] = 1
	shouldbe2[tcp.MESSAGE_TYPE_OFFSET] = 2
	binary.BigEndian.PutUint16(shouldbe2[tcp.DATA_LENGTH_OFFSET:], uint16(0))
	messageChan := make(chan tcp.TcpCommand)

	shouldbes := [][]byte{shouldbe1, shouldbe2}

	go func() {
		msgCount := 0
		for {
			command, ok := <- messageChan
			if !ok {
				wg.Done()
				break
			}

			assert.NotNil(command.Connection, "connection should not be nil here", "array", msgCount + 1)
			assert.Assert(command.Type == shouldbes[msgCount][tcp.MESSAGE_TYPE_OFFSET], "message type is not equal", "expected", shouldbes[msgCount][tcp.MESSAGE_TYPE_OFFSET], "got", command.Type, "array", msgCount + 1)
			length := binary.BigEndian.Uint16(shouldbes[msgCount][tcp.DATA_LENGTH_OFFSET:tcp.DATA_LENGTH_OFFSET + tcp.DATA_LENGTH_SIZE])
			assert.Assert(uint16(len(command.Data)) == length, "length of the data should always be equal", "expected", length, "got", len(command.Data), "array", msgCount + 1)
			for i := range command.Data {
				assert.Assert(shouldbes[msgCount][tcp.HEADER_OFFSET + i] == command.Data[i], "all the bytes should be equal", "expected", shouldbes[msgCount][tcp.HEADER_OFFSET + i], "got", command.Data[i], "array", msgCount + 1)
			}
			
			msgCount++
			wg.Done()
		}
	}()

	outFile, err := os.Create("debug.txt")
	if err != nil {
		panic("debug file could not be deleted")
	}
	defer outFile.Close()
	outFile.Write([]byte("--- NEW TEST ---\n"))

	log := logger.CreateLogger(outFile, outFile, outFile, true)
	assert.SetLogger(&log)

	testServer := tcp.NewTcpServer(PORT, &log)
	go func() {
		testServer.Start(messageChan)
	}()


	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", PORT))
	assert.Nil(err, "connection to the test server should always be working", "err", err)

	conn.Write(shouldbe1)
	wg.Add(1)
	time.Sleep(time.Microsecond)
	conn.Write(shouldbe2)
	wg.Add(1)

	wg.Wait()
}