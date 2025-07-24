package game

import (
	"fmt"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/uci"
)

// GameState represents the complete state of a chess game
type GameState struct {
	Board            *board.Board    `json:"board"`
	Message          string          `json:"message"`
	Error            string          `json:"error,omitempty"`
	GameOver         bool            `json:"gameOver"`
	InCheck          bool            `json:"inCheck"`
	IsCheckmate      bool            `json:"isCheckmate"`
	Draw             bool            `json:"draw"`
	DrawReason       string          `json:"drawReason"`
	ThreefoldRep     bool            `json:"threefoldRepetition"`
	PositionCount    int             `json:"positionCount"`
	Evaluation       int             `json:"evaluation"`       // Position evaluation in centipawns
	CapturedWhite    []CapturedPiece `json:"capturedWhite"`    // Pieces captured by White
	CapturedBlack    []CapturedPiece `json:"capturedBlack"`    // Pieces captured by Black
	StockfishVersion string          `json:"stockfishVersion"` // Stockfish engine version
	LastUCIMove      string          `json:"lastUCIMove"`      // Last UCI move played
}

// CapturedPiece represents a captured piece with its value
type CapturedPiece struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// EngineRequest represents a request to the chess engine
type EngineRequest struct {
	Depth int `json:"depth,omitempty"`
	Elo   int `json:"elo,omitempty"` // Target ELO rating (1350-2850, 0 = full strength)
}

// GetCapturedPieces analyzes the board and returns lists of captured pieces
func GetCapturedPieces(gameBoard *board.Board) ([]CapturedPiece, []CapturedPiece) {
	// Initial piece counts for a standard chess game
	initialCounts := map[int]int{
		board.WP: 8, board.WN: 2, board.WB: 2, board.WR: 2, board.WQ: 1, board.WK: 1,
		board.BP: 8, board.BN: 2, board.BB: 2, board.BR: 2, board.BQ: 1, board.BK: 1,
	}

	// Count current pieces on the board
	currentCounts := make(map[int]int)
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := gameBoard.GetPiece(rank, file)
			if piece != board.Empty {
				currentCounts[piece]++
			}
		}
	}

	var capturedWhite []CapturedPiece // Pieces captured by White (black pieces taken)
	var capturedBlack []CapturedPiece // Pieces captured by Black (white pieces taken)

	// Check what pieces are missing (captured)
	for pieceType, initialCount := range initialCounts {
		currentCount := currentCounts[pieceType]
		capturedCount := initialCount - currentCount

		if capturedCount > 0 {
			pieceTypeStr := board.GetPieceType(pieceType)
			pieceValue := board.GetPieceValue(pieceType)

			// Add each captured piece individually to the appropriate list
			for i := 0; i < capturedCount; i++ {
				capturedPiece := CapturedPiece{
					Type:  pieceTypeStr,
					Value: pieceValue,
				}

				// If it's a white piece that's missing, black captured it
				// If it's a black piece that's missing, white captured it
				if pieceType < board.BP { // White piece captured by black
					capturedBlack = append(capturedBlack, capturedPiece)
				} else { // Black piece captured by white
					capturedWhite = append(capturedWhite, capturedPiece)
				}
			}
		}
	}

	return capturedWhite, capturedBlack
}

// CreateCompleteGameState creates a complete game state with all necessary information
func CreateCompleteGameState(gameBoard *board.Board, message string, evaluation int, stockfishEngine *uci.Engine) GameState {
	capturedWhite, capturedBlack := GetCapturedPieces(gameBoard)

	// Ensure arrays are never nil
	if capturedWhite == nil {
		capturedWhite = []CapturedPiece{}
	}
	if capturedBlack == nil {
		capturedBlack = []CapturedPiece{}
	}

	// Get Stockfish version
	stockfishVersion := "Not Available"
	if stockfishEngine != nil {
		if version, err := stockfishEngine.GetEngineInfo(); err == nil {
			stockfishVersion = version
		}
	}

	state := GameState{
		Board:            gameBoard,
		Message:          message,
		Evaluation:       evaluation,
		CapturedWhite:    capturedWhite,
		CapturedBlack:    capturedBlack,
		StockfishVersion: stockfishVersion,
	}

	// Update check/checkmate status
	state.InCheck = gameBoard.IsInCheck(gameBoard.WhiteToMove)
	state.IsCheckmate = gameBoard.IsCheckmate(gameBoard.WhiteToMove)
	state.GameOver = state.IsCheckmate

	// Check for draws
	isDraw := gameBoard.IsDraw()
	drawReason := ""
	if isDraw {
		if gameBoard.IsThreefoldRepetition() {
			drawReason = "Threefold repetition"
		} else {
			drawReason = "Stalemate"
		}
	}
	state.Draw = isDraw
	state.DrawReason = drawReason
	state.GameOver = state.IsCheckmate || isDraw
	state.ThreefoldRep = gameBoard.IsThreefoldRepetition()
	if gameBoard.PositionHistory != nil {
		for _, count := range gameBoard.PositionHistory {
			if count > state.PositionCount {
				state.PositionCount = count
			}
		}
	}

	// Enhance message with check/checkmate announcements
	if state.IsCheckmate {
		if gameBoard.WhiteToMove {
			state.Message = "Checkmate! Black wins!"
		} else {
			state.Message = "Checkmate! White wins!"
		}
	} else if state.InCheck {
		if gameBoard.WhiteToMove {
			state.Message = "White is in check!"
		} else {
			state.Message = "Black is in check!"
		}
	} else if state.Draw {
		state.Message = fmt.Sprintf("Draw! %s", drawReason)
	} else if message == "" {
		if gameBoard.WhiteToMove {
			state.Message = "White to move"
		} else {
			state.Message = "Black to move"
		}
	}

	return state
} 