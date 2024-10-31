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

func (r *GameRoom) Loop() {
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