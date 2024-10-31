package tcp

type CommandTypeEnum struct {
	JoinRequest       int
	DuplicateUsername int
	MatchFound        int
	ShipsReady        int
	PlayerReady       int
	MatchStart        int
	PlayerGuess       int
	GuessConfirm      int
	GameOver          int
}

var (
	CommandType = CommandTypeEnum{
		JoinRequest:       1,
		DuplicateUsername: 2,
		MatchFound:        3,
		ShipsReady:        4,
		PlayerReady:       5,
		MatchStart:        6,
		PlayerGuess:       7,
		GuessConfirm:      8,
		GameOver:          9,
	}
)

type TcpCommand struct {
	Connection *Connection
	Type       int
	Data       []byte
}