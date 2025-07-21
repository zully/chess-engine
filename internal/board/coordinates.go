package board

// getSquareCoords converts algebraic notation (e.g., "e4") to rank and file (0-7)
// The board is indexed with white at the bottom (rank 1-8) and files a-h from left to right.
func getSquareCoords(square string) (rank int, file int) {
	if len(square) < 1 {
		return -1, -1
	}

	// Handle file
	file = int(square[0] - 'a')
	if file < 0 || file > 7 {
		return -1, -1
	}

	// Handle rank
	if len(square) == 2 {
		// Full square specification (e.g., "e4")
		if square[1] < '1' || square[1] > '8' {
			return -1, -1
		}
		rank = 7 - (int(square[1] - '1')) // Convert '4' to array index 4 from bottom
	} else if len(square) > 2 && square[1] == '*' {
		// Partial square specification (e.g., "e*" for pawn captures)
		rank = -1 // Special value indicating any rank
	} else {
		return -1, -1
	}

	return rank, file
}

// GetSquareName returns the algebraic notation for a square given its rank and file (0-7)
// The board is indexed with white at the bottom (rank 1-8) and files a-h from left to right.
func GetSquareName(rank, file int) string {
	files := "abcdefgh"
	ranks := "12345678"
	return string(files[file]) + string(ranks[7-rank])
}
