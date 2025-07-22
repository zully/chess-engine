package engine

import (
	"fmt"
	"math"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/moves"
)

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// moveInfo stores information needed to undo a move
type moveInfo struct {
	capturedPiece     int
	wasWhiteToMove    bool
	oldEnPassant      string
	enPassantSquare   *board.Square // Square where en passant pawn was removed
	oldCastlingRights int           // Castling rights before move
	positionHash      uint64        // Hash of position before move
	positionCount     int           // Count of position before move
}

// SearchResult contains the best move and its evaluation score
type SearchResult struct {
	BestMove moves.Move
	Score    int
}

// FindBestMove searches for the best move using iterative deepening and optimized search
func FindBestMove(b *board.Board, maxDepth int) SearchResult {
	// Check for opening book moves first
	if openingMove := getOpeningMove(b); openingMove != nil {
		return SearchResult{BestMove: *openingMove, Score: 0}
	}

	var bestMove moves.Move
	var bestScore int

	// Iterative deepening - search progressively deeper
	// This allows us to use results from shallower searches to order moves better
	for currentDepth := 1; currentDepth <= maxDepth; currentDepth++ {
		_, move, score := minimax(b, currentDepth, -math.MaxInt32, math.MaxInt32, b.WhiteToMove, bestMove)

		// Update best move and score
		bestMove = move
		bestScore = score

		// For shallow depths, continue immediately
		// For deeper depths, we have a good move if time becomes an issue
	}

	return SearchResult{
		BestMove: bestMove,
		Score:    bestScore,
	}
}

// minimax implements the minimax algorithm with alpha-beta pruning and move ordering
func minimax(b *board.Board, depth int, alpha, beta int, maximizingPlayer bool, prevBestMove moves.Move) (bool, moves.Move, int) {
	// Check for immediate draw conditions
	if b.IsDraw() {
		return true, moves.Move{}, 0 // Draw is always 0
	}

	// Base case: reached maximum depth or game over
	if depth == 0 {
		// Use quiescence search to avoid horizon effects
		score := quiesceSearch(b, alpha, beta, 0)
		return false, moves.Move{}, score
	}

	// Generate all legal moves
	legalMoves := GenerateMoves(b)

	// Order moves for better alpha-beta pruning
	legalMoves = orderMoves(legalMoves, b, prevBestMove)

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		if b.IsInCheck(b.WhiteToMove) {
			// Checkmate - very bad for current player
			if maximizingPlayer {
				return true, moves.Move{}, -10000 - depth // Prefer faster checkmates
			} else {
				return true, moves.Move{}, 10000 + depth
			}
		} else {
			// Stalemate
			return true, moves.Move{}, 0
		}
	}

	var bestMove moves.Move
	if maximizingPlayer {
		maxScore := -math.MaxInt32
		for _, move := range legalMoves {
			// Make the move and store undo info
			undoInfo := makeMove(b, move)
			if undoInfo == nil {
				continue
			}

			// Recursively evaluate position
			_, _, score := minimax(b, depth-1, alpha, beta, false, moves.Move{})

			// Undo the move
			undoMove(b, move, undoInfo)

			// Update best move if this is better
			if score > maxScore {
				maxScore = score
				bestMove = move
			}

			// Alpha-beta pruning
			alpha = max(alpha, score)
			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return false, bestMove, maxScore
	} else {
		minScore := math.MaxInt32
		for _, move := range legalMoves {
			// Make the move and store undo info
			undoInfo := makeMove(b, move)
			if undoInfo == nil {
				continue
			}

			// Recursively evaluate position
			_, _, score := minimax(b, depth-1, alpha, beta, true, moves.Move{})

			// Undo the move
			undoMove(b, move, undoInfo)

			// Update best move if this is better
			if score < minScore {
				minScore = score
				bestMove = move
			}

			// Alpha-beta pruning
			beta = min(beta, score)
			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return false, bestMove, minScore
	}
}

// ExecuteEngineMove directly executes an engine move on the board
func ExecuteEngineMove(b *board.Board, move moves.Move) error {
	fromSquare := b.GetSquare(move.From)
	toSquare := b.GetSquare(move.To)

	if fromSquare == nil || toSquare == nil {
		return fmt.Errorf("invalid square: from=%s, to=%s", move.From, move.To)
	}

	if fromSquare.Piece == board.Empty {
		return fmt.Errorf("no piece at %s", move.From)
	}

	// Verify it's the right color's turn
	isWhite := fromSquare.Piece < board.BP
	if isWhite != b.WhiteToMove {
		return fmt.Errorf("wrong color piece at %s", move.From)
	}

	// Execute the move
	capturedPiece := toSquare.Piece
	movedPiece := fromSquare.Piece
	toSquare.Piece = movedPiece
	fromSquare.Piece = board.Empty

	// Handle en passant capture - remove the captured pawn
	if move.EnPassant {
		toRank := 7 - int(move.To[1]-'1')
		toFile := int(move.To[0] - 'a')

		capturedPawnRank := toRank
		if movedPiece == board.WP {
			capturedPawnRank = toRank + 1 // White captures black pawn one rank below
		} else {
			capturedPawnRank = toRank - 1 // Black captures white pawn one rank above
		}
		capturedPawnSquare := b.GetSquareByCoords(capturedPawnRank, toFile)
		if capturedPawnSquare != nil {
			capturedPiece = capturedPawnSquare.Piece // For notation purposes
			capturedPawnSquare.Piece = board.Empty
		}
	}

	// Check for pawn promotion
	if move.Promote != "" {
		// Convert promotion piece string to piece constant
		var promotedPiece int
		isWhitePiece := movedPiece == board.WP
		switch move.Promote {
		case "Q":
			if isWhitePiece {
				promotedPiece = board.WQ
			} else {
				promotedPiece = board.BQ
			}
		case "R":
			if isWhitePiece {
				promotedPiece = board.WR
			} else {
				promotedPiece = board.BR
			}
		case "B":
			if isWhitePiece {
				promotedPiece = board.WB
			} else {
				promotedPiece = board.BB
			}
		case "N":
			if isWhitePiece {
				promotedPiece = board.WN
			} else {
				promotedPiece = board.BN
			}
		default:
			// Default to Queen
			if isWhitePiece {
				promotedPiece = board.WQ
			} else {
				promotedPiece = board.BQ
			}
		}
		toSquare.Piece = promotedPiece
		fmt.Printf("Pawn promoted to %s!\n", move.Promote)
	}

	// Switch turns
	b.WhiteToMove = !b.WhiteToMove

	// Check if opponent (whose turn it now is) is in check after this move
	opponentInCheck := b.IsInCheck(b.WhiteToMove)
	opponentInCheckmate := false
	if opponentInCheck {
		opponentInCheckmate = b.IsCheckmate(b.WhiteToMove)
	}

	// Display check/checkmate messages
	if opponentInCheck {
		if opponentInCheckmate {
			if b.WhiteToMove {
				fmt.Println("Checkmate! White is checkmated!")
			} else {
				fmt.Println("Checkmate! Black is checkmated!")
			}
		} else {
			if b.WhiteToMove {
				fmt.Println("White is in check!")
			} else {
				fmt.Println("Black is in check!")
			}
		}
	}

	// Build notation for move history display
	notation := ""
	if move.Piece != "P" {
		notation += move.Piece
		// Add disambiguation if needed
		disambiguation := getDisambiguation(b, move)
		notation += disambiguation
	}
	if capturedPiece != board.Empty {
		if move.Piece == "P" {
			notation += move.From[0:1] + "x"
		} else {
			notation += "x"
		}
	}
	notation += move.To

	// Add promotion notation
	if move.Promote != "" {
		notation += "=" + move.Promote
	}

	// Add check/checkmate notation
	if opponentInCheck {
		if opponentInCheckmate {
			notation += "#"
		} else {
			notation += "+"
		}
	}

	// Add to move history
	b.MovesPlayed = append(b.MovesPlayed, notation)

	return nil
}

// makeMove executes a move on the board and returns undo info or nil if failed
func makeMove(b *board.Board, move moves.Move) *moveInfo {
	fromSquare := b.GetSquare(move.From)
	toSquare := b.GetSquare(move.To)

	if fromSquare == nil || toSquare == nil || fromSquare.Piece == board.Empty {
		return nil
	}

	// Store move info for undo, including position history state
	positionHash := b.GetPositionHash()
	positionCount := b.PositionHistory[positionHash]

	undoInfo := &moveInfo{
		capturedPiece:     toSquare.Piece,
		wasWhiteToMove:    b.WhiteToMove,
		oldEnPassant:      b.EnPassant,
		oldCastlingRights: b.CastlingRights,
		positionHash:      positionHash,
		positionCount:     positionCount,
	}

	// Clear en passant target
	b.EnPassant = ""

	// Update castling rights if king or rook moves
	// We need to create a temporary method call - let me add a helper
	piece := fromSquare.Piece
	switch move.From {
	case "e1": // White king
		if piece == board.WK {
			b.CastlingRights &^= 3 // Remove both white castling rights
		}
	case "a1": // White queenside rook
		if piece == board.WR {
			b.CastlingRights &^= 2 // Remove white queenside
		}
	case "h1": // White kingside rook
		if piece == board.WR {
			b.CastlingRights &^= 1 // Remove white kingside
		}
	case "e8": // Black king
		if piece == board.BK {
			b.CastlingRights &^= 12 // Remove both black castling rights
		}
	case "a8": // Black queenside rook
		if piece == board.BR {
			b.CastlingRights &^= 8 // Remove black queenside
		}
	case "h8": // Black kingside rook
		if piece == board.BR {
			b.CastlingRights &^= 4 // Remove black kingside
		}
	}

	// Handle en passant capture - remove the captured pawn
	if move.EnPassant {
		toRank := 7 - int(move.To[1]-'1')
		toFile := int(move.To[0] - 'a')

		capturedPawnRank := toRank
		if fromSquare.Piece == board.WP {
			capturedPawnRank = toRank + 1
		} else {
			capturedPawnRank = toRank - 1
		}
		capturedPawnSquare := b.GetSquareByCoords(capturedPawnRank, toFile)
		if capturedPawnSquare != nil {
			undoInfo.enPassantSquare = capturedPawnSquare
			undoInfo.capturedPiece = capturedPawnSquare.Piece
			capturedPawnSquare.Piece = board.Empty
		}
	}

	// Execute the move
	toSquare.Piece = fromSquare.Piece
	fromSquare.Piece = board.Empty

	// Check for pawn two-square move to set new en passant target
	if move.Piece == "P" && fromSquare != nil && toSquare != nil {
		fromRank := 7 - int(move.From[1]-'1')
		toRank := 7 - int(move.To[1]-'1')
		if abs(toRank-fromRank) == 2 {
			targetRank := (fromRank + toRank) / 2
			targetFile := int(move.To[0] - 'a')
			b.EnPassant = board.GetSquareName(targetRank, targetFile)
		}
	}

	// Switch turns
	b.WhiteToMove = !b.WhiteToMove

	// Record the new position for repetition detection during search
	b.RecordPosition()

	return undoInfo
}

// undoMove reverses the move using the provided undo info
func undoMove(b *board.Board, move moves.Move, undoInfo *moveInfo) {
	fromSquare := b.GetSquare(move.From)
	toSquare := b.GetSquare(move.To)

	if fromSquare == nil || toSquare == nil || undoInfo == nil {
		return
	}

	// Remove the current position from history
	currentHash := b.GetPositionHash()
	if count, exists := b.PositionHistory[currentHash]; exists && count > 0 {
		if count == 1 {
			delete(b.PositionHistory, currentHash)
		} else {
			b.PositionHistory[currentHash] = count - 1
		}
	}

	// Restore the board position
	fromSquare.Piece = toSquare.Piece       // Move piece back
	toSquare.Piece = undoInfo.capturedPiece // Restore captured piece

	// If this was an en passant capture, restore the captured pawn
	if move.EnPassant && undoInfo.enPassantSquare != nil {
		undoInfo.enPassantSquare.Piece = undoInfo.capturedPiece
		toSquare.Piece = board.Empty // The destination square should be empty for en passant
	}

	// Restore en passant target
	b.EnPassant = undoInfo.oldEnPassant

	// Restore castling rights
	b.CastlingRights = undoInfo.oldCastlingRights

	// Restore turn
	b.WhiteToMove = undoInfo.wasWhiteToMove

	// Restore position history count
	if undoInfo.positionCount > 0 {
		b.PositionHistory[undoInfo.positionHash] = undoInfo.positionCount
	} else {
		delete(b.PositionHistory, undoInfo.positionHash)
	}
}

// orderMoves reorders moves to improve alpha-beta pruning effectiveness
func orderMoves(moveList []moves.Move, b *board.Board, prevBestMove moves.Move) []moves.Move {
	// Advanced move ordering: prioritize best moves first for better pruning
	var bestMove []moves.Move
	var winningCaptures []moves.Move
	var equalCaptures []moves.Move
	var checks []moves.Move
	var regularMoves []moves.Move

	for _, move := range moveList {
		// Prioritize previous best move first
		if movesEqual(move, prevBestMove) {
			bestMove = append(bestMove, move)
			continue
		}

		// Check if it's a capture
		toSquare := b.GetSquare(move.To)
		if toSquare != nil && toSquare.Piece != board.Empty {
			fromSquare := b.GetSquare(move.From)
			if fromSquare != nil {
				capturedValue := getPieceValueForOrdering(toSquare.Piece)
				capturingValue := getPieceValueForOrdering(fromSquare.Piece)

				// Good captures: capture higher value with lower value piece
				if capturedValue >= capturingValue {
					winningCaptures = append(winningCaptures, move)
				} else {
					equalCaptures = append(equalCaptures, move)
				}
			}
			continue
		}

		// Check if it gives check (expensive but important for move ordering)
		undoInfo := makeMove(b, move)
		if undoInfo != nil {
			givesCheck := b.IsInCheck(b.WhiteToMove)
			undoMove(b, move, undoInfo)

			if givesCheck {
				checks = append(checks, move)
				continue
			}
		}

		// Regular moves
		regularMoves = append(regularMoves, move)
	}

	// Order: Previous best, winning captures, checks, equal captures, regular moves
	var orderedMoves []moves.Move
	orderedMoves = append(orderedMoves, bestMove...)
	orderedMoves = append(orderedMoves, winningCaptures...)
	orderedMoves = append(orderedMoves, checks...)
	orderedMoves = append(orderedMoves, equalCaptures...)
	orderedMoves = append(orderedMoves, regularMoves...)

	return orderedMoves
}

// getPieceValueForOrdering returns piece value for move ordering
func getPieceValueForOrdering(piece int) int {
	switch piece {
	case board.WP, board.BP:
		return 1
	case board.WN, board.BN, board.WB, board.BB:
		return 3
	case board.WR, board.BR:
		return 5
	case board.WQ, board.BQ:
		return 9
	case board.WK, board.BK:
		return 100
	default:
		return 0
	}
}

// movesEqual compares two moves for equality
func movesEqual(a, b moves.Move) bool {
	return a.From == b.From && a.To == b.To && a.Piece == b.Piece
}

// quiesceSearch searches only tactical moves (captures, checks) until position is quiet
// This prevents horizon effects where the engine misses tactics just beyond search depth
func quiesceSearch(b *board.Board, alpha, beta int, qDepth int) int {
	// Limit quiescence search depth to prevent infinite loops
	if qDepth > 8 {
		return Evaluate(b)
	}

	// Get static evaluation as baseline
	standPat := Evaluate(b)

	// Beta cutoff - this position is already too good for the opponent
	if standPat >= beta {
		return beta
	}

	// Update alpha if we can improve it
	if standPat > alpha {
		alpha = standPat
	}

	// Generate only tactical moves (captures and checks)
	tacticalMoves := generateTacticalMoves(b)

	// If no tactical moves, return the static evaluation
	if len(tacticalMoves) == 0 {
		return standPat
	}

	// Search all tactical moves
	for _, move := range tacticalMoves {
		undoInfo := makeMove(b, move)
		if undoInfo == nil {
			continue
		}

		// Recursively search this tactical line
		score := -quiesceSearch(b, -beta, -alpha, qDepth+1)

		undoMove(b, move, undoInfo)

		if score >= beta {
			return beta // Beta cutoff
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

// generateTacticalMoves generates only captures and checks for quiescence search
func generateTacticalMoves(b *board.Board) []moves.Move {
	allMoves := GenerateMoves(b)
	var tacticalMoves []moves.Move

	for _, move := range allMoves {
		// Check if it's a capture
		toSquare := b.GetSquare(move.To)
		if toSquare != nil && toSquare.Piece != board.Empty {
			tacticalMoves = append(tacticalMoves, move)
			continue
		}

		// Check if it gives check (expensive but important)
		undoInfo := makeMove(b, move)
		if undoInfo != nil {
			givesCheck := b.IsInCheck(b.WhiteToMove)
			undoMove(b, move, undoInfo)

			if givesCheck {
				tacticalMoves = append(tacticalMoves, move)
			}
		}
	}

	return tacticalMoves
}

// Global variable to store the last move for undo (simple implementation)
var lastMove moveInfo

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getOpeningMove returns a good opening move if position matches known patterns
func getOpeningMove(b *board.Board) *moves.Move {
	moveCount := len(b.MovesPlayed)

	// Only apply opening knowledge in first few moves
	if moveCount > 6 {
		return nil
	}

	// White opening moves
	if b.WhiteToMove {
		if moveCount == 0 {
			// First move: play e4 or d4
			return &moves.Move{From: "e2", To: "e4", Piece: "P"}
		}
		if moveCount == 2 {
			// Second move: develop knight to f3
			piece := b.GetPiece(7, 6) // g1 square
			if piece == board.WN {
				return &moves.Move{From: "g1", To: "f3", Piece: "N"}
			}
		}
	} else {
		// Black responses
		if moveCount == 1 {
			// Respond to e4 with e5, to d4 with d5
			if len(b.MovesPlayed) > 0 && b.MovesPlayed[0] == "e4" {
				return &moves.Move{From: "e7", To: "e5", Piece: "P"}
			}
			if len(b.MovesPlayed) > 0 && b.MovesPlayed[0] == "d4" {
				return &moves.Move{From: "d7", To: "d5", Piece: "P"}
			}
		}
	}

	return nil
}

// getDisambiguation returns the file/rank needed to disambiguate moves when multiple pieces can reach the same square
func getDisambiguation(b *board.Board, move moves.Move) string {
	// Only non-pawn pieces need disambiguation
	if move.Piece == "P" {
		return ""
	}

	// Queens and Kings NEVER need disambiguation - there's only one per side!
	if move.Piece == "Q" || move.Piece == "K" {
		return ""
	}

	// Get source and destination coordinates
	fromRank := 7 - int(move.From[1]-'1')
	fromFile := int(move.From[0] - 'a')
	toRank := 7 - int(move.To[1]-'1')
	toFile := int(move.To[0] - 'a')

	// Find the piece type and color
	movedPiece := b.GetPiece(fromRank, fromFile)
	isWhite := movedPiece < board.BP

	// Find other pieces of the same type that could also move to the destination
	var ambiguousPieces []struct{ rank, file int }

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			// Skip the piece that's actually moving
			if rank == fromRank && file == fromFile {
				continue
			}

			piece := b.GetPiece(rank, file)

			// Check if this is the same type of piece and same color
			if piece == movedPiece {
				// Check if this piece could legally move to the destination
				canMove := false
				switch move.Piece {
				case "N":
					canMove = isValidKnightMove(rank, file, toRank, toFile)
				case "B":
					canMove = isValidBishopMove(b, rank, file, toRank, toFile)
				case "R":
					canMove = isValidRookMove(b, rank, file, toRank, toFile)
				case "Q":
					canMove = isValidQueenMove(b, rank, file, toRank, toFile)
				case "K":
					canMove = isValidKingMove(rank, file, toRank, toFile)
				}

				// Also check that the destination square is valid (empty or enemy piece)
				if canMove {
					targetPiece := b.GetPiece(toRank, toFile)
					if targetPiece == board.Empty || (targetPiece < board.BP) != isWhite {
						ambiguousPieces = append(ambiguousPieces, struct{ rank, file int }{rank, file})
					}
				}
			}
		}
	}

	// If no ambiguous pieces, no disambiguation needed
	if len(ambiguousPieces) == 0 {
		return ""
	}

	// Check if file disambiguation is sufficient
	fileSufficient := true
	for _, piece := range ambiguousPieces {
		if piece.file == fromFile {
			fileSufficient = false
			break
		}
	}

	if fileSufficient {
		// Use file disambiguation (e.g., "Rad1")
		return string(rune('a' + fromFile))
	}

	// Check if rank disambiguation is sufficient
	rankSufficient := true
	for _, piece := range ambiguousPieces {
		if piece.rank == fromRank {
			rankSufficient = false
			break
		}
	}

	if rankSufficient {
		// Use rank disambiguation (e.g., "R1d4")
		return string(rune('1' + (7 - fromRank)))
	}

	// Need both file and rank disambiguation (e.g., "Ra1d4")
	return string(rune('a'+fromFile)) + string(rune('1'+(7-fromRank)))
}

// Helper functions for piece movement validation (used for disambiguation)

func isValidKnightMove(startRank, startFile, endRank, endFile int) bool {
	rankDiff := absInt(endRank - startRank)
	fileDiff := absInt(endFile - startFile)
	return (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2)
}

func isValidBishopMove(b *board.Board, fromRank, fromFile, toRank, toFile int) bool {
	// Must move diagonally
	rankDiff := absInt(toRank - fromRank)
	fileDiff := absInt(toFile - fromFile)
	if rankDiff != fileDiff {
		return false
	}

	// Check path for obstacles
	rankStep := 1
	if toRank < fromRank {
		rankStep = -1
	}
	fileStep := 1
	if toFile < fromFile {
		fileStep = -1
	}

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

func isValidRookMove(b *board.Board, fromRank, fromFile, toRank, toFile int) bool {
	// Must move horizontally or vertically
	if fromRank != toRank && fromFile != toFile {
		return false
	}

	// Check path for obstacles
	if fromRank == toRank {
		// Horizontal move
		step := 1
		if toFile < fromFile {
			step = -1
		}
		for file := fromFile + step; file != toFile; file += step {
			if !b.IsSquareEmpty(fromRank, file) {
				return false
			}
		}
	} else {
		// Vertical move
		step := 1
		if toRank < fromRank {
			step = -1
		}
		for rank := fromRank + step; rank != toRank; rank += step {
			if !b.IsSquareEmpty(rank, fromFile) {
				return false
			}
		}
	}

	return true
}

func isValidQueenMove(b *board.Board, fromRank, fromFile, toRank, toFile int) bool {
	// Queen combines bishop and rook moves
	return isValidBishopMove(b, fromRank, fromFile, toRank, toFile) ||
		isValidRookMove(b, fromRank, fromFile, toRank, toFile)
}

func isValidKingMove(startRank, startFile, endRank, endFile int) bool {
	rankDiff := absInt(endRank - startRank)
	fileDiff := absInt(endFile - startFile)
	return rankDiff <= 1 && fileDiff <= 1
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
