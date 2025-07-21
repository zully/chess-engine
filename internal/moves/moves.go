package moves

import (
	"fmt"
	"strings"
)

// Move represents a chess move with its starting and ending positions.
type Move struct {
	From      string // Starting square (e.g., "e2")
	To        string // Ending square (e.g., "e4")
	Piece     string // The piece being moved (P, N, B, R, Q, K)
	Capture   bool   // Whether the move is a capture
	Promote   string // Promotion piece (if any)
	Castle    string // "O-O" for kingside, "O-O-O" for queenside
	Check     bool   // Whether the move gives check
	Checkmate bool   // Whether the move gives checkmate
}

// ParseAlgebraic converts a move in algebraic notation to a Move struct
func ParseAlgebraic(notation string, isWhite bool) (*Move, error) {
	move := &Move{}

	// Handle castling
	if strings.ToLower(notation) == "o-o" {
		move.Castle = "O-O"
		if isWhite {
			move.From = "e1"
			move.To = "g1"
		} else {
			move.From = "e8"
			move.To = "g8"
		}
		move.Piece = "K"
		return move, nil
	}
	if strings.ToLower(notation) == "o-o-o" {
		move.Castle = "O-O-O"
		if isWhite {
			move.From = "e1"
			move.To = "c1"
		} else {
			move.From = "e8"
			move.To = "c8"
		}
		move.Piece = "K"
		return move, nil
	}

	notation = strings.TrimRight(notation, "+#") // Remove check/mate symbols

	// Handle pawn moves (e.g., "e4", "exd5")
	if len(notation) >= 2 && !isUpperCase(notation[0]) {
		file := notation[0]
		if len(notation) == 2 { // Simple pawn move (e.g., "e4")
			rank := notation[1]
			move.To = string(file) + string(rank)
			// For pawn moves, let the board find the right starting square
			move.From = ""
		} else if len(notation) == 4 && notation[1] == 'x' { // Pawn capture (e.g., "exd5")
			move.Capture = true
			move.To = string(notation[2]) + string(notation[3])
			// For pawn captures, specify the file but let the board find the rank
			move.From = string(file) + "*"
		}
		move.Piece = "P"
		return move, nil
	}

	// Handle piece moves (e.g., "Nf3", "Bxe4")
	if len(notation) >= 3 {
		move.Piece = string(notation[0])
		idx := 1

		// Handle captures
		if strings.Contains(notation[idx:], "x") {
			move.Capture = true
			idx = strings.Index(notation, "x") + 1
		}

		// Get the target square
		to := notation[len(notation)-2:]

		// Validate the target square notation
		if len(to) != 2 ||
			to[0] < 'a' || to[0] > 'h' ||
			to[1] < '1' || to[1] > '8' {
			return nil, fmt.Errorf("invalid target square: %q", to)
		}
		move.To = to
		return move, nil
	}

	return nil, fmt.Errorf("invalid move notation: %s", notation)
}

func isUpperCase(c byte) bool {
	return c >= 'A' && c <= 'Z'
}
