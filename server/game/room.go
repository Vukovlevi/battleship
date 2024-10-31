package game

import (
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

type GameRoom struct {
	log *logger.Logger
	player1     *Player
	player2     *Player
	MessageChan chan tcp.TcpCommand
}

func (r *GameRoom) CloseRoom() {
	r.log.Info("closing room", "player1", r.player1.username, "player2", r.player2.username)
	r.player1.connection.Close()
	r.player2.connection.Close()
	close(r.MessageChan)
}

func (r *GameRoom) Loop() {
	defer r.CloseRoom()
	for {
		command, ok := <- r.MessageChan
		if !ok {
			r.log.Debug("gameroom connection closed")
			//TODO: close every players connection and finish the game
			break
		}

		r.log.Debug("gameroom got command", "command", command)
	}
}