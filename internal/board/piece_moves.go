package board

// Movement validation functions

// canPawnMove checks if a pawn can make the given move
func canPawnMove(b *Board, fromRank, fromFile, toRank, toFile int, isCapture bool) bool {
	piece := b.GetPiece(fromRank, fromFile)
	isWhite := piece < BP

	// Get direction - white moves up the board (-1 in rank), black moves down (+1 in rank)
	direction := -1              // White moves up the board (rank decreases)
	isStartRank := fromRank == 6 // White pawns start on rank 2 (index 6)
	if !isWhite {
		direction = 1               // Black moves down the board (rank increases)
		isStartRank = fromRank == 1 // Black pawns start on rank 7 (index 1)
	}

	if isCapture {
		// Pawn captures move one square diagonally
		rankDiff := toRank - fromRank
		fileDiff := abs(toFile - fromFile)

		// Must move one square diagonally forward
		if fileDiff != 1 || rankDiff != direction {
			return false
		}

		// Must capture an enemy piece (TODO: add en passant)
		targetPiece := b.GetPiece(toRank, toFile)
		// Check if target piece is enemy: white captures black (targetPiece >= BP), black captures white (targetPiece < BP)
		isTargetEnemy := (isWhite && targetPiece >= BP) || (!isWhite && targetPiece < BP)
		return targetPiece != Empty && isTargetEnemy
	}

	// Regular pawn move - must stay in the same file
	if fromFile != toFile {
		return false
	}

	// Can move one square forward
	rankDiff := toRank - fromRank // Difference from the pawn's perspective
	if rankDiff == direction {
		return b.IsSquareEmpty(toRank, toFile)
	}

	// Can move two squares forward from starting position
	if isStartRank && rankDiff == 2*direction {
		midRank := fromRank + direction // Move in direction of travel
		return b.IsSquareEmpty(toRank, toFile) && b.IsSquareEmpty(midRank, fromFile)
	}

	return false
}

// canBishopMove checks if a bishop can make the given move
func canBishopMove(b *Board, fromRank, fromFile, toRank, toFile int) bool {
	// Must move diagonally
	rankDiff := abs(toRank - fromRank)
	fileDiff := abs(toFile - fromFile)
	if rankDiff != fileDiff {
		return false
	}

	// Check path for obstacles
	rankStep := sign(toRank - fromRank)
	fileStep := sign(toFile - fromFile)
	rank, file := fromRank+rankStep, fromFile+fileStep
	for rank != toRank {
		if !b.IsSquareEmpty(rank, file) {
			return false
		}
		rank += rankStep
		file += fileStep
	}

	return true
}

// canRookMove checks if a rook can make the given move
func canRookMove(b *Board, fromRank, fromFile, toRank, toFile int) bool {
	// Must move horizontally or vertically
	if fromRank != toRank && fromFile != toFile {
		return false
	}

	// Check path for obstacles
	if fromRank == toRank {
		// Horizontal move
		step := sign(toFile - fromFile)
		for file := fromFile + step; file != toFile; file += step {
			if !b.IsSquareEmpty(fromRank, file) {
				return false
			}
		}
	} else {
		// Vertical move
		step := sign(toRank - fromRank)
		for rank := fromRank + step; rank != toRank; rank += step {
			if !b.IsSquareEmpty(rank, fromFile) {
				return false
			}
		}
	}

	return true
}

// canQueenMove checks if a queen can make the given move
func canQueenMove(b *Board, fromRank, fromFile, toRank, toFile int) bool {
	// Queen combines bishop and rook moves
	return canBishopMove(b, fromRank, fromFile, toRank, toFile) ||
		canRookMove(b, fromRank, fromFile, toRank, toFile)
}

// canKnightMove checks if a knight can make the given move
func canKnightMove(startRank, startFile, endRank, endFile int) bool {
	rankDiff := abs(endRank - startRank)
	fileDiff := abs(endFile - startFile)
	return (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2)
}

// canKingMove checks if a king can make the given move
func canKingMove(startRank, startFile, endRank, endFile int) bool {
	rankDiff := abs(endRank - startRank)
	fileDiff := abs(endFile - startFile)
	return rankDiff <= 1 && fileDiff <= 1
}

// Helper functions

// findKing returns the position of the specified color's king
func (b *Board) findKing(isWhite bool) (rank, file int) {
	kingPiece := BK
	if isWhite {
		kingPiece = WK
	}
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			if b.GetPiece(r, f) == kingPiece {
				return r, f
			}
		}
	}
	return -1, -1 // Should never happen in a valid game
}

// isSquareAttacked returns true if the given square can be captured by any enemy piece
func (b *Board) isSquareAttacked(rank, file int, attackerIsWhite bool) bool {
	// Check for attacking pawns
	direction := 1
	if attackerIsWhite {
		direction = -1
	}
	// Check pawn captures
	if rank-direction >= 0 && rank-direction < 8 {
		if file-1 >= 0 {
			piece := b.GetPiece(rank-direction, file-1)
			if piece == WP && attackerIsWhite || piece == BP && !attackerIsWhite {
				return true
			}
		}
		if file+1 < 8 {
			piece := b.GetPiece(rank-direction, file+1)
			if piece == WP && attackerIsWhite || piece == BP && !attackerIsWhite {
				return true
			}
		}
	}

	// Check knight attacks
	knightOffsets := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	attackerKnight := BN
	if attackerIsWhite {
		attackerKnight = WN
	}
	for _, offset := range knightOffsets {
		newRank, newFile := rank+offset[0], file+offset[1]
		if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
			if b.GetPiece(newRank, newFile) == attackerKnight {
				return true
			}
		}
	}

	// Check diagonal attacks (bishop/queen)
	attackerBishop, attackerQueen := BB, BQ
	if attackerIsWhite {
		attackerBishop, attackerQueen = WB, WQ
	}
	directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	for _, dir := range directions {
		r, f := rank+dir[0], file+dir[1]
		for r >= 0 && r < 8 && f >= 0 && f < 8 {
			piece := b.GetPiece(r, f)
			if piece != Empty {
				if piece == attackerBishop || piece == attackerQueen {
					return true
				}
				break
			}
			r, f = r+dir[0], f+dir[1]
		}
	}

	// Check horizontal/vertical attacks (rook/queen)
	attackerRook := BR
	if attackerIsWhite {
		attackerRook = WR
	}
	directions = [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, dir := range directions {
		r, f := rank+dir[0], file+dir[1]
		for r >= 0 && r < 8 && f >= 0 && f < 8 {
			piece := b.GetPiece(r, f)
			if piece != Empty {
				if piece == attackerRook || piece == attackerQueen {
					return true
				}
				break
			}
			r, f = r+dir[0], f+dir[1]
		}
	}

	// Check king attacks (for completeness)
	attackerKing := BK
	if attackerIsWhite {
		attackerKing = WK
	}
	for r := rank - 1; r <= rank+1; r++ {
		for f := file - 1; f <= file+1; f++ {
			if r >= 0 && r < 8 && f >= 0 && f < 8 && (r != rank || f != file) {
				if b.GetPiece(r, f) == attackerKing {
					return true
				}
			}
		}
	}

	return false
}

// isInCheck returns true if the specified color's king is in check
func (b *Board) isInCheck(isWhite bool) bool {
	kingRank, kingFile := b.findKing(isWhite)
	return b.isSquareAttacked(kingRank, kingFile, !isWhite)
}

// isCheckmate returns true if the specified color is in checkmate
func (b *Board) isCheckmate(isWhite bool) bool {
	// First, the king must be in check
	if !b.isInCheck(isWhite) {
		return false
	}

	// Try all possible moves for this color to see if any can escape check
	for fromRank := 0; fromRank < 8; fromRank++ {
		for fromFile := 0; fromFile < 8; fromFile++ {
			piece := b.GetPiece(fromRank, fromFile)

			// Skip empty squares and opponent pieces
			if piece == Empty || (piece < BP) != isWhite {
				continue
			}

			// Try all possible destination squares for this piece
			for toRank := 0; toRank < 8; toRank++ {
				for toFile := 0; toFile < 8; toFile++ {
					// Skip moving to the same square
					if fromRank == toRank && fromFile == toFile {
						continue
					}

					// Check if this piece can legally move to this square
					canMove := false
					switch piece {
					case WP, BP:
						// Check if it's a capture
						targetPiece := b.GetPiece(toRank, toFile)
						isCapture := targetPiece != Empty
						canMove = canPawnMove(b, fromRank, fromFile, toRank, toFile, isCapture)
					case WN, BN:
						canMove = canKnightMove(fromRank, fromFile, toRank, toFile)
					case WB, BB:
						canMove = canBishopMove(b, fromRank, fromFile, toRank, toFile)
					case WR, BR:
						canMove = canRookMove(b, fromRank, fromFile, toRank, toFile)
					case WQ, BQ:
						canMove = canQueenMove(b, fromRank, fromFile, toRank, toFile)
					case WK, BK:
						canMove = canKingMove(fromRank, fromFile, toRank, toFile)
					}

					if !canMove {
						continue
					}

					// Check if the destination square is valid for capture/movement
					targetPiece := b.GetPiece(toRank, toFile)
					if targetPiece != Empty {
						// Can't capture own pieces
						if (targetPiece < BP) == isWhite {
							continue
						}
					}

					// Try the move temporarily
					originalPiece := targetPiece
					b.Squares[toRank][toFile].Piece = piece
					b.Squares[fromRank][fromFile].Piece = Empty

					// Check if the king is still in check after this move
					stillInCheck := b.isInCheck(isWhite)

					// Undo the move
					b.Squares[fromRank][fromFile].Piece = piece
					b.Squares[toRank][toFile].Piece = originalPiece

					// If this move gets us out of check, it's not checkmate
					if !stillInCheck {
						return false
					}
				}
			}
		}
	}

	// No legal move can escape check, so it's checkmate
	return true
}

// abs returns the absolute value of x
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// sign returns -1 for negative numbers, 1 for positive numbers, and 0 for 0
func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}
