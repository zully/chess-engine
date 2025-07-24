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
	Squares         [8][8]Square   // 8x8 board with named squares
	WhiteToMove     bool           // true if it's white's turn
	CastlingRights  int            // stores castling availability
	EnPassant       string         // en passant target square in algebraic notation
	HalfMoveClock   int            // counts moves since last pawn move or capture
	FullMoveNumber  int            // counts full moves in the game
	MovesPlayed     []string       // list of moves in algebraic notation
	PositionHistory map[uint64]int // tracks position occurrences for repetition detection
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

// GetPieceValue returns the point value of a piece for material calculation
func GetPieceValue(piece int) int {
	switch piece {
	case WP, BP:
		return 1
	case WN, BN, WB, BB:
		return 3
	case WR, BR:
		return 5
	case WQ, BQ:
		return 9
	default:
		return 0
	}
}

// GetPieceType returns the piece type as a single letter (P, N, B, R, Q, K)
func GetPieceType(piece int) string {
	switch piece {
	case WP, BP:
		return "P"
	case WN, BN:
		return "N"
	case WB, BB:
		return "B"
	case WR, BR:
		return "R"
	case WQ, BQ:
		return "Q"
	case WK, BK:
		return "K"
	default:
		return ""
	}
}

// NewBoard creates and returns a new board in the initial chess position
func NewBoard() *Board {
	b := &Board{
		WhiteToMove:     true,
		CastlingRights:  15, // 1111 in binary - all castling available
		EnPassant:       "", // no en passant target initially
		HalfMoveClock:   0,
		FullMoveNumber:  1,
		MovesPlayed:     make([]string, 0),
		PositionHistory: make(map[uint64]int),
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

	// Record the initial position
	b.RecordPosition()

	return b
}

// GetPiece returns the piece at the given rank and file (0-7)
func (b *Board) GetPiece(rank, file int) int {
	return b.Squares[rank][file].Piece
}

// GetSquare returns the square at the given algebraic notation (e.g., "e4")
func (b *Board) GetSquare(algebraicNotation string) *Square {
	rank, file := GetSquareCoords(algebraicNotation)
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return nil
	}
	return &b.Squares[rank][file]
}

// GetSquareByCoords returns the square at the given rank and file coordinates
func (b *Board) GetSquareByCoords(rank, file int) *Square {
	if rank < 0 || file < 0 || rank > 7 || file > 7 {
		return nil
	}
	return &b.Squares[rank][file]
}

// IsSquareEmpty returns true if the given square is empty
func (b *Board) IsSquareEmpty(rank, file int) bool {
	return b.GetPiece(rank, file) == Empty
}

// GetPositionHash generates a hash of the current position for repetition detection
// Hash includes: piece positions, whose turn, castling rights, en passant target
func (b *Board) GetPositionHash() uint64 {
	var hash uint64 = 14695981039346656037 // FNV-1a offset basis

	// Hash piece positions using FNV-1a algorithm (better distribution)
	const fnvPrime uint64 = 1099511628211
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			squareIndex := uint64(rank*8 + file)

			// Combine piece and square index into a single value
			value := (uint64(piece) << 8) | squareIndex

			// FNV-1a hash
			hash ^= value
			hash *= fnvPrime
		}
	}

	// Include whose turn it is
	turnValue := uint64(0)
	if b.WhiteToMove {
		turnValue = 1
	}
	hash ^= turnValue
	hash *= fnvPrime

	// Include castling rights
	hash ^= uint64(b.CastlingRights)
	hash *= fnvPrime

	// Include en passant target
	if b.EnPassant != "" {
		for _, c := range b.EnPassant {
			hash ^= uint64(c)
			hash *= fnvPrime
		}
	}

	return hash
}

// RecordPosition records the current position in history
func (b *Board) RecordPosition() {
	hash := b.GetPositionHash()
	b.PositionHistory[hash]++
}

// GetPositionCount returns how many times the current position has occurred
func (b *Board) GetPositionCount() int {
	hash := b.GetPositionHash()
	return b.PositionHistory[hash]
}

// IsThreefoldRepetition returns true if current position has occurred 3+ times
func (b *Board) IsThreefoldRepetition() bool {
	return b.GetPositionCount() >= 3
}

// IsDraw returns true if the position is a draw by repetition or stalemate
func (b *Board) IsDraw() bool {
	// Check for threefold repetition
	if b.IsThreefoldRepetition() {
		return true
	}

	// Check for stalemate (no legal moves but not in check)
	if !b.IsInCheck(b.WhiteToMove) {
		// Generate all legal moves to see if there are any
		// This is a simplified check - ideally we'd use the move generator
		hasLegalMove := false

		// Quick check: try to find at least one legal move
		for fromRank := 0; fromRank < 8 && !hasLegalMove; fromRank++ {
			for fromFile := 0; fromFile < 8 && !hasLegalMove; fromFile++ {
				piece := b.GetPiece(fromRank, fromFile)

				// Skip empty squares and opponent pieces
				if piece == Empty || (piece < BP) != b.WhiteToMove {
					continue
				}

				// Try a few potential moves for this piece
				for toRank := 0; toRank < 8 && !hasLegalMove; toRank++ {
					for toFile := 0; toFile < 8 && !hasLegalMove; toFile++ {
						if fromRank == toRank && fromFile == toFile {
							continue
						}

						// Test if this would be a legal move (simplified test)
						targetPiece := b.GetPiece(toRank, toFile)
						if targetPiece != Empty && (targetPiece < BP) == b.WhiteToMove {
							continue // Can't capture own piece
						}

						// Try the move temporarily
						b.Squares[toRank][toFile].Piece = piece
						b.Squares[fromRank][fromFile].Piece = Empty

						// Check if still in check after move
						stillInCheck := b.IsInCheck(b.WhiteToMove)

						// Undo the move
						b.Squares[fromRank][fromFile].Piece = piece
						b.Squares[toRank][toFile].Piece = targetPiece

						if !stillInCheck {
							hasLegalMove = true
						}
					}
				}
			}
		}

		if !hasLegalMove {
			return true // Stalemate
		}
	}

	return false
}

// MakeUCIMove makes a move on the board using UCI notation (e.g., "e2e4", "a1h8")
func (b *Board) MakeUCIMove(uciMove string) error {
	if len(uciMove) < 4 || len(uciMove) > 5 {
		return fmt.Errorf("invalid UCI move format: %s", uciMove)
	}

	// Parse from and to squares
	fromSquare := uciMove[0:2]
	toSquare := uciMove[2:4]

	// Parse promotion piece if present
	var promotionPiece string
	if len(uciMove) == 5 {
		promotionPiece = strings.ToUpper(string(uciMove[4]))
	}

	// Get coordinates
	fromRank, fromFile := GetSquareCoords(fromSquare)
	toRank, toFile := GetSquareCoords(toSquare)

	if fromRank < 0 || fromRank > 7 || fromFile < 0 || fromFile > 7 {
		return fmt.Errorf("invalid from square: %s", fromSquare)
	}
	if toRank < 0 || toRank > 7 || toFile < 0 || toFile > 7 {
		return fmt.Errorf("invalid to square: %s", toSquare)
	}

	// Get the squares
	fromSquareObj := b.GetSquareByCoords(fromRank, fromFile)
	toSquareObj := b.GetSquareByCoords(toRank, toFile)

	if fromSquareObj == nil || toSquareObj == nil {
		return fmt.Errorf("invalid square coordinates")
	}

	// Check that there's a piece to move
	if fromSquareObj.Piece == Empty {
		return fmt.Errorf("no piece on square %s", fromSquare)
	}

	// Check that the piece belongs to the current player
	isWhitePiece := fromSquareObj.Piece < BP
	if b.WhiteToMove != isWhitePiece {
		return fmt.Errorf("not your piece to move")
	}

	// Handle castling moves specially
	if fromSquareObj.Piece == WK || fromSquareObj.Piece == BK {
		// Check for castling
		if b.WhiteToMove && fromSquare == "e1" {
			if toSquare == "g1" && b.canCastle("O-O", true) {
				b.executeCastling("O-O", true)
				b.WhiteToMove = false
				b.RecordPosition()
				b.MovesPlayed = append(b.MovesPlayed, "O-O")
				return nil
			}
			if toSquare == "c1" && b.canCastle("O-O-O", true) {
				b.executeCastling("O-O-O", true)
				b.WhiteToMove = false
				b.RecordPosition()
				b.MovesPlayed = append(b.MovesPlayed, "O-O-O")
				return nil
			}
		} else if !b.WhiteToMove && fromSquare == "e8" {
			if toSquare == "g8" && b.canCastle("O-O", false) {
				b.executeCastling("O-O", false)
				b.WhiteToMove = true
				b.RecordPosition()
				b.MovesPlayed = append(b.MovesPlayed, "O-O")
				return nil
			}
			if toSquare == "c8" && b.canCastle("O-O-O", false) {
				b.executeCastling("O-O-O", false)
				b.WhiteToMove = true
				b.RecordPosition()
				b.MovesPlayed = append(b.MovesPlayed, "O-O-O")
				return nil
			}
		}
	}

	// Validate the move is legal for this piece type
	piece := fromSquareObj.Piece
	isCapture := toSquareObj.Piece != Empty

	if !b.isValidMove(piece, fromRank, fromFile, toRank, toFile, isCapture) {
		return fmt.Errorf("illegal move for piece")
	}

	// Check if moving would capture own piece
	if toSquareObj.Piece != Empty {
		targetIsWhite := toSquareObj.Piece < BP
		if isWhitePiece == targetIsWhite {
			return fmt.Errorf("cannot capture your own piece")
		}
	}

	// If the current player is in check, verify that this move gets them out of check
	currentPlayerIsWhite := b.WhiteToMove
	if b.IsInCheck(currentPlayerIsWhite) {
		// Try the move temporarily
		originalToPiece := toSquareObj.Piece
		toSquareObj.Piece = piece
		fromSquareObj.Piece = Empty

		stillInCheck := b.IsInCheck(currentPlayerIsWhite)

		// Undo the temporary move
		fromSquareObj.Piece = piece
		toSquareObj.Piece = originalToPiece

		if stillInCheck {
			return fmt.Errorf("must respond to check")
		}
	}

	// Convert UCI to algebraic BEFORE making the move (so we can still see the piece)
	algebraicMove := b.uciToAlgebraic(uciMove)

	// Store the original target piece for potential restoration
	originalTargetPiece := toSquareObj.Piece

	// Execute the move
	b.Squares[toRank][toFile].Piece = piece
	b.Squares[fromRank][fromFile].Piece = Empty

	// Verify that this move doesn't put our own king in check
	if b.IsInCheck(currentPlayerIsWhite) {
		// Undo the move
		b.Squares[fromRank][fromFile].Piece = piece
		b.Squares[toRank][toFile].Piece = originalTargetPiece
		return fmt.Errorf("move would put king in check")
	}

	// Handle pawn promotion
	if (piece == WP && toRank == 0) || (piece == BP && toRank == 7) {
		var newPiece int
		switch promotionPiece {
		case "Q":
			if piece == WP {
				newPiece = WQ
			} else {
				newPiece = BQ
			}
		case "R":
			if piece == WP {
				newPiece = WR
			} else {
				newPiece = BR
			}
		case "B":
			if piece == WP {
				newPiece = WB
			} else {
				newPiece = BB
			}
		case "N":
			if piece == WP {
				newPiece = WN
			} else {
				newPiece = BN
			}
		default:
			// Default to Queen if no promotion piece specified
			if piece == WP {
				newPiece = WQ
			} else {
				newPiece = BQ
			}
		}
		b.Squares[toRank][toFile].Piece = newPiece
	}

	// Handle en passant capture
	if (piece == WP || piece == BP) && isCapture && b.EnPassant == toSquare {
		// Remove the captured pawn
		capturedPawnRank := toRank
		if piece == WP {
			capturedPawnRank = toRank + 1
		} else {
			capturedPawnRank = toRank - 1
		}
		b.Squares[capturedPawnRank][toFile].Piece = Empty
	}

	// Handle en passant target setting
	if (piece == WP || piece == BP) && abs(toRank-fromRank) == 2 {
		targetRank := (fromRank + toRank) / 2
		b.EnPassant = GetSquareName(targetRank, toFile)
	} else {
		b.EnPassant = ""
	}

	// Update castling rights
	b.updateCastlingRights(fromSquare, piece)

	// Switch turns
	b.WhiteToMove = !b.WhiteToMove

	// Record position for repetition detection
	b.RecordPosition()

	// Add to move history
	b.MovesPlayed = append(b.MovesPlayed, algebraicMove)

	return nil
}

// isValidMove validates if a piece can legally move from one square to another
func (b *Board) isValidMove(piece int, fromRank, fromFile, toRank, toFile int, isCapture bool) bool {
	switch piece {
	case WP, BP:
		return canPawnMove(b, fromRank, fromFile, toRank, toFile, isCapture)
	case WN, BN:
		return CanKnightMove(fromRank, fromFile, toRank, toFile)
	case WB, BB:
		return CanBishopMove(b, fromRank, fromFile, toRank, toFile)
	case WR, BR:
		return CanRookMove(b, fromRank, fromFile, toRank, toFile)
	case WQ, BQ:
		return CanQueenMove(b, fromRank, fromFile, toRank, toFile)
	case WK, BK:
		return canKingMove(fromRank, fromFile, toRank, toFile)
	default:
		return false
	}
}

// uciToAlgebraic converts UCI move to algebraic notation for move history display
func (b *Board) uciToAlgebraic(uciMove string) string {
	// For now, return a simplified algebraic notation
	// This can be enhanced later for full algebraic notation with disambiguation

	if len(uciMove) < 4 {
		return uciMove
	}

	fromSquare := uciMove[0:2]
	toSquare := uciMove[2:4]

	// Handle castling
	if uciMove == "e1g1" || uciMove == "e8g8" {
		return "O-O"
	}
	if uciMove == "e1c1" || uciMove == "e8c8" {
		return "O-O-O"
	}

	// Get piece type from the from square
	fromRank, fromFile := GetSquareCoords(fromSquare)
	if fromRank < 0 || fromFile < 0 || fromRank > 7 || fromFile > 7 {
		return uciMove
	}

	piece := b.GetPiece(fromRank, fromFile)
	if piece == Empty {
		return uciMove
	}

	pieceType := GetPieceType(piece)

	// For pawns, just return the target square (or capture notation)
	if pieceType == "P" {
		toRank, toFileCoord := GetSquareCoords(toSquare)
		if toRank < 0 || toRank > 7 {
			return uciMove
		}

		// Check if it's a capture (diagonal move for pawn)
		if fromFile != toFileCoord {
			// Pawn capture
			result := fromSquare[0:1] + "x" + toSquare
			// Add promotion if present
			if len(uciMove) == 5 {
				result += "=" + strings.ToUpper(string(uciMove[4]))
			}
			return result
		} else {
			// Regular pawn move
			result := toSquare
			// Add promotion if present
			if len(uciMove) == 5 {
				result += "=" + strings.ToUpper(string(uciMove[4]))
			}
			return result
		}
	}

	// For other pieces, check if it's a capture and add disambiguation if needed
	toRank, toFile := GetSquareCoords(toSquare)
	if toRank < 0 || toFile < 0 || toRank > 7 || toFile > 7 {
		return uciMove
	}

	targetPiece := b.GetPiece(toRank, toFile)
	isCapture := targetPiece != Empty

	// Check if disambiguation is needed (other pieces of same type can move to same square)
	disambiguation := b.getDisambiguation(piece, fromRank, fromFile, toRank, toFile)

	var result string
	if isCapture {
		result = pieceType + disambiguation + "x" + toSquare
	} else {
		result = pieceType + disambiguation + toSquare
	}
	return result
}

// getDisambiguation returns the disambiguation string needed when multiple pieces can move to same square
func (b *Board) getDisambiguation(piece int, fromRank, fromFile, toRank, toFile int) string {
	// Find all pieces of the same type that could move to the same target square
	sameTypePieces := []struct{ rank, file int }{}

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if rank == fromRank && file == fromFile {
				continue // Skip the piece we're moving
			}

			boardPiece := b.GetPiece(rank, file)
			if boardPiece == piece {
				// Check if this piece could legally move to the target square
				targetPiece := b.GetPiece(toRank, toFile)
				isCapture := targetPiece != Empty
				if b.isValidMove(boardPiece, rank, file, toRank, toFile, isCapture) {
					sameTypePieces = append(sameTypePieces, struct{ rank, file int }{rank, file})
				}
			}
		}
	}

	// If no other pieces can move there, no disambiguation needed
	if len(sameTypePieces) == 0 {
		return ""
	}

	// Check if file disambiguation is sufficient
	fileUnique := true
	for _, p := range sameTypePieces {
		if p.file == fromFile {
			fileUnique = false
			break
		}
	}

	if fileUnique {
		return string(rune('a' + fromFile))
	}

	// Check if rank disambiguation is sufficient
	rankUnique := true
	for _, p := range sameTypePieces {
		if p.rank == fromRank {
			rankUnique = false
			break
		}
	}

	if rankUnique {
		return string(rune('1' + (7 - fromRank)))
	}

	// If neither file nor rank alone is sufficient, use both
	return string(rune('a'+fromFile)) + string(rune('1'+(7-fromRank)))
}
