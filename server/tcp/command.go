package tcp

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type CommandTypeEnum struct {
	JoinRequest       byte
	DuplicateUsername byte
	MatchFound        byte
	ShipsReady        byte
	PlayerReady       byte
	MatchStart        byte
	PlayerGuess       byte
	GuessConfirm      byte
	GameOver          byte
}

var (
	CommandType = CommandTypeEnum{
		JoinRequest:       1,
		DuplicateUsername: 2,
		MatchFound:        3,
		ShipsReady:        4,
		PlayerReady:       5,
		MatchStart:        6,
		PlayerGuess:       7,
		GuessConfirm:      8,
		GameOver:          9,
	}
)

type TcpCommand struct {
	Connection *Connection
	Type       byte
	Data       []byte
}

func (t *TcpCommand) EncodeToBytes() []byte {
	msg := make([]byte, HEADER_OFFSET)
	msg[VERSION_OFFSET] = VERSION
	msg[MESSAGE_TYPE_OFFSET] = t.Type
	binary.BigEndian.PutUint16(msg[DATA_LENGTH_OFFSET:], uint16(len(t.Data)))
	msg = append(msg, t.Data...)

	return msg
}

func parseTcpCommand(data []byte) (*TcpCommand, error) {
	if data[VERSION_OFFSET:VERSION_OFFSET + VERSION_SIZE][0] != VERSION {
		return nil, errors.New("version mismatch")
	}

	command := new(TcpCommand)
	command.Type = data[MESSAGE_TYPE_OFFSET:MESSAGE_TYPE_OFFSET + MESSAGE_TYPE_SIZE][0]

	length := binary.BigEndian.Uint16(data[DATA_LENGTH_OFFSET:DATA_LENGTH_OFFSET + DATA_LENGTH_SIZE])
	if HEADER_OFFSET + length != uint16(len(data)) {
		return nil, fmt.Errorf("length mismatch, expected(len):%d, got(len):%d, data:%v, data as string:%s", HEADER_OFFSET + length, len(data), data, string(data))
	}

	command.Data = data[HEADER_OFFSET:]
	return command, nil
}