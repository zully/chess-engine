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
					moveList = append(moveList, moves.Move{
						From:  from,
						To:    to,
						Piece: "P",
					})

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
							if targetPiece != board.Empty && (targetPiece >= board.BP) != isWhite {
								from := board.GetSquareName(fromRank, fromFile)
								to := board.GetSquareName(newRank, newFile)
								moveList = append(moveList, moves.Move{
									From:    from,
									To:      to,
									Piece:   "P",
									Capture: true,
								})
							}
						}
					}
				}
			}

			// TODO: Add moves for other piece types
		}
	}

	return moveList
}
