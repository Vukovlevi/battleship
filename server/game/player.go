package game

import (
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

    p.ships = ships
    return nil
}
