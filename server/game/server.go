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
	Rooms           []*GameRoom
	IncomingRequestChan chan tcp.TcpCommand 
	mutex sync.RWMutex
}

func NewGameServer(log *logger.Logger) *GameServer {
	return &GameServer{
		log: log,
		MatchMaking: MatchMaking{Players: make(map[*Player]bool), log: log},
		Rooms:       make([]*GameRoom, 0),
		IncomingRequestChan: make(chan tcp.TcpCommand),
	}
}

func (g *GameServer) Start() {
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
	if command.Type != tcp.CommandType.JoinRequest {
		g.log.Warning("command type mismatch while handling join request, closing connection", "expected type", tcp.CommandType.JoinRequest, "got type", command.Type, "connectionId", command.Connection.Id)
		//TODO: return a general error message to client
		command.Connection.Close()
		return
	}

	if g.MatchMaking.HasPlayer(string(command.Data)) {
		g.log.Debug("duplicate username", "username", string(command.Data))
		//TODO: return the correct tcp error message (DuplicateUsername)
		command.Connection.Close()
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
		g.Rooms = append(g.Rooms, room)
		g.mutex.Unlock()

		g.log.Info("new room set", "player1", room.player1.username, "player2", room.player2.username)

		go room.Loop()
	}
}