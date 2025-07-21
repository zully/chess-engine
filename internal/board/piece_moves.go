package board

import "fmt"

// Movement validation functions

// canPawnMove checks if a pawn can make the given move
func canPawnMove(b *Board, fromRank, fromFile, toRank, toFile int, isCapture bool) bool {
	piece := b.GetPiece(fromRank, fromFile)
	isWhite := piece < BP

	fmt.Printf("DEBUG: canPawnMove: fromRank=%d, fromFile=%d, toRank=%d, toFile=%d, isCapture=%v, isWhite=%v\n",
		fromRank, fromFile, toRank, toFile, isCapture, isWhite)

	// Get direction - white moves up the board (-1 in rank), black moves down (+1 in rank)
	direction := -1              // White moves up the board (rank decreases)
	isStartRank := fromRank == 6 // White pawns start on rank 2 (index 6)
	if !isWhite {
		direction = 1               // Black moves down the board (rank increases)
		isStartRank = fromRank == 1 // Black pawns start on rank 7 (index 1)
	}

	fmt.Printf("DEBUG: Pawn direction=%d, isStartRank=%v\n", direction, isStartRank)

	if isCapture {
		// Pawn captures move one square diagonally
		rankDiff := toRank - fromRank
		fileDiff := abs(toFile - fromFile)
		fmt.Printf("DEBUG: Pawn capture from (%d,%d) to (%d,%d), direction %d\n", fromRank, fromFile, toRank, toFile, direction)
		fmt.Printf("DEBUG: Pawn isWhite=%v, rankDiff=%d, fileDiff=%d\n", isWhite, rankDiff, fileDiff)

		// Must move one square diagonally forward
		if fileDiff != 1 || rankDiff != direction {
			fmt.Printf("DEBUG: Invalid capture: fileDiff=%d, rankDiff=%d, expected direction %d\n", fileDiff, rankDiff, direction)
			return false
		}

		// Must capture an enemy piece (TODO: add en passant)
		targetPiece := b.GetPiece(toRank, toFile)
		fmt.Printf("DEBUG: Target piece: %s\n", PieceToString(targetPiece))
		return targetPiece != Empty && ((targetPiece >= BP) != isWhite)
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

	// Get direction - white moves up the board (-1 in rank), black moves down (+1 in rank)
	direction := -1              // White moves up the board (rank decreases)
	isStartRank := fromRank == 6 // White pawns start on rank 2 (index 6)
	if !isWhite {
		direction = 1               // Black moves down the board (rank increases)
		isStartRank = fromRank == 1 // Black pawns start on rank 7 (index 1)
	}

	fmt.Printf("DEBUG: Pawn direction=%d, isStartRank=%v\n", direction, isStartRank)

	if isCapture {
		// Pawn captures move one square diagonally
		rankDiff := toRank - fromRank
		fileDiff := abs(toFile - fromFile)
		fmt.Printf("DEBUG: Pawn capture from (%d,%d) to (%d,%d), direction %d\n", fromRank, fromFile, toRank, toFile, direction)
		fmt.Printf("DEBUG: Pawn isWhite=%v, rankDiff=%d, fileDiff=%d\n", isWhite, rankDiff, fileDiff)

		// Must move one square diagonally forward
		if fileDiff != 1 || rankDiff != direction {
			fmt.Printf("DEBUG: Invalid capture: fileDiff=%d, rankDiff=%d, expected direction %d\n", fileDiff, rankDiff, direction)
			return false
		}

		// Must capture an enemy piece (TODO: add en passant)
		targetPiece := b.GetPiece(toRank, toFile)
		fmt.Printf("DEBUG: Target piece: %s\n", PieceToString(targetPiece))
		return targetPiece != Empty && ((targetPiece >= BP) != isWhite)
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

// getSquareCoords converts algebraic notation (e.g., "e4") to rank and file (0-7)
// The board is indexed with white at the bottom (rank 1-8) and files a-h from left to right.
// So for example, e4 should convert to (4,4) in array indices.
func getSquareCoords(square string) (rank int, file int) {
	if len(square) != 2 {
		return -1, -1
	}
	file = int(square[0] - 'a')
	rank = 7 - (int(square[1] - '1')) // Convert '4' to array index 4 from bottom (7-3=4)
	fmt.Printf("DEBUG: Converting %s to rank %d, file %d\n", square, rank, file)
	return rank, file
}
