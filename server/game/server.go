package game

import (
	"sync"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

type GameServer struct {
	log *logger.Logger
	MatchMaking     MatchMaking
	Rooms           map[*GameRoom]bool
	IncomingRequestChan chan tcp.TcpCommand 
	GameRoomCloseChan chan *GameRoom
	mutex sync.RWMutex
}

func NewGameServer(log *logger.Logger) *GameServer {
	return &GameServer{
		log: log,
		MatchMaking: MatchMaking{Players: make(map[*Player]bool), log: log},
		Rooms:       make(map[*GameRoom]bool),
		IncomingRequestChan: make(chan tcp.TcpCommand),
		GameRoomCloseChan: make(chan *GameRoom),
		mutex: sync.RWMutex{},
	}
}

func (g *GameServer) Start() {
	go func() {
		for {
			room, ok := <- g.GameRoomCloseChan
			assert.Assert(ok, "gameroom closing channel should never be closed")
			g.mutex.Lock()
			delete(g.Rooms, room)
			g.mutex.Unlock()
		}
	}()

	go func() {
		g.log.Info("game server started")
		for {
			msg, ok := <- g.IncomingRequestChan
			assert.Assert(ok, "the channel of the game server should always be open")
			g.log.Debug("game server got a new message", "msg", msg)

			go handleJoinRequest(g, msg)
		}
	}() 
}

func handleJoinRequest(g *GameServer, command tcp.TcpCommand) {
	switch command.Type {
	case tcp.CommandType.Close:
		if player, ok := g.MatchMaking.HasConnection(command.Connection); ok {
			g.MatchMaking.mutex.Lock()
			delete(g.MatchMaking.Players, player)
			g.MatchMaking.mutex.Unlock()
		}
	case tcp.CommandType.JoinRequest:
		break
	default:
		g.log.Warning("command type mismatch while handling join request", "expected type", tcp.CommandType.JoinRequest, "got type", command.Type, "connectionId", command.Connection.Id)
		cmd := tcp.CommandTypeMismatchCommand
		cmd.Connection = command.Connection
		g.log.Debug("sending command to client", "command", cmd, "bytes", cmd.EncodeToBytes())
		command.Connection.Send(cmd.EncodeToBytes())
		return
	}

	if g.MatchMaking.HasPlayer(string(command.Data)) {
		g.log.Debug("duplicate username", "username", string(command.Data))
		cmd := tcp.TcpCommand{
			Connection: command.Connection,
			Type: tcp.CommandType.DuplicateUsername,
			Data: make([]byte, 0),
		}
		command.Connection.Send(cmd.EncodeToBytes())
		return
	}

	player := Player{
		username: string(command.Data),
		connection: command.Connection,
		ships: make([]Ship, 0),
	}

	g.log.Info("player put into matchmaking", "player", player.username)

	g.MatchMaking.mutex.Lock()
	g.MatchMaking.Players[&player]=true
	g.MatchMaking.mutex.Unlock()

	room := g.MatchMaking.SetupGame()
	if room != nil {
		g.MatchMaking.mutex.Lock()
		delete(g.MatchMaking.Players, room.player1)
		delete(g.MatchMaking.Players, room.player2)
		g.MatchMaking.mutex.Unlock()
		g.log.Debug("deleted players from matchmaking", "player1", room.player1.username, "player2", room.player2.username, "mm len", len(g.MatchMaking.Players))

		g.mutex.Lock()
		g.Rooms[room] = true
		g.mutex.Unlock()

		g.log.Info("new room set", "player1", room.player1.username, "player2", room.player2.username)

		go room.Loop()
	}
}