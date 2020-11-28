package gostockfish

import "testing"

func TestQuickCheckmate(t *testing.T) {
	// 1. e4 e5 2. Bc4 Nc6 3. Qf3 d6
	e1, err := NewEngineWithDepth(6)
	if err != nil {
		t.Fatalf(err.Error())
	}
	e2, err := NewEngineWithDepth(6)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// match creation must be before calling SetPosition() as NewMatch() resets position
	m, err := NewMatch("e1", e1, "e2", e2)

	if err != nil {
		t.Fatalf(err.Error())
	}

	e1.SetPosition([]string{"e2e4", "e7e5", "f1c4", "b8c6", "d1f3", "d7d6"})
	e2.SetPosition([]string{"e2e4", "e7e5", "f1c4", "b8c6", "d1f3", "d7d6"})

	m.Run()

	if m.Winner != "e1" && m.Winner != "e2" {
		t.Fatalf("Expected winner \"e1\" or \"e2\", got \"%s\"", m.Winner)
	}
}
