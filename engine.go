package gostockfish

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// UCIMoveRegex describes the regular expression for UCI moves
const UCIMoveRegex string = `[a-h]\d[a-h]\d[qrnb]?`

// PVRegex describe the regular expression for PV
var PVRegex string = fmt.Sprintf(" pv (?P<move_list>%s( %s)*)", UCIMoveRegex, UCIMoveRegex)

// Engine is the chess engine with a UCI compatible interface (e.g. stockfish)
type Engine struct {
	Executable string
	Stdin      *io.WriteCloser
	Stdout     *bufio.Reader
	Depth      int
	Ponder     bool
	Param      map[string]string
}

// BestMove contains info on the next best move
type BestMove struct {
	Move   string
	Ponder string
	Info   *Info
}

// Info describes a stockfish evaluation output
type Info struct {
	Depth    int
	Seldepth int
	Multipv  int
	Score    Score
	Nodes    int
	Nps      int
	Tbhits   int
	Time     int
	Pv       string
}

// Score describes the score of an evaluation
type Score struct {
	Eval  string
	Value int
}

// NewEngine initiates the Stockfish chess engine with Ponder set to false.
// 'param' allows parameters to be specified by a map with 'Name' and 'value'
// with value as strings.
// i.e. the following explicitly sets the default parameters
// {
//     "Contempt": 0,
//     "Threads": 1,
//     "Hash": 16,
//     "MultiPV": 1,
//     "Skill Level": 20,
//     "Move Overhead": 30,
//     "Slow Mover": 80,
// }
func NewEngine() (*Engine, error) {
	return NewEngineWithAllOptions("stockfish", 2, false, map[string]string{}, false, -10, 10)
}

// NewEngineWithDepth initiates the Stockfish chess engine with the given depth
func NewEngineWithDepth(depth int) (*Engine, error) {
	return NewEngineWithAllOptions("stockfish", depth, false, map[string]string{}, false, -10, 10)
}

// NewEngineWithAllOptions initiates the Stockfish chess engine
// If 'random' is set to false, any options not explicitly set will be set to the default
// value.
// -----
// USING RANDOM PARAMETERS
// -----
// If you set 'random' to true, the 'Contempt' parameter will be set to a random value between
// 'randMin' and 'randMax' so that you may run automated matches against slightly different
// engines.
func NewEngineWithAllOptions(stockfishExecutable string, depth int, ponder bool, param map[string]string, random bool, randMin int, randMax int) (*Engine, error) {
	engine := &Engine{
		Executable: stockfishExecutable,
		Depth:      depth,
		Ponder:     ponder,
		Param:      param,
	}

	cmd := exec.Command(stockfishExecutable)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	engine.Stdin = &stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd.Start()

	engine.Stdout = bufio.NewReader(stdout)

	engine.Put("uci")

	if !ponder {
		engine.SetOption("Ponder", "false")
	}

	baseParam := map[string]string{
		"Contempt":      "0",
		"Threads":       "1",
		"Hash":          "16",
		"MultiPV":       "1",
		"Skill Level":   "20",
		"Move Overhead": "30",
		"Slow Mover":    "80",
		"UCI_Chess960":  "false",
	}

	if random {
		baseParam["Contempt"] = strconv.Itoa(rand.Intn(randMax-randMin) + randMin)
	}

	for name, value := range param {
		baseParam[name] = value
	}
	engine.Param = baseParam

	for name, value := range engine.Param {
		err = engine.SetOption(name, value)
		if err != nil {
			return nil, err
		}
	}

	return engine, nil
}

// Put command to chess engine
func (engine *Engine) Put(command string) {
	io.WriteString(*engine.Stdin, command+"\n")
}

// SetOption sets an engine option
func (engine *Engine) SetOption(optionName string, value string) error {
	engine.Put(fmt.Sprintf("setoption name %s value %s", optionName, value))
	return engine.IsReady()
}

// IsReady is used to synchronize the golang engine object with the back-end engine. Sends 'isready' and waits for 'readyok.'
func (engine *Engine) IsReady() error {
	engine.Put("isready")
	for {
		text, _, err := engine.Stdout.ReadLine()
		if err != nil {
			return err
		}
		line := strings.TrimSpace(string(text))
		if strings.Contains(line, "No such option:") {
			return errors.New(line)
		} else if strings.Contains(line, "Unknown command:") {
			return errors.New(line)
		}
		if line == "readyok" {
			return nil
		}
	}
}

// NewGame calls 'ucinewgame' - this should be run before a new game
func (engine *Engine) NewGame() error {
	engine.Put("ucinewgame")
	return engine.IsReady()
}

// SetPosition sets start position to list of moves (i.e. ['e2e4', 'e7e5', ...]).  Moves must be in full algebraic notation.
func (engine *Engine) SetPosition(moves []string) error {
	engine.Put(fmt.Sprintf("position startpos moves %s", strings.Join(moves, " ")))
	return engine.IsReady()
}

// SetFENPosition sets start position in FEN notation. Input is a FEN string i.e. "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
func (engine *Engine) SetFENPosition(fen string) error {
	engine.Put(fmt.Sprintf("position fen %s", fen))
	return engine.IsReady()
}

// Go starts calculating on the current position
func (engine *Engine) Go() error {
	engine.Put(fmt.Sprintf("go depth %s", strconv.Itoa(engine.Depth)))
	return engine.IsReady()
}

// BestMove gets the proposed best move for current position.
func (engine *Engine) BestMove() (*BestMove, error) {
	var lastInfo *Info

	engine.Go()

	for {
		text, _, err := engine.Stdout.ReadLine()
		if err != nil {
			return nil, err
		}
		line := strings.TrimSpace(string(text))
		splitText := strings.Split(line, " ")
		if splitText[0] == "info" {
			lastInfo, err = ParseInfo(line)
			if err != nil {
				return nil, err
			}
		}
		if splitText[0] == "bestmove" {
			bestMove, err := ParseBestMove(line)
			if err != nil {
				return nil, err
			}
			bestMove.Info = lastInfo
			return bestMove, nil
		}
	}
}

// ParseInfo parses stockfish evaluation output
//
// Examples of input:
// "info depth 2 seldepth 3 multipv 1 score cp -656 nodes 43 nps 43000 tbhits 0 time 1 pv g7g6 h3g3 g6f7"
// "info depth 10 seldepth 12 multipv 1 score mate 5 nodes 2378 nps 1189000 tbhits 0 time 2 pv h3g3 g6f7 g3c7 b5d7 d1d7 f7g6 c7g3 g6h5 e6f4"
func ParseInfo(line string) (*Info, error) {
	var err error
	result := &Info{}

	info := regexp.MustCompile("info string [^ ]+ evaluation using [^ ]+ enabled")
	matches := info.FindAllStringSubmatch(line, -1)
	if matches != nil {
		return result, nil
	}

	pv := regexp.MustCompile(PVRegex)
	matches = pv.FindAllStringSubmatch(line, -1)
	if matches == nil {
		return nil, fmt.Errorf("Could not parse pv: %s", line)
	}
	result.Pv = matches[0][1]

	// Example values:
	// score cp -100        <- engine is behind 100 centipawns
	// score mate 3         <- engine has big lead or checkmated opponent
	score := regexp.MustCompile(`score (?P<eval>\w+) (?P<value>-?\d+)`)
	matches = score.FindAllStringSubmatch(line, -1)
	if matches == nil {
		return nil, fmt.Errorf("Could not parse score: %s", line)
	}
	result.Score.Eval = matches[0][1]
	result.Score.Value, err = strconv.Atoi(matches[0][2])
	if err != nil {
		return nil, err
	}

	singleValueFields := []string{"depth", "seldepth", "multipv", "nodes", "nps", "tbhits", "time"}
	for _, field := range singleValueFields {
		search := regexp.MustCompile(field + ` (?P<value>\d+)`)
		matches = search.FindAllStringSubmatch(line, -1)
		if matches == nil {
			return nil, fmt.Errorf("Could not parse %s: %s", field, line)
		}
		value, err := strconv.Atoi(matches[0][1])
		if err != nil {
			return nil, err
		}
		if field == "depth" {
			result.Depth = value
		} else if field == "seldepth" {
			result.Seldepth = value
		} else if field == "multipv" {
			result.Multipv = value
		} else if field == "nodes" {
			result.Nodes = value
		} else if field == "nps" {
			result.Nps = value
		} else if field == "tbhits" {
			result.Tbhits = value
		} else if field == "time" {
			result.Time = value
		}
	}

	return result, nil
}

// ParseBestMove parses stockfish bestmove output
//
// Examples of input:
// "bestmove d2d4 ponder a7a6"
//
func ParseBestMove(line string) (*BestMove, error) {
	var ponder string

	splitText := strings.Split(line, " ")

	if len(splitText) < 3 {
		ponder = ""
	} else {
		ponder = splitText[3]
	}

	return &BestMove{
		Move:   splitText[1],
		Ponder: ponder,
	}, nil
}
