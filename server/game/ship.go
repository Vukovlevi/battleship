package game

import (
	"encoding/binary"
	"slices"

	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

const (
    SPOT_LEN_SIZE = 1
)

type Ship struct {
	health    int
	positions map[int]bool
}

func parseSingleShip(data []byte, log *logger.Logger) (Ship, error) { //the data contains the informations of spots and only spots
    if len(data) % 2 != 0 {
        log.Warning("the len of spots should always be even because one spot is 2 bytes long", "len spots", len(data))
        cmd := tcp.DataMismatchCommand
        err := tcp.CreateTcpError("the len of spots should always be even because one spot is 2 bytes long", cmd)
        return Ship{}, err
    }

    ship := Ship{
        health: len(data) / 2,
        positions: make(map[int]bool),
    }

    for i := 0; i < len(data); i += 2 {
        pos := binary.BigEndian.Uint16(data[i:i+2])
        if err := validPositionBounds(int(pos), log); err != nil {
            return Ship{}, err
        }
        ship.positions[int(pos)] = true
    }

    return ship, nil
}

func validPositionBounds(spot int, log *logger.Logger) error {
    x := spot / 1000
    y := spot % 1000
    if x > 10 || y > 10 || x < 0 || y < 0 { //check the bounds of a single positions, return error if pos is out of bounds
        log.Warning("coords at max should be 10 and at min 1", "x", x, "y", y)
        cmd := tcp.DataMismatchCommand
        err := tcp.CreateTcpError("coords at max should be 10 and at min 1", cmd)
        return err
    }

    return nil
}

func checkValidShipPositions(ship Ship, log *logger.Logger) error {
    stepSize := 1 //assuming that the ship is positioned horizontal
    positions := make([]int, 0)
    for pos := range ship.positions {
        positions = append(positions, pos)
    }
    slices.Sort(positions)

    if positions[0] + stepSize != positions[1] { //change positions to vertical if not horizontal
        stepSize = 1000
    }

    if positions[0] + stepSize != positions[1] { //if the ship is not horizontal and not vertical, return error
        log.Warning("the ship's positions are not beside each other", "pos1", positions[0], "pos2", positions[1])
        cmd := tcp.DataMismatchCommand
        err := tcp.CreateTcpError("the ship's positions are not beside each other", cmd)
        return err
    }

    for i := 1; i < len(positions) - 1; i++ {
        if positions[i] + stepSize != positions[i + 1] { //check the rest of the positions to make sure they are beside each other, return error if not
            log.Warning("the ship's positions are not beside each other", "pos1", positions[i], "pos2", positions[i + 1])
            cmd := tcp.DataMismatchCommand
            err := tcp.CreateTcpError("the ship's positions are not beside each other", cmd)
            return err
        }
    }

    return nil
}

func parseShips(data []byte, log *logger.Logger) ([]Ship, error) { //the data is the entire received data with all the ship informations (spots len + spots)
    ships := make([]Ship, 0)
    idx := 0
    for idx < len(data) {

        spotsLen := int(data[idx]) //get how many spots does the current ship have

        if idx + SPOT_LEN_SIZE + spotsLen > len(data) {
            log.Warning("spots len is out of bounds", "max len", len(data), "tried len", idx + SPOT_LEN_SIZE + spotsLen)
            cmd := tcp.DataMismatchCommand
            err := tcp.CreateTcpError("spots len is out of bounds", cmd)
            return []Ship{}, err
        }

        ship, err := parseSingleShip(data[idx + SPOT_LEN_SIZE: idx + SPOT_LEN_SIZE + spotsLen], log) //parse ship with that many spots

        if err != nil {
            return []Ship{}, err
        }

        if len(ship.positions) < 2 || len(ship.positions) > 5 { //check if a single ship's len is inbound, return error if not
            log.Warning("a single ship's length should be at least 2 and at most 5", "ship len", len(ship.positions))
            cmd := tcp.DataMismatchCommand
            err := tcp.CreateTcpError("a single ship's length should be at least 2 and at most 5", cmd)
            return []Ship{}, err
        }

        err = checkValidShipPositions(ship, log)
        if err != nil {
            return []Ship{}, err
        }

        ships = append(ships, ship)
        idx += SPOT_LEN_SIZE + spotsLen //update the index to the next spotlen byte (next ship info)
    }
    return ships, nil
}
