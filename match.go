package main

import "math/rand"

// MaxMoves is the maximum number of move in the play
const MaxMoves int = 500

// Match represents a match between two engines
type Match struct {
	White        string
	WhiteEngine  *Engine
	Black        string
	BlackEngine  *Engine
	Moves        []string
	Winner       string
	WinnerEngine *Engine
}

// NewMatch setups a chess match between two specified engines. The white player
// is randomly chosen.
//
// deepEngine := NewEngineWithDepth(20)
// shallowEngine := NewEngineWithDepth(10)
//
// m := NewMatch("deep", deepEngine, "shallow", shallowEngine)
func NewMatch(e1 string, engine1 *Engine, e2 string, engine2 *Engine) (*Match, error) {
	var m *Match

	if rand.Int()%2 == 0 {
		m = &Match{
			White:       e1,
			WhiteEngine: engine1,
			Black:       e2,
			BlackEngine: engine2,
		}
	} else {
		m = &Match{
			White:       e2,
			WhiteEngine: engine2,
			Black:       e1,
			BlackEngine: engine1,
		}
	}

	err := engine1.NewGame()
	if err != nil {
		return nil, err
	}
	err = engine2.NewGame()
	if err != nil {
		return nil, err
	}

	m.Winner = ""
	m.WinnerEngine = nil

	return m, nil
}

// Move advances the game by single move, if possible. Returns a bool on whether the move was performed.
func (match *Match) Move() (bool, error) {
	var activeEngine *Engine
	var activeEngineName string
	var inactiveEngine *Engine
	var inactiveEngineName string

	if len(match.Moves) == MaxMoves {
		return false, nil
	} else if len(match.Moves)%2 != 0 {
		activeEngine = match.BlackEngine
		activeEngineName = match.Black
		inactiveEngine = match.WhiteEngine
		inactiveEngineName = match.White
	} else {
		activeEngine = match.WhiteEngine
		activeEngineName = match.White
		inactiveEngine = match.BlackEngine
		inactiveEngineName = match.Black
	}
	activeEngine.SetPosition(match.Moves)
	bestMove, err := activeEngine.BestMove()
	if err != nil {
		return false, err
	}
	match.Moves = append(match.Moves, bestMove.Move)

	if bestMove.Info.Score.Eval == "mate" {
		matenum := bestMove.Info.Score.Value
		if matenum > 0 {
			match.WinnerEngine = activeEngine
			match.Winner = activeEngineName
		} else if matenum < 0 {
			match.WinnerEngine = inactiveEngine
			match.Winner = inactiveEngineName
		}
		return false, nil
	}

	if bestMove.Ponder != "(none)" {
		return true, nil
	}

	return false, nil
}

// Run plays the game until completion or 200 moves have been played,
// returning the winning engine name. Returns empty string if there
// is a draw.
func (match *Match) Run() (string, error) {
	for {
		move, err := match.Move()
		if err != nil {
			return "", err
		}
		if !move {
			break
		}
	}
	return match.Winner, nil
}
