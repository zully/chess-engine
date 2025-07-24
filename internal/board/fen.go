package board

import (
	"strconv"
	"strings"
)

// ToFEN converts the current board position to FEN notation
func (b *Board) ToFEN() string {
	var fen strings.Builder

	// 1. Piece placement
	for rank := 0; rank < 8; rank++ {
		emptyCount := 0
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == Empty {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				fen.WriteRune(pieceToFENChar(piece))
			}
		}
		if emptyCount > 0 {
			fen.WriteString(strconv.Itoa(emptyCount))
		}
		if rank < 7 {
			fen.WriteRune('/')
		}
	}

	// 2. Active color
	fen.WriteRune(' ')
	if b.WhiteToMove {
		fen.WriteRune('w')
	} else {
		fen.WriteRune('b')
	}

	// 3. Castling availability
	fen.WriteRune(' ')
	castling := ""
	if b.CastlingRights&1 != 0 { // White kingside
		castling += "K"
	}
	if b.CastlingRights&2 != 0 { // White queenside
		castling += "Q"
	}
	if b.CastlingRights&4 != 0 { // Black kingside
		castling += "k"
	}
	if b.CastlingRights&8 != 0 { // Black queenside
		castling += "q"
	}
	if castling == "" {
		fen.WriteRune('-')
	} else {
		fen.WriteString(castling)
	}

	// 4. En passant target square
	fen.WriteRune(' ')
	if b.EnPassant == "" {
		fen.WriteRune('-')
	} else {
		fen.WriteString(b.EnPassant)
	}

	// 5. Halfmove clock
	fen.WriteRune(' ')
	fen.WriteString(strconv.Itoa(b.HalfMoveClock))

	// 6. Fullmove number
	fen.WriteRune(' ')
	fen.WriteString(strconv.Itoa(b.FullMoveNumber))

	return fen.String()
}

// pieceToFENChar converts a piece constant to its FEN character representation
func pieceToFENChar(piece int) rune {
	switch piece {
	case WP:
		return 'P'
	case WN:
		return 'N'
	case WB:
		return 'B'
	case WR:
		return 'R'
	case WQ:
		return 'Q'
	case WK:
		return 'K'
	case BP:
		return 'p'
	case BN:
		return 'n'
	case BB:
		return 'b'
	case BR:
		return 'r'
	case BQ:
		return 'q'
	case BK:
		return 'k'
	default:
		return '?'
	}
}
