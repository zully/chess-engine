package engine

import (
	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/moves"
)

// GenerateMoves generates all possible legal moves for the current position
func GenerateMoves(b *board.Board) []moves.Move {
	var moveList []moves.Move

	// Generate moves for each square
	for fromRank := 0; fromRank < 8; fromRank++ {
		for fromFile := 0; fromFile < 8; fromFile++ {
			piece := b.GetPiece(fromRank, fromFile)
			isWhite := piece < board.BP

			// Skip empty squares and pieces of the wrong color
			if piece == board.Empty || isWhite != b.WhiteToMove {
				continue
			}

			// Generate pawn moves
			if piece == board.WP || piece == board.BP {
				// Direction for pawn movement
				direction := 1
				if isWhite {
					direction = -1
				}

				// One square forward
				newRank := fromRank + direction
				if newRank >= 0 && newRank < 8 && b.IsSquareEmpty(newRank, fromFile) {
					from := board.GetSquareName(fromRank, fromFile)
					to := board.GetSquareName(newRank, fromFile)

					// Check if this is a promotion move
					if (isWhite && newRank == 0) || (!isWhite && newRank == 7) {
						// Generate promotion moves (Queen, Rook, Bishop, Knight)
						for _, promotionPiece := range []string{"Q", "R", "B", "N"} {
							moveList = append(moveList, moves.Move{
								From:    from,
								To:      to,
								Piece:   "P",
								Promote: promotionPiece,
							})
						}
					} else {
						moveList = append(moveList, moves.Move{
							From:  from,
							To:    to,
							Piece: "P",
						})
					}

					// Two squares forward from starting position
					if (isWhite && fromRank == 6) || (!isWhite && fromRank == 1) {
						newRank = fromRank + 2*direction
						if b.IsSquareEmpty(newRank, fromFile) && b.IsSquareEmpty(fromRank+direction, fromFile) {
							to = board.GetSquareName(newRank, fromFile)
							moveList = append(moveList, moves.Move{
								From:  from,
								To:    to,
								Piece: "P",
							})
						}
					}
				}

				// Captures
				for _, fileOffset := range []int{-1, 1} {
					newFile := fromFile + fileOffset
					if newFile >= 0 && newFile < 8 {
						newRank := fromRank + direction
						if newRank >= 0 && newRank < 8 {
							targetPiece := b.GetPiece(newRank, newFile)
							canCapture := (isWhite && targetPiece >= board.BP) || (!isWhite && targetPiece < board.BP)

							// Regular capture
							if targetPiece != board.Empty && canCapture {
								from := board.GetSquareName(fromRank, fromFile)
								to := board.GetSquareName(newRank, newFile)

								// Check if this is a promotion capture
								if (isWhite && newRank == 0) || (!isWhite && newRank == 7) {
									// Generate promotion capture moves
									for _, promotionPiece := range []string{"Q", "R", "B", "N"} {
										moveList = append(moveList, moves.Move{
											From:    from,
											To:      to,
											Piece:   "P",
											Capture: true,
											Promote: promotionPiece,
										})
									}
								} else {
									moveList = append(moveList, moves.Move{
										From:    from,
										To:      to,
										Piece:   "P",
										Capture: true,
									})
								}
							}

							// En passant capture
							if b.EnPassant != "" {
								targetSquareName := board.GetSquareName(newRank, newFile)
								if b.EnPassant == targetSquareName {
									from := board.GetSquareName(fromRank, fromFile)
									to := targetSquareName
									moveList = append(moveList, moves.Move{
										From:      from,
										To:        to,
										Piece:     "P",
										Capture:   true,
										EnPassant: true,
									})
								}
							}
						}
					}
				}
			}

			// Generate knight moves
			if piece == board.WN || piece == board.BN {
				knightOffsets := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
				for _, offset := range knightOffsets {
					newRank, newFile := fromRank+offset[0], fromFile+offset[1]
					if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
						targetPiece := b.GetPiece(newRank, newFile)
						// Can move to empty square or capture enemy piece
						canCapture := (isWhite && targetPiece >= board.BP) || (!isWhite && targetPiece < board.BP)
						if targetPiece == board.Empty || canCapture {
							from := board.GetSquareName(fromRank, fromFile)
							to := board.GetSquareName(newRank, newFile)
							moveList = append(moveList, moves.Move{
								From:    from,
								To:      to,
								Piece:   "N",
								Capture: targetPiece != board.Empty,
							})
						}
					}
				}
			}

			// Generate bishop moves
			if piece == board.WB || piece == board.BB {
				directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
				for _, dir := range directions {
					for distance := 1; distance < 8; distance++ {
						newRank := fromRank + dir[0]*distance
						newFile := fromFile + dir[1]*distance
						if newRank < 0 || newRank >= 8 || newFile < 0 || newFile >= 8 {
							break
						}
						targetPiece := b.GetPiece(newRank, newFile)
						if targetPiece != board.Empty {
							// Can capture enemy piece
							canCapture := (isWhite && targetPiece >= board.BP) || (!isWhite && targetPiece < board.BP)
							if canCapture {
								from := board.GetSquareName(fromRank, fromFile)
								to := board.GetSquareName(newRank, newFile)
								moveList = append(moveList, moves.Move{
									From:    from,
									To:      to,
									Piece:   "B",
									Capture: true,
								})
							}
							break // Path blocked
						}
						// Empty square
						from := board.GetSquareName(fromRank, fromFile)
						to := board.GetSquareName(newRank, newFile)
						moveList = append(moveList, moves.Move{
							From:  from,
							To:    to,
							Piece: "B",
						})
					}
				}
			}

			// Generate rook moves
			if piece == board.WR || piece == board.BR {
				directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
				for _, dir := range directions {
					for distance := 1; distance < 8; distance++ {
						newRank := fromRank + dir[0]*distance
						newFile := fromFile + dir[1]*distance
						if newRank < 0 || newRank >= 8 || newFile < 0 || newFile >= 8 {
							break
						}
						targetPiece := b.GetPiece(newRank, newFile)
						if targetPiece != board.Empty {
							// Can capture enemy piece
							canCapture := (isWhite && targetPiece >= board.BP) || (!isWhite && targetPiece < board.BP)
							if canCapture {
								from := board.GetSquareName(fromRank, fromFile)
								to := board.GetSquareName(newRank, newFile)
								moveList = append(moveList, moves.Move{
									From:    from,
									To:      to,
									Piece:   "R",
									Capture: true,
								})
							}
							break // Path blocked
						}
						// Empty square
						from := board.GetSquareName(fromRank, fromFile)
						to := board.GetSquareName(newRank, newFile)
						moveList = append(moveList, moves.Move{
							From:  from,
							To:    to,
							Piece: "R",
						})
					}
				}
			}

			// Generate queen moves (combination of bishop and rook)
			if piece == board.WQ || piece == board.BQ {
				directions := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
				for _, dir := range directions {
					for distance := 1; distance < 8; distance++ {
						newRank := fromRank + dir[0]*distance
						newFile := fromFile + dir[1]*distance
						if newRank < 0 || newRank >= 8 || newFile < 0 || newFile >= 8 {
							break
						}
						targetPiece := b.GetPiece(newRank, newFile)
						if targetPiece != board.Empty {
							// Can capture enemy piece
							canCapture := (isWhite && targetPiece >= board.BP) || (!isWhite && targetPiece < board.BP)
							if canCapture {
								from := board.GetSquareName(fromRank, fromFile)
								to := board.GetSquareName(newRank, newFile)
								moveList = append(moveList, moves.Move{
									From:    from,
									To:      to,
									Piece:   "Q",
									Capture: true,
								})
							}
							break // Path blocked
						}
						// Empty square
						from := board.GetSquareName(fromRank, fromFile)
						to := board.GetSquareName(newRank, newFile)
						moveList = append(moveList, moves.Move{
							From:  from,
							To:    to,
							Piece: "Q",
						})
					}
				}
			}

			// Generate king moves
			if piece == board.WK || piece == board.BK {
				directions := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
				for _, dir := range directions {
					newRank := fromRank + dir[0]
					newFile := fromFile + dir[1]
					if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
						targetPiece := b.GetPiece(newRank, newFile)
						canCapture := (isWhite && targetPiece >= board.BP) || (!isWhite && targetPiece < board.BP)
						if targetPiece == board.Empty || canCapture {
							from := board.GetSquareName(fromRank, fromFile)
							to := board.GetSquareName(newRank, newFile)
							moveList = append(moveList, moves.Move{
								From:    from,
								To:      to,
								Piece:   "K",
								Capture: targetPiece != board.Empty,
							})
						}
					}
				}
			}
		}
	}

	// Filter out illegal moves (moves that leave king in check)
	var legalMoves []moves.Move
	for _, move := range moveList {
		if isLegalMove(b, move) {
			legalMoves = append(legalMoves, move)
		}
	}

	return legalMoves
}

// isLegalMove checks if a move is legal (doesn't leave king in check)
func isLegalMove(b *board.Board, move moves.Move) bool {
	fromSquare := b.GetSquare(move.From)
	toSquare := b.GetSquare(move.To)

	if fromSquare == nil || toSquare == nil {
		return false
	}

	// Determine whose move this is (piece color)
	isWhiteMove := fromSquare.Piece < board.BP

	// Make the move temporarily
	originalFromPiece := fromSquare.Piece
	originalToPiece := toSquare.Piece
	originalEnPassant := b.EnPassant

	// Handle en passant capture
	var enPassantSquare *board.Square
	if move.EnPassant {
		toRank := 7 - int(move.To[1]-'1')
		toFile := int(move.To[0] - 'a')

		capturedPawnRank := toRank
		if isWhiteMove {
			capturedPawnRank = toRank + 1
		} else {
			capturedPawnRank = toRank - 1
		}
		enPassantSquare = b.GetSquareByCoords(capturedPawnRank, toFile)
		if enPassantSquare != nil {
			enPassantSquare.Piece = board.Empty
		}
	}

	toSquare.Piece = fromSquare.Piece
	fromSquare.Piece = board.Empty

	// Check if the player who just moved left their own king in check
	inCheck := b.IsInCheck(isWhiteMove)

	// Undo the move
	fromSquare.Piece = originalFromPiece
	toSquare.Piece = originalToPiece
	b.EnPassant = originalEnPassant

	// Restore en passant captured pawn
	if move.EnPassant && enPassantSquare != nil {
		if isWhiteMove {
			enPassantSquare.Piece = board.BP
		} else {
			enPassantSquare.Piece = board.WP
		}
	}

	return !inCheck
}
