package gostockfish

import "testing"

func TestParseInfo(t *testing.T) {
	var tests = []struct {
		input    string
		expected *Info
	}{
		{
			"info depth 2 seldepth 3 multipv 1 score cp -656 nodes 43 nps 43000 tbhits 0 time 1 pv g7g6 h3g3 g6f7",
			&Info{
				Depth:    2,
				Seldepth: 3,
				Multipv:  1,
				Score: Score{
					Eval:  "cp",
					Value: -656,
				},
				Nodes:  43,
				Nps:    43000,
				Tbhits: 0,
				Time:   1,
				Pv:     "g7g6 h3g3 g6f7",
			},
		},
		{
			"info depth 10 seldepth 12 multipv 1 score mate 5 nodes 2378 nps 1189000 tbhits 0 time 2 pv h3g3 g6f7 g3c7 b5d7 d1d7 f7g6 c7g3 g6h5 e6f4",
			&Info{
				Depth:    10,
				Seldepth: 12,
				Multipv:  1,
				Score: Score{
					Eval:  "mate",
					Value: 5,
				},
				Nodes:  2378,
				Nps:    1189000,
				Tbhits: 0,
				Time:   2,
				Pv:     "h3g3 g6f7 g3c7 b5d7 d1d7 f7g6 c7g3 g6h5 e6f4",
			},
		},
		{
			"info string NNUE evaluation using nn-82215d0fd0df.nnue enabled",
			&Info{},
		},
	}
	for _, tt := range tests {
		actual, err := ParseInfo(tt.input)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if *actual != *tt.expected {
			t.Errorf("ParseInfo(\"%s\"): expected %v, actual %v", tt.input, tt.expected, actual)
		}
	}
}

func TestBestMove(t *testing.T) {
	var tests = []struct {
		input    string
		expected *BestMove
	}{
		{
			"bestmove d2d4 ponder a7a6",
			&BestMove{
				Move:   "d2d4",
				Ponder: "a7a6",
			},
		},
		{
			"bestmove d2d4",
			&BestMove{
				Move:   "d2d4",
				Ponder: "",
			},
		},
	}
	for _, tt := range tests {
		actual, err := ParseBestMove(tt.input)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if *actual != *tt.expected {
			t.Errorf("ParseBestMove(\"%s\"): expected %v, actual %v", tt.input, tt.expected, actual)
		}
	}
}
