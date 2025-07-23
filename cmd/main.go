package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/uci"
)

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
}

type CapturedPiece struct {
	Type  string `json:"type"`  // Piece type (P, N, B, R, Q)
	Value int    `json:"value"` // Point value
}

type MoveRequest struct {
	Move string `json:"move"`
}

type EngineRequest struct {
	Depth int `json:"depth,omitempty"`
	Elo   int `json:"elo,omitempty"` // Target ELO rating (1350-2850, 0 = full strength)
}

var gameBoard *board.Board
var stockfishEngine *uci.Engine

func main() {
	// Initialize the game board
	gameBoard = board.NewBoard()

	// Initialize Stockfish engine
	stockfishPath := filepath.Join("stockfish", "stockfish-macos-m1-apple-silicon")
	var err error
	stockfishEngine, err = uci.NewEngine(stockfishPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize Stockfish engine: %v", err)
		log.Println("Engine features will be disabled")
	}

	// Serve static files (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API endpoints
	http.HandleFunc("/api/state", getGameState)
	http.HandleFunc("/api/move", makeMove)
	http.HandleFunc("/api/engine", engineMove)
	http.HandleFunc("/api/undo", undoMove)
	http.HandleFunc("/api/reset", resetGame)

	// Main page
	http.HandleFunc("/", homePage)

	fmt.Println("Chess Web GUI with Stockfish starting on http://localhost:8080")
	if stockfishEngine != nil {
		fmt.Println("Stockfish engine initialized successfully")
	} else {
		fmt.Println("Running without engine (moves disabled)")
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Serve the HTML template file
	http.ServeFile(w, r, "web/templates/index.html")
}

func getGameState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current position evaluation from Stockfish if available
	evaluation := 0
	if stockfishEngine != nil {
		currentFEN := gameBoard.ToFEN()
		if eval, err := stockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		}
	}

	// Create complete game state
	message := "Ready to play"
	if gameBoard.WhiteToMove {
		message = "White to move"
	} else {
		message = "Black to move"
	}

	state := createCompleteGameState(gameBoard, message, evaluation)
	json.NewEncoder(w).Encode(state)
}

func makeMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		state := GameState{Board: gameBoard, Error: "Invalid request format"}
		json.NewEncoder(w).Encode(state)
		return
	}

	err := gameBoard.MakeMove(req.Move)

	if err != nil {
		state := GameState{Board: gameBoard, Error: err.Error()}
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get current position evaluation from Stockfish if available
	evaluation := 0
	if stockfishEngine != nil {
		currentFEN := gameBoard.ToFEN()
		if eval, err := stockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		} else {
			fmt.Printf("Warning: Failed to get evaluation: %v\n", err)
		}
	}

	// Create message based on game state
	baseMessage := fmt.Sprintf("Played %s", req.Move)

	// Create complete game state with evaluation
	state := createCompleteGameState(gameBoard, baseMessage, evaluation)
	json.NewEncoder(w).Encode(state)
}

func engineMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EngineRequest
	json.NewDecoder(r.Body).Decode(&req)

	state := GameState{Board: gameBoard}

	// Check if Stockfish engine is available
	if stockfishEngine == nil {
		state.Error = "Stockfish engine not available"
		json.NewEncoder(w).Encode(state)
		return
	}

	// Set depth (default to 6 if not specified)
	depth := 6
	if req.Depth > 0 && req.Depth <= 15 {
		depth = req.Depth
	}

	// Set ELO/strength if specified
	if req.Elo > 0 {
		if req.Elo >= 1350 && req.Elo <= 2850 {
			err := stockfishEngine.SetEloRating(req.Elo)
			if err != nil {
				fmt.Printf("Warning: Failed to set ELO rating: %v\n", err)
			} else {
				fmt.Printf("Set Stockfish ELO to %d\n", req.Elo)
			}
		} else {
			fmt.Printf("Warning: Invalid ELO rating %d (must be 1350-2850), using default strength\n", req.Elo)
			// Use default strength when invalid ELO is provided
			err := stockfishEngine.DisableStrengthLimit()
			if err != nil {
				fmt.Printf("Warning: Failed to disable strength limit: %v\n", err)
			}
		}
	} else {
		// Full strength (disable ELO limiting)
		err := stockfishEngine.DisableStrengthLimit()
		if err != nil {
			fmt.Printf("Warning: Failed to disable strength limit: %v\n", err)
		}
	}

	// Set current position in Stockfish using FEN
	fen := gameBoard.ToFEN()
	err := stockfishEngine.SetPosition(fen)
	if err != nil {
		state.Error = fmt.Sprintf("Failed to set position: %v", err)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get best move from Stockfish
	currentFEN := gameBoard.ToFEN()

	bestMove, err := stockfishEngine.GetBestMove(currentFEN, depth)
	if err != nil {
		state.Error = fmt.Sprintf("Engine error: %v", err)
		json.NewEncoder(w).Encode(state)
		return
	}

	if bestMove == nil {
		state.Error = "No move received from engine"
		json.NewEncoder(w).Encode(state)
		return
	}

	// Convert UCI move to algebraic notation and execute it
	// First, we need to determine what piece is moving by examining the board
	fromRank, fromFile := board.GetSquareCoords(bestMove.From)
	if fromRank < 0 || fromFile < 0 {
		state.Error = fmt.Sprintf("Invalid UCI move from square: %s", bestMove.From)
		json.NewEncoder(w).Encode(state)
		return
	}

	toRank, toFile := board.GetSquareCoords(bestMove.To)
	if toRank < 0 || toFile < 0 {
		state.Error = fmt.Sprintf("Invalid UCI move to square: %s", bestMove.To)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get the piece that's moving
	piece := gameBoard.GetPiece(fromRank, fromFile)
	if piece == board.Empty {
		state.Error = fmt.Sprintf("No piece at source square %s", bestMove.From)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Check if it's a capture
	targetPiece := gameBoard.GetPiece(toRank, toFile)
	isCapture := targetPiece != board.Empty

	// Convert UCI to algebraic notation based on piece type
	var moveNotation string
	switch piece {
	case board.WP, board.BP:
		// Pawn moves
		if isCapture {
			moveNotation = bestMove.From[:1] + "x" + bestMove.To
		} else {
			moveNotation = bestMove.To
		}
	case board.WN, board.BN:
		// Knight moves - check for disambiguation
		moveNotation = "N"
		if needsDisambiguation(gameBoard, piece, fromRank, fromFile, toRank, toFile) {
			moveNotation += bestMove.From[:1] // Add file for disambiguation
		}
		if isCapture {
			moveNotation += "x"
		}
		moveNotation += bestMove.To
	case board.WB, board.BB:
		// Bishop moves - check for disambiguation
		moveNotation = "B"
		if needsDisambiguation(gameBoard, piece, fromRank, fromFile, toRank, toFile) {
			moveNotation += bestMove.From[:1] // Add file for disambiguation
		}
		if isCapture {
			moveNotation += "x"
		}
		moveNotation += bestMove.To
	case board.WR, board.BR:
		// Rook moves - check for disambiguation
		moveNotation = "R"
		if needsDisambiguation(gameBoard, piece, fromRank, fromFile, toRank, toFile) {
			moveNotation += bestMove.From[:1] // Add file for disambiguation
		}
		if isCapture {
			moveNotation += "x"
		}
		moveNotation += bestMove.To
	case board.WQ, board.BQ:
		// Queen moves - check for disambiguation
		moveNotation = "Q"
		if needsDisambiguation(gameBoard, piece, fromRank, fromFile, toRank, toFile) {
			moveNotation += bestMove.From[:1] // Add file for disambiguation
		}
		if isCapture {
			moveNotation += "x"
		}
		moveNotation += bestMove.To
	case board.WK, board.BK:
		// Check for castling
		if bestMove.From == "e1" && (bestMove.To == "g1" || bestMove.To == "c1") ||
			bestMove.From == "e8" && (bestMove.To == "g8" || bestMove.To == "c8") {
			if bestMove.To[0] == 'g' {
				moveNotation = "O-O"
			} else {
				moveNotation = "O-O-O"
			}
		} else {
			moveNotation = "K"
			if isCapture {
				moveNotation += "x"
			}
			moveNotation += bestMove.To
		}
	default:
		state.Error = fmt.Sprintf("Unknown piece type: %d", piece)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Execute the move
	err = gameBoard.MakeMove(moveNotation)
	if err != nil {
		state.Error = fmt.Sprintf("Failed to execute engine move %s (%s): %v", bestMove.UCI, moveNotation, err)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Update game state
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

	// Set message with engine evaluation
	baseMessage := fmt.Sprintf("Stockfish played %s (depth: %d, score: %d)",
		moveNotation, bestMove.Depth, bestMove.Score)

	if isDraw {
		baseMessage += fmt.Sprintf(" - Draw! %s", drawReason)
	} else if state.IsCheckmate {
		if gameBoard.WhiteToMove {
			baseMessage += " - Black wins!"
		} else {
			baseMessage += " - White wins!"
		}
	} else if state.InCheck {
		if gameBoard.WhiteToMove {
			baseMessage += " - White in check!"
		} else {
			baseMessage += " - Black in check!"
		}
	}

	// Create complete game state with evaluation
	state = createCompleteGameState(gameBoard, baseMessage, bestMove.Evaluation)
	json.NewEncoder(w).Encode(state)
}

func undoMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if there are moves to undo
	if len(gameBoard.MovesPlayed) == 0 {
		state := GameState{
			Board: gameBoard,
			Error: "No moves to undo!",
		}
		json.NewEncoder(w).Encode(state)
		return
	}

	// Store the current moves list
	currentMoves := make([]string, len(gameBoard.MovesPlayed))
	copy(currentMoves, gameBoard.MovesPlayed)

	// Remove the last move
	movesToReplay := currentMoves[:len(currentMoves)-1]

	// Create a fresh board
	gameBoard = board.NewBoard()

	// Replay all moves except the last one
	for _, move := range movesToReplay {
		err := gameBoard.MakeMove(move)
		if err != nil {
			// If replay fails, restore the original board state
			// This shouldn't happen, but just in case
			gameBoard = board.NewBoard()
			for _, originalMove := range currentMoves {
				gameBoard.MakeMove(originalMove)
			}
			state := GameState{
				Board: gameBoard,
				Error: fmt.Sprintf("Failed to undo move: %v", err),
			}
			json.NewEncoder(w).Encode(state)
			return
		}
	}

	// Create and return the updated game state
	inCheck := gameBoard.IsInCheck(gameBoard.WhiteToMove)
	isCheckmate := false
	if inCheck {
		isCheckmate = gameBoard.IsCheckmate(gameBoard.WhiteToMove)
	}

	isDraw := gameBoard.IsDraw()
	drawReason := ""
	if isDraw {
		if gameBoard.IsThreefoldRepetition() {
			drawReason = "Threefold repetition"
		} else {
			drawReason = "Stalemate"
		}
	}

	state := GameState{
		Board:         gameBoard,
		InCheck:       inCheck,
		IsCheckmate:   isCheckmate,
		GameOver:      isCheckmate || isDraw,
		Draw:          isDraw,
		DrawReason:    drawReason,
		ThreefoldRep:  gameBoard.IsThreefoldRepetition(),
		PositionCount: gameBoard.GetPositionCount(),
	}

	lastMove := currentMoves[len(currentMoves)-1]
	state.Message = fmt.Sprintf("Undid move %s", lastMove)

	json.NewEncoder(w).Encode(state)
}

func resetGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resetBtn := "reset-btn"
	_ = resetBtn // Avoid unused variable

	// Create a new board
	gameBoard = board.NewBoard()

	// Get initial evaluation
	evaluation := 0
	if stockfishEngine != nil {
		currentFEN := gameBoard.ToFEN()
		if eval, err := stockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		}
	}

	// Create complete game state with evaluation
	state := createCompleteGameState(gameBoard, "Game reset. White to move.", evaluation)
	json.NewEncoder(w).Encode(state)
}

// needsDisambiguation checks if there are other pieces of the same type that could move to the same destination
func needsDisambiguation(b *board.Board, pieceType int, fromRank, fromFile, toRank, toFile int) bool {
	// Check for pieces of the same type that could also move to the destination
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if rank == fromRank && file == fromFile {
				continue // Skip the piece we're moving
			}

			if b.GetPiece(rank, file) != pieceType {
				continue // Not the same piece type
			}

			// Check if this piece could also move to the same destination
			canMove := false
			switch pieceType {
			case board.WN, board.BN:
				canMove = board.CanKnightMove(rank, file, toRank, toFile)
			case board.WB, board.BB:
				canMove = board.CanBishopMove(b, rank, file, toRank, toFile)
			case board.WR, board.BR:
				canMove = board.CanRookMove(b, rank, file, toRank, toFile)
			case board.WQ, board.BQ:
				canMove = board.CanQueenMove(b, rank, file, toRank, toFile)
			}

			if canMove {
				// Check if the destination square is valid (not capturing own piece)
				targetPiece := b.GetPiece(toRank, toFile)
				if targetPiece != board.Empty {
					// Different color piece - can capture
					if (targetPiece >= board.BP) != (pieceType >= board.BP) {
						return true // Disambiguation needed
					}
				} else {
					return true // Empty square, disambiguation needed
				}
			}
		}
	}
	return false
}

// Helper function to calculate captured pieces by comparing current board to starting position
func getCapturedPieces(gameBoard *board.Board) ([]CapturedPiece, []CapturedPiece) {
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

// Helper function to create complete game state with evaluation
func createCompleteGameState(gameBoard *board.Board, message string, evaluation int) GameState {
	capturedWhite, capturedBlack := getCapturedPieces(gameBoard)

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

	return state
}
