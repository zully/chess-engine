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
	EnPassant bool   // Whether this is an en passant capture
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

	// Handle pawn moves (e.g., "e4", "exd5", "a1=Q", "exd8=Q")
	if len(notation) >= 2 && !isUpperCase(notation[0]) {
		file := notation[0]

		// Check for promotion notation
		var promotionPiece string
		if strings.Contains(notation, "=") {
			parts := strings.Split(notation, "=")
			notation = parts[0]
			if len(parts) > 1 && len(parts[1]) > 0 {
				promotionPiece = string(parts[1][0])
			}
		} else if len(notation) >= 3 && isUpperCase(notation[len(notation)-1]) {
			// Handle promotion without = (e.g., "a1Q")
			promotionPiece = string(notation[len(notation)-1])
			notation = notation[:len(notation)-1]
		}

		if len(notation) == 2 { // Simple pawn move (e.g., "e4", "a1=Q")
			rank := notation[1]
			move.To = string(file) + string(rank)
			// For pawn moves, let the board find the right starting square
			move.From = ""
		} else if len(notation) == 4 && notation[1] == 'x' { // Pawn capture (e.g., "exd5", "exd8=Q")
			move.Capture = true
			move.To = string(notation[2]) + string(notation[3])
			// For pawn captures, specify the file but let the board find the rank
			move.From = string(file) + "*"
		} else if len(notation) >= 5 && notation[1] == 'x' && len(promotionPiece) > 0 { // Pawn capture with promotion (e.g., "exd8=Q")
			move.Capture = true
			move.To = string(notation[2]) + string(notation[3])
			move.From = string(file) + "*"
		}

		move.Piece = "P"
		if promotionPiece != "" {
			move.Promote = promotionPiece
		}
		return move, nil
	}

	// Handle piece moves (e.g., "Nf3", "Bxe4", "Rae8", "R1e8")
	if len(notation) >= 3 {
		move.Piece = string(notation[0])
		idx := 1

		// Handle disambiguation (e.g., "Ra" in "Rae8" or "1" in "R1e8")
		var disambiguation string

		// Check if there's disambiguation before the capture or target square
		if idx < len(notation) {
			char := notation[idx]
			// Check if it's a file letter (a-h) or rank number (1-8)
			if (char >= 'a' && char <= 'h') || (char >= '1' && char <= '8') {
				// Check if the next character is 'x' or if we're near the end
				if idx+1 < len(notation) && notation[idx+1] == 'x' {
					// Disambiguation followed by capture (e.g., "Raxe8")
					disambiguation = string(char)
					idx++
				} else if idx+2 < len(notation) {
					// Check if this looks like disambiguation (e.g., "Rae8" where 'a' is disambiguation)
					nextChar := notation[idx+1]
					if (nextChar >= 'a' && nextChar <= 'h') && idx+3 < len(notation) {
						// This looks like disambiguation + target square (e.g., "Rae8")
						disambiguation = string(char)
						idx++
					} else if (nextChar >= '1' && nextChar <= '8') && idx+2 == len(notation)-1 {
						// This looks like disambiguation + target square (e.g., "R1e8")
						disambiguation = string(char)
						idx++
					}
				}
			}
		}

		// Handle captures
		if idx < len(notation) && notation[idx] == 'x' {
			move.Capture = true
			idx++
		}

		// Get the target square (should be the last 2 characters)
		if idx+1 < len(notation) {
			to := notation[idx : idx+2]

			// Validate the target square notation
			if len(to) != 2 ||
				to[0] < 'a' || to[0] > 'h' ||
				to[1] < '1' || to[1] > '8' {
				return nil, fmt.Errorf("invalid target square: %q", to)
			}
			move.To = to

			// Set disambiguation in the From field if present
			if disambiguation != "" {
				move.From = disambiguation
			}

			return move, nil
		}
	}

	return nil, fmt.Errorf("invalid move notation: %s", notation)
}

func isUpperCase(c byte) bool {
	return c >= 'A' && c <= 'Z'
}
