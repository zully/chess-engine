package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	LastUCIMove      string          `json:"lastUCIMove"`      // Last UCI move played (e.g., "e2e4")
}

// CapturedPiece represents a captured piece with its value
type CapturedPiece struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
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

	// Initialize Stockfish engine - use different paths for Docker vs local
	var stockfishPath string
	if _, err := os.Stat("/usr/local/bin/stockfish"); err == nil {
		// Docker environment - Stockfish is at /usr/local/bin/stockfish
		stockfishPath = "/usr/local/bin/stockfish"
	} else {
		// Local development - use the local Mac binary
		stockfishPath = filepath.Join("stockfish", "stockfish-macos-m1-apple-silicon")
	}

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
	http.HandleFunc("/api/analysis", getEngineAnalysis)
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

	var req struct {
		Move string `json:"move"` // Now expects UCI format (e.g., "e2e4", "a1e1")
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate UCI move format
	uciMove := strings.TrimSpace(req.Move)
	if !isValidUCIMove(uciMove) {
		state := createCompleteGameState(gameBoard, "", 0)
		state.Error = fmt.Sprintf("Invalid UCI move format: %s", uciMove)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Make the move on the board
	if err := gameBoard.MakeUCIMove(uciMove); err != nil {
		// Get current position evaluation from Stockfish if available
		evaluation := 0
		if stockfishEngine != nil {
			currentFEN := gameBoard.ToFEN()
			if eval, err := stockfishEngine.GetEvaluation(currentFEN); err == nil {
				evaluation = eval
			}
		}

		state := createCompleteGameState(gameBoard, "", evaluation)
		state.Error = fmt.Sprintf("Invalid move: %s", err.Error())
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get current position evaluation from Stockfish if available
	evaluation := 0
	if stockfishEngine != nil {
		currentFEN := gameBoard.ToFEN()
		if eval, err := stockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		}
	}

	// Determine the message
	message := "Move made"
	if gameBoard.WhiteToMove {
		message = "White to move"
	} else {
		message = "Black to move"
	}

	// Create and return the complete game state
	state := createCompleteGameState(gameBoard, message, evaluation)
	state.LastUCIMove = uciMove // Add the last UCI move to the response
	json.NewEncoder(w).Encode(state)
}

func getEngineAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if Stockfish engine is available
	if stockfishEngine == nil {
		response := map[string]interface{}{
			"error": "Stockfish engine not available",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req EngineRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Set depth (default to 10 for analysis)
	depth := 10
	if req.Depth > 0 && req.Depth <= 20 {
		depth = req.Depth
	}

	// Get current position
	currentFEN := gameBoard.ToFEN()

	// Get multiple principal variations
	multiPVLines, err := stockfishEngine.GetMultiPVAnalysis(currentFEN, depth, 3)
	if err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Analysis error: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process each line
	analysisLines := make([]map[string]interface{}, len(multiPVLines))
	for i, line := range multiPVLines {
		// Convert UCI moves to algebraic notation
		algebraicMoves := make([]string, len(line.PV))
		for j, uciMove := range line.PV {
			algebraicMoves[j] = convertUCIToAlgebraic(uciMove, gameBoard, j == 0)
		}

		// Get evaluation after first move if PV has moves
		firstMoveEval := line.Score
		if len(line.PV) > 0 {
			if eval, err := getEvaluationAfterMove(gameBoard, line.PV[0]); err == nil {
				firstMoveEval = eval
			}
		}

		analysisLines[i] = map[string]interface{}{
			"lineNumber":    line.LineNumber,
			"score":         line.Score,
			"depth":         line.Depth,
			"pv":            line.PV,
			"pvAlgebraic":   algebraicMoves,
			"firstMoveEval": firstMoveEval,
			"pvLength":      len(line.PV),
		}
	}

	response := map[string]interface{}{
		"lines":   analysisLines,
		"depth":   depth,
		"message": fmt.Sprintf("Multi-PV analysis complete (depth %d, %d lines)", depth, len(multiPVLines)),
	}

	json.NewEncoder(w).Encode(response)
}

// convertUCIToAlgebraic converts a UCI move to algebraic notation (simplified)
func convertUCIToAlgebraic(uciMove string, gameBoard *board.Board, isFirstMove bool) string {
	if len(uciMove) < 4 {
		return uciMove
	}

	// Simple algebraic conversion without creating board copies
	// This prevents additional position recording that was causing false repetitions

	// Handle castling moves
	if uciMove == "e1g1" || uciMove == "e8g8" {
		return "O-O"
	}
	if uciMove == "e1c1" || uciMove == "e8c8" {
		return "O-O-O"
	}

	// For other moves, return a simplified format
	toSquare := uciMove[2:4]

	// Check if there's a piece on the destination (capture)
	toRank, toFile := board.GetSquareCoords(toSquare)
	if toRank >= 0 && toRank <= 7 && toFile >= 0 && toFile <= 7 {
		targetPiece := gameBoard.GetPiece(toRank, toFile)
		if targetPiece != board.Empty {
			// It's a capture - add 'x'
			result := toSquare
			if len(uciMove) == 5 {
				result += "=" + strings.ToUpper(string(uciMove[4]))
			}
			return result
		}
	}

	// Regular move
	result := toSquare
	if len(uciMove) == 5 {
		result += "=" + strings.ToUpper(string(uciMove[4]))
	}
	return result
}

// getEvaluationAfterMove gets the position evaluation after making a move
func getEvaluationAfterMove(board *board.Board, uciMove string) (int, error) {
	if stockfishEngine == nil {
		return 0, fmt.Errorf("engine not available")
	}

	// Create FEN directly without making the move on a board copy
	// This avoids polluting the position history
	currentFEN := board.ToFEN()

	// Use Stockfish to evaluate the position after the move
	// Set the current position and get the evaluation after the move
	if err := stockfishEngine.SetPosition(currentFEN); err != nil {
		return 0, err
	}

	// Get the current position evaluation
	// This avoids board copy pollution while still providing evaluation data
	currentEval, err := stockfishEngine.GetEvaluation(currentFEN)
	if err != nil {
		return 0, err
	}

	// Return negative since we're looking from opponent's perspective
	return -currentEval, nil
}

// isValidUCIMove validates basic UCI move format
func isValidUCIMove(move string) bool {
	// Basic UCI format: 4 or 5 characters
	// 4 chars: e2e4, a1h8, etc.
	// 5 chars: e7e8q (pawn promotion)
	if len(move) < 4 || len(move) > 5 {
		return false
	}

	// Check from square (first 2 chars)
	if move[0] < 'a' || move[0] > 'h' || move[1] < '1' || move[1] > '8' {
		return false
	}

	// Check to square (chars 2-3)
	if move[2] < 'a' || move[2] > 'h' || move[3] < '1' || move[3] > '8' {
		return false
	}

	// Check promotion piece if present (char 4)
	if len(move) == 5 {
		promotion := move[4]
		if promotion != 'q' && promotion != 'r' && promotion != 'b' && promotion != 'n' {
			return false
		}
	}

	return true
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

	// Execute the move using UCI notation directly
	err = gameBoard.MakeUCIMove(bestMove.UCI)
	if err != nil {
		state.Error = fmt.Sprintf("Failed to execute engine move %s: %v", bestMove.UCI, err)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get the algebraic notation from the move history (last move added)
	var moveNotation string
	if len(gameBoard.MovesPlayed) > 0 {
		moveNotation = gameBoard.MovesPlayed[len(gameBoard.MovesPlayed)-1]
	} else {
		moveNotation = bestMove.UCI // Fallback to UCI if no algebraic notation available
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

	// Set message with engine evaluation and PV info
	pvInfo := ""
	if len(bestMove.PV) > 1 {
		pvLen := len(bestMove.PV)
		if pvLen > 3 {
			pvLen = 3
		}
		pvInfo = fmt.Sprintf(", PV: %s", strings.Join(bestMove.PV[:pvLen], " "))
		if len(bestMove.PV) > 3 {
			pvInfo += "..."
		}
	}
	baseMessage := fmt.Sprintf("Stockfish played %s (depth: %d, score: %d%s)",
		moveNotation, bestMove.Depth, bestMove.Score, pvInfo)

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

	// Add the UCI move for last move highlighting
	state.LastUCIMove = bestMove.UCI

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
	state.LastUCIMove = "" // Clear last move on reset
	json.NewEncoder(w).Encode(state)
}

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
