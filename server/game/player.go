package game

import (
	"encoding/binary"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
    SHIP_COUNT = 5
)

type Player struct {
	username   string
	connection *tcp.Connection
	ships []Ship
    cannotGuessHereSpots map[int]bool
    //temporaryFile io.WriteCloser //uncomment in case of debugging
}

func getAllShipMap() map[int]int { //returns a map where key is the length of the ship and value is how many ships should with that length
    allShipMap := make(map[int]int)
    allShipMap[2] = 1
    allShipMap[3] = 2
    allShipMap[4] = 1
    allShipMap[5] = 1

    sum := 0
    for _, count := range allShipMap {
        sum += count
    }

    assert.Assert(sum == SHIP_COUNT, "there should be as many ships in the counting map than expected", "expected", SHIP_COUNT, "got", sum)

    return allShipMap
}

func (p *Player) CanGuessSpot(spot int) bool {
    _, ok := p.cannotGuessHereSpots[spot]
    return !ok
}

func (p *Player) IsHit(spot int) (bool, byte, *Ship) { //returns if the spot is a hit and wether the hit sink the ship
    for i, _ := range p.ships {
        if _, ok := p.ships[i].positions[spot]; ok {
            p.ships[i].health--
            var sink byte = 0

            //i keep it here just in case i need to debug this again
            /*
            p.temporaryFile.Write([]byte(fmt.Sprintf("spot: %d, ship: %+v\nships: %+v\n", spot, p.ships[i], p.ships)))
            if spot == 10010 {
                p.temporaryFile.Close()
            }
            */
            //end of debugging code

            if p.ships[i].health == 0 {
                sink = 1
            }
            return true, sink, &p.ships[i]
        }
    }
    return false, 0, nil
}

func (p *Player) RemainingHealth() (byte, byte) { //returns the remaining ships and the remaining health of the player
    var remainingShips byte = 0
    var remainingHealth byte = 0

    for i, _ := range p.ships {
        remainingHealth += byte(p.ships[i].health)
        if p.ships[i].health != 0 {
            remainingShips++
        }
    }

    return remainingShips, remainingHealth
}

func (p *Player) SetShips(data []byte, log *logger.Logger) error { //the data is the entire received data with all the ships info
    ships, err := parseShips(data, log)
    if err != nil {
        return err
    }

    if len(ships) != SHIP_COUNT {
        log.Warning("there are not as many ships as expected by the server", "expected", SHIP_COUNT, "got", len(ships))
        cmd := tcp.DataMismatchCommand
        err := tcp.CreateTcpError("there are not as many ships as expected by the server", cmd)
        return err
    }

    positions := make(map[int]bool)
    shipMap := getAllShipMap()

    for _, ship := range ships {
        count, ok := shipMap[ship.health]
        if !ok {
            log.Warning("there is more ship with that length than expected by the server", "ship len", ship.health)
            cmd := tcp.DataMismatchCommand
            err := tcp.CreateTcpError("there is more ship with that length than expected by the server", cmd)
            return err
        }

        count--
        if count == 0 {
            delete(shipMap, ship.health)
        } else {
            shipMap[ship.health] = count
        }

        for pos := range ship.positions {
            _, ok = positions[pos]
            if ok {
                log.Warning("this position is already taken, overlapping the ships is not possible", "pos", pos, "ship len", ship.health)
                cmd := tcp.DataMismatchCommand
                err := tcp.CreateTcpError("this position is already taken, overlapping the ships is not possible", cmd)
                return err
            }

            positions[pos] = true
        }
    }

    assert.Assert(len(shipMap) == 0, "every ship len should be deleted with it's count from shipmap if there is no error", "remaining map", shipMap)

    //p.temporaryFile, _ = os.Create("dikaz") //DELETE THIS AFTER TEST IS GOOD
    p.ships = ships
    log.Debug("players ships", "player", p.username, "ships", p.ships)
    return nil
}

func (p *Player) GetRemainingSpots(enemyCannotGuessSpots map[int]bool) []byte { //returns the remaining spots in case of a win for the loser to display
    positions := make([]byte, 0)
    for _, ship := range p.ships {
        for pos, _ := range ship.positions {
            if _, ok := enemyCannotGuessSpots[pos]; !ok {
                positions = binary.BigEndian.AppendUint16(positions, uint16(pos))
            }
        }
    }
    return positions
}
