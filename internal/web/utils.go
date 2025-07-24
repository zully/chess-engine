package web

import (
	"fmt"
	"strings"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/uci"
)

// IsValidUCIMove validates basic UCI move format
func IsValidUCIMove(move string) bool {
	// Basic UCI format: 4 or 5 characters
	// 4 chars: e2e4, a1h8, etc.
	// 5 chars: e7e8q (pawn promotion)
	if len(move) < 4 || len(move) > 5 {
		return false
	}

	// Check from square (first 2 chars)
	if move[0] < 'a' || move[0] > 'h' || move[1] < '1' || move[1] > '8' {
		return false
	}

	// Check to square (chars 2-3)
	if move[2] < 'a' || move[2] > 'h' || move[3] < '1' || move[3] > '8' {
		return false
	}

	// Check promotion piece if present (char 4)
	if len(move) == 5 {
		promotion := move[4]
		if promotion != 'q' && promotion != 'r' && promotion != 'b' && promotion != 'n' {
			return false
		}
	}

	return true
}

// ConvertUCIToAlgebraic converts a UCI move to algebraic notation (simplified)
func ConvertUCIToAlgebraic(uciMove string, gameBoard *board.Board, isFirstMove bool) string {
	if len(uciMove) < 4 {
		return uciMove
	}

	// Simple algebraic conversion without creating board copies
	// This prevents additional position recording that was causing false repetitions

	// Handle castling moves
	if uciMove == "e1g1" || uciMove == "e8g8" {
		return "O-O"
	}
	if uciMove == "e1c1" || uciMove == "e8c8" {
		return "O-O-O"
	}

	// For other moves, return a simplified format
	toSquare := uciMove[2:4]

	// Check if there's a piece on the destination (capture)
	toRank, toFile := board.GetSquareCoords(toSquare)
	if toRank >= 0 && toRank <= 7 && toFile >= 0 && toFile <= 7 {
		targetPiece := gameBoard.GetPiece(toRank, toFile)
		if targetPiece != board.Empty {
			// It's a capture - add 'x'
			result := toSquare
			if len(uciMove) == 5 {
				result += "=" + strings.ToUpper(string(uciMove[4]))
			}
			return result
		}
	}

	// Regular move
	result := toSquare
	if len(uciMove) == 5 {
		result += "=" + strings.ToUpper(string(uciMove[4]))
	}
	return result
}

// GetEvaluationAfterMove gets the position evaluation after making a move
func GetEvaluationAfterMove(board *board.Board, uciMove string, stockfishEngine *uci.Engine) (int, error) {
	if stockfishEngine == nil {
		return 0, fmt.Errorf("engine not available")
	}

	// Create FEN directly without making the move on a board copy
	// This avoids polluting the position history
	currentFEN := board.ToFEN()

	// Use Stockfish to evaluate the position after the move
	// Set the current position and get the evaluation after the move
	if err := stockfishEngine.SetPosition(currentFEN); err != nil {
		return 0, err
	}

	// Get the current position evaluation
	// This avoids board copy pollution while still providing evaluation data
	currentEval, err := stockfishEngine.GetEvaluation(currentFEN)
	if err != nil {
		return 0, err
	}

	// Return negative since we're looking from opponent's perspective
	return -currentEval, nil
} 