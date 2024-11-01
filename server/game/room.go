package game

import (
	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
	correctlyClosed byte = 0x80
	playerLeftClosed byte = 0x00
	
	winner byte = 0x40
	loser byte = 0x00
)

type GameRoom struct {
	log *logger.Logger
	player1     *Player
	player2     *Player
	MessageChan chan tcp.TcpCommand
	closeChan chan *GameRoom
}

func (r *GameRoom) CloseRoom(command *tcp.TcpCommand) {
	r.log.Info("closing room", "player1", r.player1.username, "player2", r.player2.username)
	if command != nil {
		r.player1.connection.Send(command.EncodeToBytes())
		r.player2.connection.Send(command.EncodeToBytes())
	}
	r.player1.connection.Close()
	r.player2.connection.Close()

	close(r.MessageChan)
	r.closeChan <- r
}

func (r *GameRoom) GetStatsByte(closer *Player, sendingTo *Player) []byte {
	if closer != nil {
		assert.Assert(closer != sendingTo, "sending stats to the player closing the connection is not possible", "closer", closer.username, "sendintTo", sendingTo.username)

		firstByte := playerLeftClosed | loser
		return []byte{firstByte, 0}
	}

	return make([]byte, 0)
}

func (r *GameRoom) HandleConnectionClosed(command *tcp.TcpCommand) *tcp.TcpCommand {
	closer := r.player2
	sendTo := r.player1
	if command.Connection == r.player1.connection {
		closer = r.player1
		sendTo = r.player2
	}
	
	cmd := tcp.TcpCommand{
		Connection: sendTo.connection,
		Type: tcp.CommandType.GameOver,
		Data: r.GetStatsByte(closer, sendTo),
	}

	return &cmd
}

func (r *GameRoom) Loop() {
	defer r.CloseRoom(nil)
	for {
		command, ok := <- r.MessageChan
		if !ok {
			r.log.Debug("gameroom connection closed")
			break
		}

		switch command.Type {
		case tcp.CommandType.Close:
			cmd := r.HandleConnectionClosed(&command)
			r.CloseRoom(cmd)
			return
		}

		r.log.Debug("gameroom got command", "command", command)
	}
}