package game

import "github.com/vukovlevi/battleship/server/tcp"

type Player struct {
	username   string
	connection *tcp.Connection
	ships []Ship
}