package test

import (
	"testing"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/engine"
)

func TestEvaluate(t *testing.T) {
	b := board.NewBoard()
	score := engine.Evaluate(b)
	// Initial position should evaluate to 0 (equal material and position)
	if score != 0 {
		t.Errorf("Expected zero score for initial position, got %d", score)
	}
}

func TestGenerateMoves(t *testing.T) {
	b := board.NewBoard()
	moves := engine.GenerateMoves(b)
	if len(moves) == 0 {
		t.Error("Expected moves to be generated for initial position")
	}
}
