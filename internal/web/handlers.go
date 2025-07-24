package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/game"
	"github.com/zully/chess-engine/internal/uci"
)

// Server holds the dependencies for web handlers
type Server struct {
	GameBoard       *board.Board
	StockfishEngine *uci.Engine
}

// NewServer creates a new web server instance
func NewServer(gameBoard *board.Board, stockfishEngine *uci.Engine) *Server {
	return &Server{
		GameBoard:       gameBoard,
		StockfishEngine: stockfishEngine,
	}
}

func (s *Server) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	// Serve the HTML template file
	http.ServeFile(w, r, "web/templates/index.html")
}

func (s *Server) GetGameState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current position evaluation from Stockfish if available
	evaluation := 0
	if s.StockfishEngine != nil {
		currentFEN := s.GameBoard.ToFEN()
		if eval, err := s.StockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		}
	}

	// Create complete game state
	message := "Ready to play"
	if s.GameBoard.WhiteToMove {
		message = "White to move"
	} else {
		message = "Black to move"
	}

	state := game.CreateCompleteGameState(s.GameBoard, message, evaluation, s.StockfishEngine)
	json.NewEncoder(w).Encode(state)
}

func (s *Server) MakeMove(w http.ResponseWriter, r *http.Request) {
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
	if !IsValidUCIMove(uciMove) {
		state := game.CreateCompleteGameState(s.GameBoard, "", 0, s.StockfishEngine)
		state.Error = fmt.Sprintf("Invalid UCI move format: %s", uciMove)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Make the move on the board
	if err := s.GameBoard.MakeUCIMove(uciMove); err != nil {
		// Get current position evaluation from Stockfish if available
		evaluation := 0
		if s.StockfishEngine != nil {
			currentFEN := s.GameBoard.ToFEN()
			if eval, err := s.StockfishEngine.GetEvaluation(currentFEN); err == nil {
				evaluation = eval
			}
		}

		state := game.CreateCompleteGameState(s.GameBoard, "", evaluation, s.StockfishEngine)
		state.Error = fmt.Sprintf("Invalid move: %s", err.Error())
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get current position evaluation from Stockfish if available
	evaluation := 0
	if s.StockfishEngine != nil {
		currentFEN := s.GameBoard.ToFEN()
		if eval, err := s.StockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		}
	}

	// Determine the message
	message := "Move made"
	if s.GameBoard.WhiteToMove {
		message = "White to move"
	} else {
		message = "Black to move"
	}

	// Create and return the complete game state
	state := game.CreateCompleteGameState(s.GameBoard, message, evaluation, s.StockfishEngine)
	state.LastUCIMove = uciMove // Add the last UCI move to the response
	json.NewEncoder(w).Encode(state)
}

func (s *Server) GetEngineAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if Stockfish engine is available
	if s.StockfishEngine == nil {
		response := map[string]interface{}{
			"error": "Stockfish engine not available",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req game.EngineRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Set depth (default to 10 for analysis)
	depth := 10
	if req.Depth > 0 && req.Depth <= 20 {
		depth = req.Depth
	}

	// Get current position
	currentFEN := s.GameBoard.ToFEN()

	// Get multiple principal variations
	multiPVLines, err := s.StockfishEngine.GetMultiPVAnalysis(currentFEN, depth, 3)
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
			algebraicMoves[j] = ConvertUCIToAlgebraic(uciMove, s.GameBoard, j == 0)
		}

		// Get evaluation after first move if PV has moves
		firstMoveEval := line.Score
		if len(line.PV) > 0 {
			if eval, err := GetEvaluationAfterMove(s.GameBoard, line.PV[0], s.StockfishEngine); err == nil {
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

func (s *Server) EngineMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req game.EngineRequest
	json.NewDecoder(r.Body).Decode(&req)

	state := game.GameState{Board: s.GameBoard}

	// Check if Stockfish engine is available
	if s.StockfishEngine == nil {
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
			err := s.StockfishEngine.SetEloRating(req.Elo)
			if err != nil {
				// ELO setting failed, engine will use default strength
			}
		} else {
			// Invalid ELO rating, use default strength
			err := s.StockfishEngine.DisableStrengthLimit()
			if err != nil {
				// Failed to disable strength limit, engine will use current settings
			}
		}
	} else {
		// Full strength (disable ELO limiting)
		err := s.StockfishEngine.DisableStrengthLimit()
		if err != nil {
			// Failed to disable strength limit, engine will use current settings
		}
	}

	// Set current position in Stockfish using FEN
	fen := s.GameBoard.ToFEN()
	err := s.StockfishEngine.SetPosition(fen)
	if err != nil {
		state.Error = fmt.Sprintf("Failed to set position: %v", err)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get best move from Stockfish
	currentFEN := s.GameBoard.ToFEN()

	bestMove, err := s.StockfishEngine.GetBestMove(currentFEN, depth)
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
	err = s.GameBoard.MakeUCIMove(bestMove.UCI)
	if err != nil {
		state.Error = fmt.Sprintf("Failed to execute engine move %s: %v", bestMove.UCI, err)
		json.NewEncoder(w).Encode(state)
		return
	}

	// Get the algebraic notation from the move history (last move added)
	var moveNotation string
	if len(s.GameBoard.MovesPlayed) > 0 {
		moveNotation = s.GameBoard.MovesPlayed[len(s.GameBoard.MovesPlayed)-1]
	} else {
		moveNotation = bestMove.UCI // Fallback to UCI if no algebraic notation available
	}

	// Update game state
	state.InCheck = s.GameBoard.IsInCheck(s.GameBoard.WhiteToMove)
	state.IsCheckmate = s.GameBoard.IsCheckmate(s.GameBoard.WhiteToMove)
	state.GameOver = state.IsCheckmate

	// Check for draws
	isDraw := s.GameBoard.IsDraw()
	drawReason := ""
	if isDraw {
		if s.GameBoard.IsThreefoldRepetition() {
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
		if s.GameBoard.WhiteToMove {
			baseMessage += " - Black wins!"
		} else {
			baseMessage += " - White wins!"
		}
	} else if state.InCheck {
		if s.GameBoard.WhiteToMove {
			baseMessage += " - White in check!"
		} else {
			baseMessage += " - Black in check!"
		}
	}

	// Create complete game state with evaluation
	state = game.CreateCompleteGameState(s.GameBoard, baseMessage, bestMove.Evaluation, s.StockfishEngine)

	// Add the UCI move for last move highlighting
	state.LastUCIMove = bestMove.UCI

	json.NewEncoder(w).Encode(state)
}

func (s *Server) UndoMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if there are moves to undo
	if len(s.GameBoard.MovesPlayed) == 0 {
		state := game.GameState{
			Board: s.GameBoard,
			Error: "No moves to undo!",
		}
		json.NewEncoder(w).Encode(state)
		return
	}

	// Store the current moves list
	currentMoves := make([]string, len(s.GameBoard.MovesPlayed))
	copy(currentMoves, s.GameBoard.MovesPlayed)

	// Remove the last move
	movesToReplay := currentMoves[:len(currentMoves)-1]

	// Create a fresh board
	s.GameBoard = board.NewBoard()

	// Replay all moves except the last one
	for _, move := range movesToReplay {
		err := s.GameBoard.MakeMove(move)
		if err != nil {
			// If replay fails, restore the original board state
			// This shouldn't happen, but just in case
			s.GameBoard = board.NewBoard()
			for _, originalMove := range currentMoves {
				s.GameBoard.MakeMove(originalMove)
			}
			state := game.GameState{
				Board: s.GameBoard,
				Error: fmt.Sprintf("Failed to undo move: %v", err),
			}
			json.NewEncoder(w).Encode(state)
			return
		}
	}

	// Create and return the updated game state
	inCheck := s.GameBoard.IsInCheck(s.GameBoard.WhiteToMove)
	isCheckmate := false
	if inCheck {
		isCheckmate = s.GameBoard.IsCheckmate(s.GameBoard.WhiteToMove)
	}

	isDraw := s.GameBoard.IsDraw()
	drawReason := ""
	if isDraw {
		if s.GameBoard.IsThreefoldRepetition() {
			drawReason = "Threefold repetition"
		} else {
			drawReason = "Stalemate"
		}
	}

	state := game.GameState{
		Board:         s.GameBoard,
		InCheck:       inCheck,
		IsCheckmate:   isCheckmate,
		GameOver:      isCheckmate || isDraw,
		Draw:          isDraw,
		DrawReason:    drawReason,
		ThreefoldRep:  s.GameBoard.IsThreefoldRepetition(),
		PositionCount: s.GameBoard.GetPositionCount(),
	}

	lastMove := currentMoves[len(currentMoves)-1]
	state.Message = fmt.Sprintf("Undid move %s", lastMove)

	json.NewEncoder(w).Encode(state)
}

func (s *Server) ResetGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a new board
	s.GameBoard = board.NewBoard()

	// Get initial evaluation
	evaluation := 0
	if s.StockfishEngine != nil {
		currentFEN := s.GameBoard.ToFEN()
		if eval, err := s.StockfishEngine.GetEvaluation(currentFEN); err == nil {
			evaluation = eval
		}
	}

	// Create complete game state with evaluation
	state := game.CreateCompleteGameState(s.GameBoard, "Game reset. White to move.", evaluation, s.StockfishEngine)
	state.LastUCIMove = "" // Clear last move on reset
	json.NewEncoder(w).Encode(state)
}
