package game

import (
	"sync"

	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

type MatchMaking struct {
	log *logger.Logger
	Players map[*Player]bool
	MessageChan chan tcp.TcpCommand
	mutex sync.RWMutex
}

func (m *MatchMaking) HasPlayer(username string) bool {
	for user := range m.Players {
		if user.username == username {
			return true
		}
	}
	return false
}

func (m *MatchMaking) CanStartGame() bool {
	return len(m.Players) > 1
}

func (m *MatchMaking) SetupGame() *GameRoom {
	if !m.CanStartGame() {
		return nil
	}

	gameRoom := new(GameRoom)
	
	players := 0
	for player := range m.Players {
		if players == 0 {
			gameRoom.player1 = player
		} else {
			gameRoom.player2 = player
		}
		players++

		if players == 2 {
			break
		}
	}

	gameRoom.log = m.log
	gameRoom.MessageChan = make(chan tcp.TcpCommand)
	gameRoom.player1.connection.SendToChan = gameRoom.MessageChan
	gameRoom.player2.connection.SendToChan = gameRoom.MessageChan

	return gameRoom
}