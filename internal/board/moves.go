package board

import (
	"fmt"
	"strings"

	"github.com/zully/chess-engine/internal/moves"
)

// FindPieceForMove finds the piece that can legally make the specified move
func (b *Board) FindPieceForMove(move *moves.Move) (string, error) {
	if len(move.To) != 2 {
		return "", fmt.Errorf("invalid target square: %s", move.To)
	}

	targetRank, targetFile := getSquareCoords(move.To)
	if targetRank < 0 || targetRank > 7 || targetFile < 0 || targetFile > 7 {
		return "", fmt.Errorf("invalid target square: %s [rank %d, file %d]", move.To, targetRank, targetFile)
	}

	// Handle castling
	if move.Castle != "" {
		// Castle moves are fully specified, so just return the from square
		return move.From, nil
	}

	pieceType := move.Piece
	isWhite := b.WhiteToMove

	// Special handling for pawns - find the actual starting square
	if pieceType == "P" {
		fromFile := targetFile
		var fromRank int

		// Process the source file for captures
		if len(move.From) > 0 {
			fromFile = int(move.From[0] - 'a')
			if fromFile < 0 || fromFile > 7 {
				return "", fmt.Errorf("invalid file: %s", string(move.From[0]))
			}
		}

		// Expected piece type
		expectedPiece := WP
		if !isWhite {
			expectedPiece = BP
		}

		// For captures, look one square diagonally back
		if move.Capture {
			// Get the rank for the capturing pawn
			if !isWhite {
				fromRank = targetRank - 1 // Black capturing pawn must be one rank higher
			} else {
				fromRank = targetRank + 1 // White capturing pawn must be one rank lower
			}

			if fromFile < 0 || fromFile > 7 || fromRank < 0 || fromRank > 7 {
				return "", fmt.Errorf("invalid source square: rank %d, file %d out of bounds", fromRank, fromFile)
			}

			piece := b.GetPiece(fromRank, fromFile)
			if piece == expectedPiece && canPawnMove(b, fromRank, fromFile, targetRank, targetFile, move.Capture) {
				return GetSquareName(fromRank, fromFile), nil
			}
		} else {
			// For normal pawn moves, try one square back first
			if isWhite {
				fromRank = targetRank + 1
			} else {
				fromRank = targetRank - 1
			}

			if fromRank >= 0 && fromRank <= 7 {
				piece := b.GetPiece(fromRank, fromFile)
				if piece == expectedPiece && canPawnMove(b, fromRank, fromFile, targetRank, targetFile, false) {
					return GetSquareName(fromRank, fromFile), nil
				}
			}

			// Then try two squares back from the starting position
			if isWhite && targetRank == 4 {
				fromRank = 6 // White pawn on rank 2 (index 6)
			} else if !isWhite && targetRank == 3 {
				fromRank = 1 // Black pawn on rank 7 (index 1)
			}

			if fromRank >= 0 && fromRank <= 7 {
				piece := b.GetPiece(fromRank, fromFile)
				if piece == expectedPiece && canPawnMove(b, fromRank, fromFile, targetRank, targetFile, false) {
					return GetSquareName(fromRank, fromFile), nil
				}
			}
		}

		return "", fmt.Errorf("no valid pawn found for move")
	}

	var piece int
	// Convert piece letter to piece constant
	switch pieceType {
	case "P":
		if isWhite {
			piece = WP
		} else {
			piece = BP
		}
	case "N":
		if isWhite {
			piece = WN
		} else {
			piece = BN
		}
	case "B":
		if isWhite {
			piece = WB
		} else {
			piece = BB
		}
	case "R":
		if isWhite {
			piece = WR
		} else {
			piece = BR
		}
	case "Q":
		if isWhite {
			piece = WQ
		} else {
			piece = BQ
		}
	case "K":
		if isWhite {
			piece = WK
		} else {
			piece = BK
		}
	default:
		return "", fmt.Errorf("invalid piece type: %s", pieceType)
	}

	// For pawns, we already know the from square
	if pieceType == "P" && move.From != "" {
		return move.From, nil
	}

	// Search for pieces of the correct type that could make this move
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if b.Squares[rank][file].Piece == piece {
				// Check if the piece can make the move according to chess rules
				canMove := false
				switch pieceType {
				case "P":
					canMove = canPawnMove(b, rank, file, targetRank, targetFile, move.Capture)
				case "N":
					canMove = canKnightMove(rank, file, targetRank, targetFile)
				case "B":
					canMove = canBishopMove(b, rank, file, targetRank, targetFile)
				case "R":
					canMove = canRookMove(b, rank, file, targetRank, targetFile)
				case "Q":
					canMove = canQueenMove(b, rank, file, targetRank, targetFile)
				case "K":
					canMove = canKingMove(rank, file, targetRank, targetFile)
				}

				if canMove {
					// Check if moving this piece would capture our own piece
					targetPiece := b.GetPiece(targetRank, targetFile)
					if targetPiece != Empty {
						// If target square has a piece, it must be an enemy piece
						if (targetPiece >= BP) == (piece >= BP) {
							continue // Same color piece, can't capture
						}
						// Different color piece, can capture
						if !move.Capture {
							continue // Move wasn't marked as a capture
						}
					} else if move.Capture {
						continue // Move was marked as a capture but square is empty
					}
					return b.Squares[rank][file].Name, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no %s found that could move to %s", pieceType, move.To)
}

// MakeMove makes a move on the board using algebraic notation
func (b *Board) MakeMove(notation string) error {
	move, err := moves.ParseAlgebraic(notation, b.WhiteToMove)
	if err != nil {
		return err
	}

	// Get ranks and files for validation
	startRank, startFile := 0, 0
	if move.From != "" {
		if move.From[len(move.From)-1] == '*' {
			// Pawn capture - we know the file but need to find the rank
			startFile = int(move.From[0] - 'a')
		} else {
			startRank, startFile = getSquareCoords(move.From)
		}
	}
	endRank, endFile := getSquareCoords(move.To)

	// If the from square isn't specified or contains wildcard (for piece moves and pawn captures), find it
	if move.From == "" || strings.Contains(move.From, "*") {
		from, err := b.FindPieceForMove(move)
		if err != nil {
			return err
		}
		move.From = from
		startRank, startFile = getSquareCoords(from)
	}

	// Get the squares
	fromSquare := b.GetSquare(move.From)
	toSquare := b.GetSquare(move.To)

	if fromSquare == nil || toSquare == nil {
		return fmt.Errorf("invalid square")
	}

	// Check that the piece belongs to the current player
	if fromSquare.Piece == Empty {
		return fmt.Errorf("invalid piece")
	}
	if b.WhiteToMove != (fromSquare.Piece < BP) {
		return fmt.Errorf("wrong player's piece")
	}

	// Additional validation based on piece type and move type
	piece := fromSquare.Piece
	isValid := false

	// First check if the move would get us out of check
	if b.IsInCheck(b.WhiteToMove) {
		// Try the move temporarily
		oldFromPiece := fromSquare.Piece
		oldToPiece := toSquare.Piece
		toSquare.Piece = fromSquare.Piece
		fromSquare.Piece = Empty

		stillInCheck := b.IsInCheck(b.WhiteToMove)

		// Undo the temporary move
		fromSquare.Piece = oldFromPiece
		toSquare.Piece = oldToPiece

		if stillInCheck {
			return fmt.Errorf("must respond to check")
		}
	}

	switch piece {
	case WP, BP:
		isValid = canPawnMove(b, startRank, startFile, endRank, endFile, move.Capture)
	case WN, BN:
		isValid = canKnightMove(startRank, startFile, endRank, endFile)
	case WB, BB:
		isValid = canBishopMove(b, startRank, startFile, endRank, endFile)
	case WR, BR:
		isValid = canRookMove(b, startRank, startFile, endRank, endFile)
	case WQ, BQ:
		isValid = canQueenMove(b, startRank, startFile, endRank, endFile)
	case WK, BK:
		// Handle castling moves specially
		if move.Castle != "" {
			isValid = b.canCastle(move.Castle, b.WhiteToMove)
		} else {
			isValid = canKingMove(startRank, startFile, endRank, endFile)
		}
	default:
		return fmt.Errorf("invalid piece")
	}

	if !isValid {
		return fmt.Errorf("illegal move for %s: %s", move.Piece, notation)
	}

	// Clear en passant target from previous move
	b.EnPassant = ""

	// Update castling rights if king or rook moves
	b.updateCastlingRights(move.From, fromSquare.Piece)

	// Handle castling moves specially
	if move.Castle != "" {
		b.executeCastling(move.Castle, b.WhiteToMove)
	} else {
		// Check for en passant capture before making the move
		isEnPassantCapture := move.EnPassant || (move.Piece == "P" && move.Capture && toSquare.Piece == Empty)

		// Make the move
		toSquare.Piece = fromSquare.Piece
		fromSquare.Piece = Empty

		// Handle en passant capture - remove the captured pawn
		if isEnPassantCapture {
			capturedPawnRank := endRank
			if b.WhiteToMove {
				capturedPawnRank = endRank + 1 // White captures black pawn one rank below
			} else {
				capturedPawnRank = endRank - 1 // Black captures white pawn one rank above
			}
			capturedPawnSquare := b.GetSquareByCoords(capturedPawnRank, endFile)
			if capturedPawnSquare != nil {
				capturedPawnSquare.Piece = Empty
			}
		}

		// Check for pawn two-square move to set en passant target
		if move.Piece == "P" && abs(endRank-startRank) == 2 {
			// Set en passant target square (the square the pawn passed over)
			targetRank := (startRank + endRank) / 2
			b.EnPassant = GetSquareName(targetRank, endFile)
		}

		// Check for pawn promotion
		if (toSquare.Piece == WP && endRank == 0) || (toSquare.Piece == BP && endRank == 7) {
			// Determine promotion piece (default to Queen if not specified)
			promotionPiece := move.Promote
			if promotionPiece == "" {
				promotionPiece = "Q" // Default to Queen
			}

			// Apply the promotion
			isWhitePiece := toSquare.Piece == WP
			switch promotionPiece {
			case "Q":
				if isWhitePiece {
					toSquare.Piece = WQ
				} else {
					toSquare.Piece = BQ
				}
			case "R":
				if isWhitePiece {
					toSquare.Piece = WR
				} else {
					toSquare.Piece = BR
				}
			case "B":
				if isWhitePiece {
					toSquare.Piece = WB
				} else {
					toSquare.Piece = BB
				}
			case "N":
				if isWhitePiece {
					toSquare.Piece = WN
				} else {
					toSquare.Piece = BN
				}
			default:
				// Fallback to Queen for invalid promotion pieces
				if isWhitePiece {
					toSquare.Piece = WQ
				} else {
					toSquare.Piece = BQ
				}
				promotionPiece = "Q"
			}
			// Add promotion notation only if not already present
			if !strings.Contains(notation, "=") {
				notation += "=" + promotionPiece
			}
			fmt.Printf("Pawn promoted to %s!\n", promotionPiece)
		}
	}

	// Switch turns
	b.WhiteToMove = !b.WhiteToMove

	// Record the position for repetition detection
	b.RecordPosition()

	// Check for draw conditions
	if b.IsDraw() {
		if b.IsThreefoldRepetition() {
			fmt.Println("Draw by threefold repetition!")
		} else {
			fmt.Println("Draw by stalemate!")
		}
	}

	// Check if the opponent is in check after this move
	if b.IsInCheck(b.WhiteToMove) {
		// Check if it's checkmate
		if b.IsCheckmate(b.WhiteToMove) {
			// Add checkmate notation only if not already present
			if !strings.Contains(notation, "#") && !strings.Contains(notation, "+") {
				notation += "#"
			}
			if b.WhiteToMove {
				fmt.Println("Checkmate! White is checkmated!")
			} else {
				fmt.Println("Checkmate! Black is checkmated!")
			}
		} else {
			// Add check notation only if not already present
			if !strings.Contains(notation, "+") && !strings.Contains(notation, "#") {
				notation += "+"
			}
			if b.WhiteToMove {
				fmt.Println("White is in check!")
			} else {
				fmt.Println("Black is in check!")
			}
		}
	}

	// Record the move (with check notation if applicable)
	b.MovesPlayed = append(b.MovesPlayed, notation)

	return nil
}

// canCastle checks if the specified castling move is legal
func (b *Board) canCastle(castleType string, isWhite bool) bool {
	// Check if king is currently in check (can't castle out of check)
	if b.IsInCheck(isWhite) {
		return false
	}

	var kingRank, rookRank int
	var kingFromFile, kingToFile, rookFromFile int

	if isWhite {
		kingRank = 7 // White king on rank 1 (index 7)
		rookRank = 7 // White rooks on rank 1 (index 7)
	} else {
		kingRank = 0 // Black king on rank 8 (index 0)
		rookRank = 0 // Black rooks on rank 8 (index 0)
	}

	// Set file positions based on castle type
	if castleType == "O-O" {
		// Kingside castling
		kingFromFile = 4 // e-file
		kingToFile = 6   // g-file
		rookFromFile = 7 // h-file
	} else if castleType == "O-O-O" {
		// Queenside castling
		kingFromFile = 4 // e-file
		kingToFile = 2   // c-file
		rookFromFile = 0 // a-file
	} else {
		return false
	}

	// Check that king and rook are in correct positions
	expectedKing := WK
	expectedRook := WR
	if !isWhite {
		expectedKing = BK
		expectedRook = BR
	}

	if b.GetPiece(kingRank, kingFromFile) != expectedKing {
		return false
	}
	if b.GetPiece(rookRank, rookFromFile) != expectedRook {
		return false
	}

	// Check that squares between king and rook are empty
	minFile := kingFromFile
	maxFile := rookFromFile
	if minFile > maxFile {
		minFile, maxFile = maxFile, minFile
	}

	for file := minFile + 1; file < maxFile; file++ {
		if !b.IsSquareEmpty(kingRank, file) {
			return false
		}
	}

	// Check that king doesn't pass through check
	// King moves from kingFromFile to kingToFile, check each square
	minFile = kingFromFile
	maxFile = kingToFile
	if minFile > maxFile {
		minFile, maxFile = maxFile, minFile
	}

	for file := minFile; file <= maxFile; file++ {
		if b.IsSquareAttacked(kingRank, file, !isWhite) {
			return false
		}
	}

	// Check castling rights
	if !b.hasCastlingRights(castleType, isWhite) {
		return false
	}

	return true
}

// hasCastlingRights checks if the player still has the specified castling rights
func (b *Board) hasCastlingRights(castleType string, isWhite bool) bool {
	// Castling rights are stored as bits: 0001=WK, 0010=WQ, 0100=BK, 1000=BQ
	if isWhite {
		if castleType == "O-O" {
			return (b.CastlingRights & 1) != 0 // White kingside
		} else if castleType == "O-O-O" {
			return (b.CastlingRights & 2) != 0 // White queenside
		}
	} else {
		if castleType == "O-O" {
			return (b.CastlingRights & 4) != 0 // Black kingside
		} else if castleType == "O-O-O" {
			return (b.CastlingRights & 8) != 0 // Black queenside
		}
	}
	return false
}

// updateCastlingRights removes castling rights when kings or rooks move
func (b *Board) updateCastlingRights(fromSquare string, piece int) {
	switch fromSquare {
	case "e1": // White king
		if piece == WK {
			b.CastlingRights &^= 3 // Remove both white castling rights (bits 0 and 1)
		}
	case "a1": // White queenside rook
		if piece == WR {
			b.CastlingRights &^= 2 // Remove white queenside (bit 1)
		}
	case "h1": // White kingside rook
		if piece == WR {
			b.CastlingRights &^= 1 // Remove white kingside (bit 0)
		}
	case "e8": // Black king
		if piece == BK {
			b.CastlingRights &^= 12 // Remove both black castling rights (bits 2 and 3)
		}
	case "a8": // Black queenside rook
		if piece == BR {
			b.CastlingRights &^= 8 // Remove black queenside (bit 3)
		}
	case "h8": // Black kingside rook
		if piece == BR {
			b.CastlingRights &^= 4 // Remove black kingside (bit 2)
		}
	}
}

// executeCastling performs the castling move (moves both king and rook)
func (b *Board) executeCastling(castleType string, isWhite bool) {
	var kingRank, rookRank int
	var kingFromFile, kingToFile, rookFromFile, rookToFile int

	if isWhite {
		kingRank = 7 // White king on rank 1 (index 7)
		rookRank = 7 // White rooks on rank 1 (index 7)
	} else {
		kingRank = 0 // Black king on rank 8 (index 0)
		rookRank = 0 // Black rooks on rank 8 (index 0)
	}

	// Set file positions based on castle type
	if castleType == "O-O" {
		// Kingside castling
		kingFromFile = 4 // e-file
		kingToFile = 6   // g-file
		rookFromFile = 7 // h-file
		rookToFile = 5   // f-file
	} else { // "O-O-O"
		// Queenside castling
		kingFromFile = 4 // e-file
		kingToFile = 2   // c-file
		rookFromFile = 0 // a-file
		rookToFile = 3   // d-file
	}

	// Move the king
	kingPiece := b.GetPiece(kingRank, kingFromFile)
	b.Squares[kingRank][kingFromFile].Piece = Empty
	b.Squares[kingRank][kingToFile].Piece = kingPiece

	// Move the rook
	rookPiece := b.GetPiece(rookRank, rookFromFile)
	b.Squares[rookRank][rookFromFile].Piece = Empty
	b.Squares[rookRank][rookToFile].Piece = rookPiece
}
