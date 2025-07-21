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

// makeTemporaryMove makes a move and returns a function to undo it
func (b *Board) makeTemporaryMove(from, to string) func() {
	fromSquare := b.GetSquare(from)
	toSquare := b.GetSquare(to)
	oldFromPiece := fromSquare.Piece
	oldToPiece := toSquare.Piece

	// Make the move
	toSquare.Piece = fromSquare.Piece
	fromSquare.Piece = Empty

	// Return an undo function
	return func() {
		fromSquare.Piece = oldFromPiece
		toSquare.Piece = oldToPiece
	}
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
	if b.isInCheck(b.WhiteToMove) {
		// Try the move temporarily
		oldFromPiece := fromSquare.Piece
		oldToPiece := toSquare.Piece
		toSquare.Piece = fromSquare.Piece
		fromSquare.Piece = Empty

		stillInCheck := b.isInCheck(b.WhiteToMove)

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
		isValid = canKingMove(startRank, startFile, endRank, endFile)
	default:
		return fmt.Errorf("invalid piece")
	}

	if !isValid {
		return fmt.Errorf("illegal move for %s: %s", move.Piece, notation)
	}

	// Make the move
	toSquare.Piece = fromSquare.Piece
	fromSquare.Piece = Empty

	// Record the move
	b.MovesPlayed = append(b.MovesPlayed, notation)

	// Switch turns
	b.WhiteToMove = !b.WhiteToMove

	// Check if the opponent is in check after this move
	if b.isInCheck(b.WhiteToMove) {
		if b.WhiteToMove {
			fmt.Println("White is in check!")
		} else {
			fmt.Println("Black is in check!")
		}
	}

	return nil
}
