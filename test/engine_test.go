package test

import (
    "testing"
    "chess-engine/internal/board"
    "chess-engine/internal/engine"
    "chess-engine/internal/moves"
)

func TestEvaluate(t *testing.T) {
    b := board.NewBoard()
    score := engine.Evaluate(b)
    if score == 0 {
        t.Error("Expected non-zero score for initial position")
    }
}

func TestGenerateMoves(t *testing.T) {
    b := board.NewBoard()
    moves := engine.GenerateMoves(b)
    if len(moves) == 0 {
        t.Error("Expected moves to be generated for initial position")
    }
}

func TestSearch(t *testing.T) {
    b := board.NewBoard()
    bestMove := engine.Search(b, 3)
    if bestMove == (moves.Move{}) {
        t.Error("Expected a valid move from search")
    }
}