package tcp

import (
	"encoding/binary"
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
    Close 			  byte
    Mismatch		  byte
    CodeJoin		  byte
    CodeJoinRejected  byte
}

type TcpError struct {
	msg string
	Command TcpCommand
}

func (t TcpError) Error() string {
	return t.msg
}

func CreateTcpError(msg string, command TcpCommand) TcpError {
	return TcpError{msg: msg, Command: command}
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
		Close: 			   10,
		Mismatch: 		   11,
		CodeJoin: 		   12,
		CodeJoinRejected:  13,
	}

	CloseCommand = TcpCommand{
		Connection: nil,
		Type: CommandType.Close,
		Data: make([]byte, 0),
	}

	VersionMismatchCommand = TcpCommand{
		Connection: nil,
		Type: CommandType.Mismatch,
		Data: []byte{0},
	}

	LengthMismatchCommand = TcpCommand{
		Connection: nil,
		Type: CommandType.Mismatch,
		Data: []byte{1},
	}

	CommandTypeMismatchCommand = TcpCommand{
		Connection: nil,
		Type: CommandType.Mismatch,
		Data: []byte{2},
	}

	DataMismatchCommand = TcpCommand{
		Connection: nil,
		Type: CommandType.Mismatch,
		Data: []byte{3},
	}

    CodeJoinRejectedCommand = TcpCommand{
        Connection: nil,
        Type: CommandType.CodeJoinRejected,
        Data: make([]byte, 0),
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
		err := CreateTcpError(fmt.Sprintf("version mismatch, expected: %d, got: %d", VERSION, data[VERSION_OFFSET:VERSION_OFFSET+VERSION_SIZE][0]), VersionMismatchCommand)
		return nil, err
	}

	command := new(TcpCommand)
	command.Type = data[MESSAGE_TYPE_OFFSET:MESSAGE_TYPE_OFFSET + MESSAGE_TYPE_SIZE][0]

	length := binary.BigEndian.Uint16(data[DATA_LENGTH_OFFSET:DATA_LENGTH_OFFSET + DATA_LENGTH_SIZE])
	if HEADER_OFFSET + length != uint16(len(data)) {
		err := CreateTcpError(fmt.Sprintf("length mismatch, expected(len): %d, got(len): %d, data: %v, data as string: %s", HEADER_OFFSET + length, len(data), data, string(data)), LengthMismatchCommand)
		return nil, err
	}

	command.Data = data[HEADER_OFFSET:]
	return command, nil
}
