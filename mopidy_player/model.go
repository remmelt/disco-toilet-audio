package mopidy_player

type State int

const ( // iota is reset to 0
	UNKNOWN State = iota
	PAUSED        = iota
	PLAYING       = iota
	STOPPED       = iota
)

func (s State) String() string {
	return [...]string{"UNKNOWN", "PAUSED", "PLAYING", "STOPPED"}[s]
}
