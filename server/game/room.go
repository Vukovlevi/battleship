package game

import (
	"encoding/binary"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
	correctlyClosed byte = 0x00
	playerLeftClosed byte = 0x80

	winner byte = 0x00
	loser byte = 0x40

    waitingForShips = "waitingForShips"

    notYourTurn byte = 0
    invalidSpot byte = 1
    miss byte = 2
    hitByte byte = 3
    shiftGuessConfirmBy = 6
    shiftGuessSinkBy = 5
)

type GameRoom struct {
	log *logger.Logger
	player1     *Player
	player2     *Player
	MessageChan chan tcp.TcpCommand
	closeChan chan *GameRoom
    state string
    code string
}

func newGameRoom(log *logger.Logger) *GameRoom {
	gameRoom := new(GameRoom)

	gameRoom.log = log
	gameRoom.MessageChan = make(chan tcp.TcpCommand)

	return gameRoom
}

func (r *GameRoom) IsFull() bool {
    return r.player1 != nil && r.player2 != nil
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

func (r *GameRoom) GetStatsByte(closer, sendingTo, otherPlayer *Player) []byte { //set closer to nil if the room is closed because of win, otherPlayer is only set if the game is over by win, otherwise nil
	if closer != nil { //this means the game is not over correctly, a player has closed the connection
		assert.Assert(closer != sendingTo, "sending stats to the player closing the connection is not possible", "closer", closer.username, "sendintTo", sendingTo.username)

		firstByte := playerLeftClosed | loser
		return []byte{firstByte, 0}
	}

    //the game is closed because of someone has won
    firstByte := correctlyClosed //setting bytes according to protocol spec
    remainingShipsOfThisPlayer, _ := sendingTo.RemainingHealth()
    if remainingShipsOfThisPlayer == 0 { //setting winner/loser according to protocol spec
        firstByte = firstByte | loser
    } else {
        firstByte = firstByte | winner
    }

    remainingShipsOfOtherPlayer, remainingHealth := otherPlayer.RemainingHealth() //getting the enemies stats back
    firstByte = firstByte | (remainingShipsOfOtherPlayer << 3)

	return []byte{firstByte, remainingHealth}
}

func (r *GameRoom) GetPlayers(command tcp.TcpCommand) (*Player, *Player) { //the first user is whose the connection is, the second is the other
    player := r.player1
    otherPlayer := r.player2
    if r.player2.connection == command.Connection {
        player = r.player2
        otherPlayer = r.player1
    }

    return player, otherPlayer
}

func (r *GameRoom) HandleConnectionClosed(command *tcp.TcpCommand) *tcp.TcpCommand { //this function should only be called if the client closed connection, not when closing is because of win
    closer, sendTo := r.GetPlayers(*command)
	sendTo.connection.GameOver = true

	r.log.Debug("gameroom closing", "close initiated by", closer.username)

	cmd := tcp.TcpCommand{
		Connection: sendTo.connection,
		Type: tcp.CommandType.GameOver,
		Data: r.GetStatsByte(closer, sendTo, nil),
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

func (r *GameRoom) HandleShipsReady(command tcp.TcpCommand) {
    player, otherPlayer := r.GetPlayers(command)

    if r.state != waitingForShips { //if we dont expect ships we send back an error
        cmd := tcp.CommandTypeMismatchCommand
        cmd.Connection = command.Connection
        player.connection.Send(cmd.EncodeToBytes())
        r.log.Debug("unexpected ships received", "player", player.username)
        return
    }

    if len(player.ships) != 0 { //if the player already has ships send error back
        cmd := tcp.DataMismatchCommand
        cmd.Connection = player.connection
        player.connection.Send(cmd.EncodeToBytes())
        return
    }

    err := player.SetShips(command.Data, r.log) //this checks every possible data mismatch that can happen while parsing ships (overlapping, out of bounds positions, ship positions not being besides each other)
    if err != nil { //if there is an error, send it to the client
        tcpError, ok := err.(tcp.TcpError)
        assert.Assert(ok, "parsing ships should only return tcpErrors", "got err", err)
        r.log.Warning("parsing ship returned error", "err", tcpError.Error())
        r.log.Debug("sending error message to client from parsing ships", "cmd", tcpError.Command.EncodeToBytes())
        player.connection.Send(tcpError.Command.EncodeToBytes())
        return
    }

    if len(otherPlayer.ships) == 0 { //if he is the first one
        //inform the other player about current player readiness
        cmd := tcp.TcpCommand{
            Connection: otherPlayer.connection,
            Type: tcp.CommandType.PlayerReady,
            Data: make([]byte, 0),
        }
        otherPlayer.connection.Send(cmd.EncodeToBytes())
    } else { //if the player is the second one to send ships
        r.state = otherPlayer.username //set the state to the other user's name, since he starts
        cmd := tcp.TcpCommand{
            Type: tcp.CommandType.MatchStart,
            Data: []byte{0},
            Connection: otherPlayer.connection,
        }
        otherPlayer.connection.Send(cmd.EncodeToBytes()) //infrom the players about the match starting
        cmd.Connection = player.connection
        cmd.Data = []byte{1}
        player.connection.Send(cmd.EncodeToBytes())
    }

    //TODO: the 30s countdown limit
}

func (r *GameRoom) HandlePlayerGuess(command tcp.TcpCommand) {
    player, otherPlayer := r.GetPlayers(command)

    if r.state != player.username { //if its not the players turn send back an error message
        cmd := tcp.TcpCommand{
            Type: tcp.CommandType.GuessConfirm,
            Data: []byte{notYourTurn << shiftGuessConfirmBy},
            Connection: player.connection,
        }
        player.connection.Send(cmd.EncodeToBytes())
        return
    }

    spot := int(binary.BigEndian.Uint16(command.Data)) //get the spot the player has guessed
    r.log.Debug("got spot", "spot", spot)
    if err := validPositionBounds(spot, r.log); err != nil { //check if the spot is inbounds
        cmd := tcp.TcpCommand{
            Type: tcp.CommandType.GuessConfirm,
            Data: []byte{invalidSpot << shiftGuessConfirmBy},
            Connection: player.connection,
        }
        player.connection.Send(cmd.EncodeToBytes())
        return
    }

    if !player.CanGuessSpot(spot) { //if he cannot guess that spot return error
        cmd := tcp.TcpCommand{
            Type: tcp.CommandType.GuessConfirm,
            Data: []byte{invalidSpot << shiftGuessConfirmBy},
            Connection: player.connection,
        }
        player.connection.Send(cmd.EncodeToBytes())
        return
    }

    hit, sink, sunkenShip := otherPlayer.IsHit(spot) //test the hit
    remainingShips, _ := otherPlayer.RemainingHealth() //check for a winner, if there is one, handle gameover event instead of guess confirm and player guess to the other player
    r.log.Debug("handling player guess", "player", player.username, "spot", spot, "isHit", hit, "didSink", sink, "enemy remaining ships", remainingShips)
    if remainingShips == 0 {
        r.HandleGameOver(player, otherPlayer)
        return
    }

    //send player guess to other player for displaying reasons if there is no game over
    otherPlayer.connection.Send(command.EncodeToBytes())

    //calculate guess confirm for player if there is no game over
    cmd := tcp.TcpCommand{
        Type: tcp.CommandType.GuessConfirm,
        Connection: player.connection,
    }

    if hit {
        cmd.Data = []byte{hitByte << shiftGuessConfirmBy}
    } else {
        cmd.Data = []byte{miss << shiftGuessConfirmBy}
    }

    // 00100000 -- sink shifted
    // 11000000 -- hit shifted
    // 11100000 -- result of or
    cmd.Data[0] = cmd.Data[0] | (sink << shiftGuessSinkBy) //setting the data according to protocol specification
    if sunkenShip != nil {
        cmd.Data = append(cmd.Data, sunkenShip.GetPositionsInBytes()...)
    }
    player.connection.Send(cmd.EncodeToBytes()) //send guess confirm to player

    player.cannotGuessHereSpots[spot] = true //mark spot as unguessable
    r.state = otherPlayer.username //set the state for the other user's turn
}

func (r *GameRoom) HandleGameOver(winner *Player, loser *Player) {
    data := r.GetStatsByte(nil, winner, loser) //get stats for the winner (nil indicating that the game is over because of a win)
    cmd := tcp.TcpCommand{
        Connection: winner.connection,
        Type: tcp.CommandType.GameOver,
        Data: data,
    }
    winner.connection.Send(cmd.EncodeToBytes())

    data = r.GetStatsByte(nil, loser, winner) //get stats for the loser (nil indicating that the game is over because of a win)
    cmd.Connection = loser.connection
    cmd.Data = data
    loser.connection.Send(cmd.EncodeToBytes())

    winner.connection.GameOver = true
    loser.connection.GameOver = true

    r.CloseRoom(nil)
}

func (r *GameRoom) Loop() {
    r.state = waitingForShips
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
        case tcp.CommandType.ShipsReady:
            r.HandleShipsReady(command)
        case tcp.CommandType.PlayerGuess:
            r.HandlePlayerGuess(command)
        default: //any other command type is unexpected
            cmd := tcp.CommandTypeMismatchCommand
            cmd.Connection = command.Connection
            command.Connection.Send(cmd.EncodeToBytes())
		}

		r.log.Debug("gameroom got command", "command", command)
	}
}
