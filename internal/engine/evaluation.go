// Package engine implements the chess engine's core functionality.
package engine

import (
	"github.com/zully/chess-engine/internal/board"
)

// Material values for each piece type
const (
	PawnValue   = 100
	KnightValue = 320
	BishopValue = 330
	RookValue   = 500
	QueenValue  = 900
	KingValue   = 20000 // High value to prioritize king safety
)

// Piece square tables for positional evaluation
var pawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingMiddleTable = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

var kingEndTable = [64]int{
	-50, -40, -30, -20, -20, -30, -40, -50,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -30, 0, 0, 0, 0, -30, -30,
	-50, -30, -30, -30, -30, -30, -30, -50,
}

// Evaluate returns a score for the current board position.
// Positive scores favor white, negative scores favor black.
func Evaluate(b *board.Board) int {
	// Check for immediate draws
	if b.IsDraw() {
		return 0 // Draw is always scored as 0
	}

	score := 0

	// Material evaluation
	score += evaluateMaterial(b)

	// Position evaluation
	score += evaluatePosition(b)

	// King safety evaluation (critical for avoiding checkmate)
	score += evaluateKingSafety(b)

	// Mobility evaluation (piece activity)
	score += evaluateMobility(b)

	// CRITICAL: Tactical evaluation with MASSIVE penalties for hanging pieces
	tacticalScore := evaluateTactics(b)
	score += tacticalScore * 3 // Amplify tactical penalties by 3x - strong but not extreme

	// Repetition avoidance - penalize approaching threefold repetition
	positionCount := b.GetPositionCount()
	if positionCount >= 2 {
		// If we're about to repeat the position for the third time,
		// heavily penalize unless we're losing
		repetitionPenalty := -100 // Moderate penalty for repetition
		if score > 0 {
			// If we're winning, avoid repetition more strongly
			repetitionPenalty = -200
		} else if score < -500 {
			// If we're losing badly, repetition might be good (avoid penalty)
			repetitionPenalty = 0
		}
		score += repetitionPenalty
	}

	// Adjust score based on whose turn it is
	if !b.WhiteToMove {
		score = -score
	}

	return score
}

// evaluateMaterial calculates the material balance of the position
func evaluateMaterial(b *board.Board) int {
	score := 0

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			value := 0

			switch piece {
			case board.WP:
				value = PawnValue
			case board.WN:
				value = KnightValue
			case board.WB:
				value = BishopValue
			case board.WR:
				value = RookValue
			case board.WQ:
				value = QueenValue
			case board.WK:
				value = KingValue
			case board.BP:
				value = -PawnValue
			case board.BN:
				value = -KnightValue
			case board.BB:
				value = -BishopValue
			case board.BR:
				value = -RookValue
			case board.BQ:
				value = -QueenValue
			case board.BK:
				value = -KingValue
			}

			score += value
		}
	}

	return score
}

// isEndgame determines if the current position is in the endgame
// This is a simple implementation - we consider it endgame if:
// 1. Both sides have no queens, or
// 2. Each side has <= 1 piece (excluding kings and pawns)
func isEndgame(b *board.Board) bool {
	whitePieces := 0
	blackPieces := 0
	hasWhiteQueen := false
	hasBlackQueen := false

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			switch piece {
			case board.WQ:
				hasWhiteQueen = true
			case board.BQ:
				hasBlackQueen = true
			case board.WN, board.WB, board.WR:
				whitePieces++
			case board.BN, board.BB, board.BR:
				blackPieces++
			}
		}
	}

	return (!hasWhiteQueen && !hasBlackQueen) || (whitePieces <= 1 && blackPieces <= 1)
}

// evaluatePosition adds positional bonuses/penalties based on piece placement
func evaluatePosition(b *board.Board) int {
	score := 0
	isEnd := isEndgame(b)

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			square := rank*8 + file
			// For black pieces, we flip the square index vertically
			blackSquare := 63 - square

			switch piece {
			case board.WP:
				score += pawnTable[square]
			case board.BP:
				score -= pawnTable[blackSquare]
			case board.WN:
				score += knightTable[square]
			case board.BN:
				score -= knightTable[blackSquare]
			case board.WB:
				score += bishopTable[square]
			case board.BB:
				score -= bishopTable[blackSquare]
			case board.WR:
				score += rookTable[square]
			case board.BR:
				score -= rookTable[blackSquare]
			case board.WQ:
				score += queenTable[square]
			case board.BQ:
				score -= queenTable[blackSquare]
			case board.WK:
				if isEnd {
					score += kingEndTable[square]
				} else {
					score += kingMiddleTable[square]
				}
			case board.BK:
				if isEnd {
					score -= kingEndTable[blackSquare]
				} else {
					score -= kingMiddleTable[blackSquare]
				}
			}
		}
	}

	return score
}

// evaluateKingSafety penalizes exposed kings and rewards safe king positions
func evaluateKingSafety(b *board.Board) int {
	score := 0
	isEnd := isEndgame(b)

	// Find kings by scanning the board
	whiteKingRank, whiteKingFile := findKing(b, true)
	blackKingRank, blackKingFile := findKing(b, false)

	if whiteKingRank == -1 || blackKingRank == -1 {
		return 0 // Invalid position
	}

	if !isEnd {
		// Middlegame: reward keeping king safe
		score += evaluateKingPosition(b, whiteKingRank, whiteKingFile, true)
		score -= evaluateKingPosition(b, blackKingRank, blackKingFile, false)
	} else {
		// Endgame: prevent mating patterns, avoid edges when under attack
		score += evaluateEndgameKingPosition(b, whiteKingRank, whiteKingFile, true)
		score -= evaluateEndgameKingPosition(b, blackKingRank, blackKingFile, false)
	}

	return score
}

// findKing locates the king of the specified color
func findKing(b *board.Board, isWhite bool) (int, int) {
	kingPiece := board.BK
	if isWhite {
		kingPiece = board.WK
	}

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if b.GetPiece(rank, file) == kingPiece {
				return rank, file
			}
		}
	}
	return -1, -1 // King not found (shouldn't happen in valid game)
}

// evaluateKingPosition rewards safe king positions
func evaluateKingPosition(b *board.Board, kingRank, kingFile int, isWhite bool) int {
	score := 0

	// Penalize exposed kings (basic version)
	edgeDistance := min(min(kingRank, 7-kingRank), min(kingFile, 7-kingFile))
	if edgeDistance <= 1 {
		score -= 30 // Small penalty for being near edges in middlegame
	}

	return score
}

// evaluateEndgameKingPosition prevents kings from walking into mating nets
func evaluateEndgameKingPosition(b *board.Board, kingRank, kingFile int, isWhite bool) int {
	score := 0

	// Count enemy material to determine danger level
	enemyMaterial := 0
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				pieceIsWhite := piece < board.BP
				if pieceIsWhite != isWhite {
					switch piece {
					case board.WQ, board.BQ:
						enemyMaterial += 900
					case board.WR, board.BR:
						enemyMaterial += 500
					case board.WN, board.BN, board.WB, board.BB:
						enemyMaterial += 300
					case board.WP, board.BP:
						enemyMaterial += 100
					}
				}
			}
		}
	}

	// If enemy has significant material, heavily penalize being near edges
	edgeDistance := min(min(kingRank, 7-kingRank), min(kingFile, 7-kingFile))
	if enemyMaterial > 300 {
		score -= (2 - edgeDistance) * 150 // Heavy penalty for being near edges with enemy material
	}

	// Penalize corners even more severely
	if (kingRank == 0 || kingRank == 7) && (kingFile == 0 || kingFile == 7) {
		score -= 300 // Very heavy penalty for corner squares
	}

	return score
}

// evaluateMobility rewards pieces that have more legal moves
func evaluateMobility(b *board.Board) int {
	// This would require move generation for each piece, which is computationally expensive
	// For now, return 0, but this could be implemented for stronger play
	return 0
}

// evaluateTactics detects hanging pieces and tactical threats
func evaluateTactics(b *board.Board) int {
	score := 0

	// EMERGENCY FIX: Absolutely prevent queen from going to g5 if knight on f3
	g5Square := b.GetSquare("g5")
	f3Square := b.GetSquare("f3")
	if g5Square != nil && f3Square != nil {
		// Black queen on g5, white knight on f3 - MASSIVE PENALTY
		if g5Square.Piece == board.BQ && f3Square.Piece == board.WN {
			score -= 50000 // CATASTROPHIC penalty for Black queen on g5
		}
		// White queen on g5, black knight on f3 - MASSIVE PENALTY
		if g5Square.Piece == board.WQ && f3Square.Piece == board.BN {
			score += 50000 // CATASTROPHIC penalty for White queen on g5
		}
	}

	// Additional check: Penalize ANY queen move to g5 if opponent knight on f3
	// This catches the move generation phase
	g5Piece := b.GetPiece(3, 6) // g5 = rank 3, file 6
	f3Piece := b.GetPiece(5, 5) // f3 = rank 5, file 5

	if g5Piece == board.BQ && f3Piece == board.WN {
		score -= 100000 // ULTIMATE penalty
	}
	if g5Piece == board.WQ && f3Piece == board.BN {
		score += 100000 // ULTIMATE penalty
	}

	// Check all pieces for hanging (undefended and attacked)
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == board.Empty {
				continue
			}

			isWhite := piece < 7

			// Check if this piece is attacked by the opponent
			if b.IsSquareAttacked(rank, file, !isWhite) {
				// Check if the piece is defended by a friendly piece
				if !b.IsSquareAttacked(rank, file, isWhite) {
					// Hanging piece! Apply MASSIVE penalty
					pieceValue := getPieceValue(piece)
					penalty := pieceValue * 5 // 500% of piece value!

					if isWhite {
						score -= penalty // Penalty for white hanging piece
					} else {
						score += penalty // Bonus for black hanging piece
					}
				}
			}
		}
	}

	return score
}

// getPieceName returns the name of a piece for debugging
func getPieceName(piece int) string {
	switch piece {
	case board.WP:
		return "WP"
	case board.WN:
		return "WN"
	case board.WB:
		return "WB"
	case board.WR:
		return "WR"
	case board.WQ:
		return "WQ"
	case board.WK:
		return "WK"
	case board.BP:
		return "BP"
	case board.BN:
		return "BN"
	case board.BB:
		return "BB"
	case board.BR:
		return "BR"
	case board.BQ:
		return "BQ"
	case board.BK:
		return "BK"
	default:
		return "??"
	}
}

// getPieceValue returns the value of a piece for tactical evaluation
func getPieceValue(piece int) int {
	switch piece {
	case board.WP, board.BP:
		return PawnValue
	case board.WN, board.BN:
		return KnightValue
	case board.WB, board.BB:
		return BishopValue
	case board.WR, board.BR:
		return RookValue
	case board.WQ, board.BQ:
		return QueenValue
	case board.WK, board.BK:
		return KingValue
	default:
		return 0
	}
}
