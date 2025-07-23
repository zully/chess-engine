package board

import (
	"fmt"
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

// FromFEN sets the board position from FEN notation
func (b *Board) FromFEN(fen string) error {
	parts := strings.Fields(fen)
	if len(parts) != 6 {
		return fmt.Errorf("invalid FEN: expected 6 parts, got %d", len(parts))
	}

	// Clear the board
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			b.Squares[rank][file].Piece = Empty
		}
	}

	// 1. Parse piece placement
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return fmt.Errorf("invalid FEN: expected 8 ranks, got %d", len(ranks))
	}

	for rankIdx, rankStr := range ranks {
		file := 0
		for _, char := range rankStr {
			if char >= '1' && char <= '8' {
				// Empty squares
				emptyCount := int(char - '0')
				file += emptyCount
			} else {
				// Piece
				piece, err := fenCharToPiece(char)
				if err != nil {
					return fmt.Errorf("invalid FEN: %v", err)
				}
				if file >= 8 {
					return fmt.Errorf("invalid FEN: too many pieces in rank %d", rankIdx+1)
				}
				b.Squares[rankIdx][file].Piece = piece
				file++
			}
		}
	}

	// 2. Parse active color
	if parts[1] == "w" {
		b.WhiteToMove = true
	} else if parts[1] == "b" {
		b.WhiteToMove = false
	} else {
		return fmt.Errorf("invalid FEN: invalid active color '%s'", parts[1])
	}

	// 3. Parse castling availability
	b.CastlingRights = 0
	if parts[2] != "-" {
		for _, char := range parts[2] {
			switch char {
			case 'K':
				b.CastlingRights |= 1 // White kingside
			case 'Q':
				b.CastlingRights |= 2 // White queenside
			case 'k':
				b.CastlingRights |= 4 // Black kingside
			case 'q':
				b.CastlingRights |= 8 // Black queenside
			default:
				return fmt.Errorf("invalid FEN: invalid castling right '%c'", char)
			}
		}
	}

	// 4. Parse en passant target square
	if parts[3] == "-" {
		b.EnPassant = ""
	} else {
		b.EnPassant = parts[3]
	}

	// 5. Parse halfmove clock
	halfmove, err := strconv.Atoi(parts[4])
	if err != nil {
		return fmt.Errorf("invalid FEN: invalid halfmove clock '%s'", parts[4])
	}
	b.HalfMoveClock = halfmove

	// 6. Parse fullmove number
	fullmove, err := strconv.Atoi(parts[5])
	if err != nil {
		return fmt.Errorf("invalid FEN: invalid fullmove number '%s'", parts[5])
	}
	b.FullMoveNumber = fullmove

	// Initialize position history for the new position
	b.PositionHistory = make(map[uint64]int)
	b.RecordPosition()

	return nil
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

// fenCharToPiece converts a FEN character to a piece constant
func fenCharToPiece(char rune) (int, error) {
	switch char {
	case 'P':
		return WP, nil
	case 'N':
		return WN, nil
	case 'B':
		return WB, nil
	case 'R':
		return WR, nil
	case 'Q':
		return WQ, nil
	case 'K':
		return WK, nil
	case 'p':
		return BP, nil
	case 'n':
		return BN, nil
	case 'b':
		return BB, nil
	case 'r':
		return BR, nil
	case 'q':
		return BQ, nil
	case 'k':
		return BK, nil
	default:
		return Empty, fmt.Errorf("invalid piece character: %c", char)
	}
}
