package board

import (
	"fmt"
	"strings"
)

// Piece constants for chess pieces
const (
	Empty = iota
	WP    // White Pawn
	WN    // White Knight
	WB    // White Bishop
	WR    // White Rook
	WQ    // White Queen
	WK    // White King
	BP    // Black Pawn
	BN    // Black Knight
	BB    // Black Bishop
	BR    // Black Rook
	BQ    // Black Queen
	BK    // Black King
)

// Square represents a chess square with its algebraic notation and piece
type Square struct {
	Name  string // algebraic notation (e.g., "e4")
	Piece int    // piece occupying the square
}

// Board represents a chess board
type Board struct {
	Squares        [8][8]Square // 8x8 board with named squares
	WhiteToMove    bool         // true if it's white's turn
	CastlingRights int          // stores castling availability
	EnPassant      string       // en passant target square in algebraic notation
	HalfMoveClock  int          // counts moves since last pawn move or capture
	FullMoveNumber int          // counts full moves in the game
	MovesPlayed    []string     // list of moves in algebraic notation
}

// PieceToString converts a piece constant to its string representation
func PieceToString(piece int) string {
	switch piece {
	case Empty:
		return "  "
	case WP:
		return "WP"
	case WN:
		return "WN"
	case WB:
		return "WB"
	case WR:
		return "WR"
	case WQ:
		return "WQ"
	case WK:
		return "WK"
	case BP:
		return "BP"
	case BN:
		return "BN"
	case BB:
		return "BB"
	case BR:
		return "BR"
	case BQ:
		return "BQ"
	case BK:
		return "BK"
	default:
		return "??"
	}
}

// NewBoard creates and returns a new board in the initial chess position
func NewBoard() *Board {
	b := &Board{
		WhiteToMove:    true,
		CastlingRights: 15, // 1111 in binary - all castling available
		EnPassant:      "", // no en passant target initially
		HalfMoveClock:  0,
		FullMoveNumber: 1,
		MovesPlayed:    make([]string, 0),
	}

	// Initialize all squares with their names
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			b.Squares[rank][file].Name = GetSquareName(rank, file)
		}
	}

	// Initial piece placement
	// First rank (white pieces)
	b.Squares[7][0] = Square{Name: "a1", Piece: WR}
	b.Squares[7][1] = Square{Name: "b1", Piece: WN}
	b.Squares[7][2] = Square{Name: "c1", Piece: WB}
	b.Squares[7][3] = Square{Name: "d1", Piece: WQ}
	b.Squares[7][4] = Square{Name: "e1", Piece: WK}
	b.Squares[7][5] = Square{Name: "f1", Piece: WB}
	b.Squares[7][6] = Square{Name: "g1", Piece: WN}
	b.Squares[7][7] = Square{Name: "h1", Piece: WR}

	// Second rank (white pawns)
	for file := 0; file < 8; file++ {
		b.Squares[6][file].Piece = WP
	}

	// Empty squares (ranks 3-6)
	for rank := 2; rank < 6; rank++ {
		for file := 0; file < 8; file++ {
			b.Squares[rank][file].Piece = Empty
		}
	}

	// Seventh rank (black pawns)
	for file := 0; file < 8; file++ {
		b.Squares[1][file].Piece = BP
	}

	// Eight rank (black pieces)
	b.Squares[0][0] = Square{Name: "a8", Piece: BR}
	b.Squares[0][1] = Square{Name: "b8", Piece: BN}
	b.Squares[0][2] = Square{Name: "c8", Piece: BB}
	b.Squares[0][3] = Square{Name: "d8", Piece: BQ}
	b.Squares[0][4] = Square{Name: "e8", Piece: BK}
	b.Squares[0][5] = Square{Name: "f8", Piece: BB}
	b.Squares[0][6] = Square{Name: "g8", Piece: BN}
	b.Squares[0][7] = Square{Name: "h8", Piece: BR}

	return b
}

// GetPiece returns the piece at the given rank and file (0-7)
func (b *Board) GetPiece(rank, file int) int {
	return b.Squares[rank][file].Piece
}

// GetSquare returns the square at the given algebraic notation (e.g., "e4")
func (b *Board) GetSquare(algebraicNotation string) *Square {
	rank, file := getSquareCoords(algebraicNotation)
	if rank < 0 || file < 0 || rank > 7 || file > 7 {
		return nil
	}
	return &b.Squares[rank][file]
}

// IsSquareEmpty returns true if the given square is empty
func (b *Board) IsSquareEmpty(rank, file int) bool {
	return b.GetPiece(rank, file) == Empty
}

// String returns a string representation of the board with algebraic notation
func (b *Board) String() string {
	var result string
	boardLines := []string{
		"\n     a    b    c    d    e    f    g    h",
		"   +----+----+----+----+----+----+----+----+",
	}

	// Add board representation
	for rank := 0; rank < 8; rank++ {
		line := fmt.Sprintf(" %d |", 8-rank)
		for file := 0; file < 8; file++ {
			piece := b.Squares[rank][file].Piece
			line += " " + PieceToString(piece) + " |"
		}
		line += fmt.Sprintf(" %d", 8-rank)
		boardLines = append(boardLines, line)
		boardLines = append(boardLines, "   +----+----+----+----+----+----+----+----+")
	}
	boardLines = append(boardLines, "     a    b    c    d    e    f    g    h")

	// Create moves list
	var movesLines []string
	if len(b.MovesPlayed) == 0 {
		movesLines = []string{"Moves: (none)"}
	} else {
		movesLines = append(movesLines, "Moves:")
		for i := 0; i < len(b.MovesPlayed); i += 2 {
			moveNum := (i / 2) + 1
			moveLine := fmt.Sprintf("  %d. %-8s", moveNum, b.MovesPlayed[i])
			if i+1 < len(b.MovesPlayed) {
				moveLine += fmt.Sprintf("%-8s", b.MovesPlayed[i+1])
			}
			movesLines = append(movesLines, moveLine)
		}
	}

	// Combine board and moves list side by side with proper spacing
	maxLines := len(boardLines)
	if len(movesLines) > maxLines {
		maxLines = len(movesLines)
	}

	// Add the board lines first
	for i := 0; i < len(boardLines); i++ {
		result += boardLines[i]
		// Add moves on the right side with proper spacing
		if i < len(movesLines) {
			// Add enough spaces to align moves to the right of the board
			padding := 55 - len(boardLines[i]) // Adjust padding based on board width
			if padding < 0 {
				padding = 2
			}
			result += strings.Repeat(" ", padding) + movesLines[i]
		}
		result += "\n"
	}

	// Add any remaining moves lines if there are more moves than board lines
	for i := len(boardLines); i < len(movesLines); i++ {
		result += strings.Repeat(" ", 55) + movesLines[i] + "\n"
	}

	// Add whose move it is and check status
	if b.WhiteToMove {
		result += "\nWhite to move"
		if b.isInCheck(true) {
			if b.isCheckmate(true) {
				result += " - CHECKMATE!"
			} else {
				result += " - CHECK!"
			}
		}
	} else {
		result += "\nBlack to move"
		if b.isInCheck(false) {
			if b.isCheckmate(false) {
				result += " - CHECKMATE!"
			} else {
				result += " - CHECK!"
			}
		}
	}
	return result
}
