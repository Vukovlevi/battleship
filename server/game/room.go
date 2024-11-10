package game

import (
	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
	correctlyClosed byte = 0x00
	playerLeftClosed byte = 0x80

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
	if command != nil { //send the players the initiating close command -> should be game over
		r.player1.connection.Send(command.EncodeToBytes())
		r.player2.connection.Send(command.EncodeToBytes())
	}

	close(r.MessageChan)
	r.closeChan <- r //inform the game server about this room being closed
}

func (r *GameRoom) GetStatsByte(closer *Player, sendingTo *Player) []byte { //set closer to nil if the room is closed because of win
	if closer != nil { //this means the game is not over correctly, a player has closed the connection
		assert.Assert(closer != sendingTo, "sending stats to the player closing the connection is not possible", "closer", closer.username, "sendintTo", sendingTo.username)

		firstByte := playerLeftClosed | loser
		return []byte{firstByte, 0}
	}
    //TODO: return back the real data if game is closed by win

	return make([]byte, 0)
}

func (r *GameRoom) HandleConnectionClosed(command *tcp.TcpCommand) *tcp.TcpCommand { //this function should only be called if the client closed connection, not when closing is because of win
	closer := r.player2
	sendTo := r.player1
	if command.Connection == r.player1.connection {
		closer = r.player1
		sendTo = r.player2
	}
	sendTo.connection.GameOver = true

	r.log.Debug("gameroom closing", "close initiated by", closer.username)

	cmd := tcp.TcpCommand{
		Connection: sendTo.connection,
		Type: tcp.CommandType.GameOver,
		Data: r.GetStatsByte(closer, sendTo),
	}

	r.log.Debug("got statistics for other user", "stat", cmd.Data)
	return &cmd
}

func (r *GameRoom) SendMatchFound() { //when a room is set up, send the correct command to the clients
    cmd := tcp.TcpCommand{
        Connection: r.player1.connection,
        Type: tcp.CommandType.MatchFound,
        Data: []byte(r.player2.username),
    }
    r.player1.connection.Send(cmd.EncodeToBytes())

    cmd.Connection = r.player2.connection
    cmd.Data = []byte(r.player1.username)
    r.player2.connection.Send(cmd.EncodeToBytes())
}

func (r *GameRoom) Loop() {
	for {
		command, ok := <- r.MessageChan
		if !ok {
			r.log.Debug("gameroom connection closed")
			break
		}

		switch command.Type {
		case tcp.CommandType.Close: //close room if close command occures
			cmd := r.HandleConnectionClosed(&command)
			r.CloseRoom(cmd)
			return
        default: //any other command type is unexpected
            cmd := tcp.CommandTypeMismatchCommand
            cmd.Connection = command.Connection
            command.Connection.Send(cmd.EncodeToBytes())
		}

		r.log.Debug("gameroom got command", "command", command)
	}
}
