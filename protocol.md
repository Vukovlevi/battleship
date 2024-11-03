# THIS FILE CONTAINS THE PROTOCOL USED BY BOTH THE CLIENT AND THE SERVER TO COMMUNICATE

## The tcp packet:

  1 byte: version  1 byte: msg type*      2 bytes: data* length (x)
| - - - - - - - - | - - - - - - - - | - - - - - - - - | - - - - - - - - |
                             x bytes: data*
| - - - - - - - - | - - - - - - - - | - - - - - - - - | - - - - - - - - |

## Msg type*:
    - 1: join request - data: username (string in bytes format) [client -> server] => this message informs the server about a new user trying to connect, if everything is ok, the user will be put in mm*
    - 2: duplicate username - data: none [server -> client] => this message informs the client that a user with that name is already in mm*
    - 3: match found - data: opponent username (string in bytes format) [server -> client] => this message informs the client about a room* being set up with the opponent user's name
    - 4: ships ready - data: ships* [client -> server] => this message informs the server about a user being ready with the locations of the user's ships
    - 5: player ready - data: none [server -> client] => this message informs the user about the opponent' readyness (the user now has 30 seconds to finish placing his ships or else the match will be cancelled)
    - 6: match start - data: starting information* [server -> client] => this message informs the user that the match has started with the information on who has the first turn
    - 7: player guess - data: spot* [client -> server] => this message informs the server about a player's guess on opponents ship position
    - 8: guess confirm - data: spot feedback* [server -> client] => this message informs the user wether the guess was successful
    - 9: game over - data: stats* [server -> client] => this message informs the client about a game being over with the statistics of the match
    - 10: close event - data: none [server -> server, server -> client] => this message informs either the server or the client about a connection being closed, therefore every other connection and open room can be closed
    - 11: mismatch - data: mismatch type* [server -> client] => this message informs the client about a mismatch that was detected in the client's message

    - mm*: matchmaking, players are put into here after their successful join request until a room can be set up
    - room*: a room is where the players are put together to play the game, basically the room runs the actual game while the server handles connections and the gateway between mm* and room*

## Data*:

### Ships*:
    - a list of ship*

### Ship*:
1 byte: spots* length                                                      2 bytes: spot*
| - - - - - - - - |                                             | - - - - - - - - | - - - - - - - - |
(how many bytes is there that contains this ships's spots)         (spot*: [x; y] -> x * 1000 + y)
(eg.: in case of 1 spot - 2 [spots* length = spot* count * 2])     (eg.: in case of [8; 7] -> 8007)

### Starting information*:
1 byte: information
| - - - - - - - - |

    - 0: you start
    - 1: the opponent starts

### Spot feedback*:
  1 byte: feedback
| - - | - | - - - - - |
 (0-3) (0-1)

    - 0: not your turn
    - 1: invalid spot* (either your spot* or you have already guessed it)
    - 2: miss (you did not hit any of the opponent's ships)
    - 3: hit (you successfully hit on of the opponent's ship)

    - 0: the opponent's ship did not sink
    - 1: the opponent's ship did sink

### Stats*:
1 byte: game/ship info          1 byte: hit info
| - | - | - - - | - - - |      | - - - - - - - - |
(0-1)(0-1)(0-3)

    - 0: the game is over because a player has won
    - 1: the game is over because a player left, therefore there is no winner

    - 0: you won
    - 1: you lost

    - (0-3): the remaining enemy ships

    - hit info: how many hit was remaining until all of your opponent's ships sink

### Mismatch type*:
    1 byte: type
| - - - - - - - - |

    - 0: version mismatch (the package sent by the client did not match the protocol version used by the server)
    - 1: length mismatch (the package sent by the client did not match the length that itself specified)
    - 2: command type* mismatch (the command type* sent by the client was unexpected during the phase of the game, therefore could not be processed)
