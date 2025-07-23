package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/uci"
)

type GameState struct {
	Board         *board.Board    `json:"board"`
	Message       string          `json:"message"`
	Error         string          `json:"error,omitempty"`
	GameOver      bool            `json:"gameOver"`
	InCheck       bool            `json:"inCheck"`
	IsCheckmate   bool            `json:"isCheckmate"`
	Draw          bool            `json:"draw"`
	DrawReason    string          `json:"drawReason"`
	ThreefoldRep  bool            `json:"threefoldRepetition"`
	PositionCount int             `json:"positionCount"`
	Evaluation    int             `json:"evaluation"`    // Position evaluation in centipawns
	CapturedWhite []CapturedPiece `json:"capturedWhite"` // Pieces captured by White
	CapturedBlack []CapturedPiece `json:"capturedBlack"` // Pieces captured by Black
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
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chess Engine GUI</title>
    <link rel="stylesheet" href="/static/chess.css">
</head>
<body>
    <div class="container">
        <div class="game-area">
            <div class="board-container">
                <div class="board-wrapper">
                    <div id="rank-labels-left" class="rank-labels"></div>
                    <div class="board-with-files">
                        <div id="chess-board"></div>
                        <div id="file-labels-bottom" class="file-labels"></div>
                    </div>
                </div>
            </div>
            <div class="controls">
                <h2>Chess Engine GUI (Stockfish)</h2>
                
                <!-- Evaluation Bar -->
                <div class="evaluation-section">
                    <h3>Position Evaluation</h3>
                    <div class="evaluation-bar-container">
                        <div class="evaluation-bar">
                            <div id="evaluation-fill" class="evaluation-fill"></div>
                            <div class="evaluation-center-line"></div>
                        </div>
                        <div id="evaluation-text" class="evaluation-text">0.00</div>
                    </div>
                </div>

                <!-- Captured Pieces -->
                <div class="captured-section">
                    <h3>Captured Pieces</h3>
                    <div class="captured-container">
                        <div class="captured-side">
                            <h4>White Captured</h4>
                            <div id="captured-white" class="captured-pieces"></div>
                            <div id="captured-white-value" class="captured-value">0</div>
                        </div>
                        <div class="captured-side">
                            <h4>Black Captured</h4>
                            <div id="captured-black" class="captured-pieces"></div>
                            <div id="captured-black-value" class="captured-value">0</div>
                        </div>
                    </div>
                </div>

                <div class="engine-section">
                    <label class="checkbox-label">
                        <input type="checkbox" id="engine-white-checkbox"> Engine plays White
                    </label>
                    <label class="checkbox-label">
                        <input type="checkbox" id="engine-black-checkbox"> Engine plays Black
                    </label>
                </div>
                <div class="strength-section">
                    <h3>Engine Strength</h3>
                    <label for="elo-select">ELO Rating:</label>
                    <select id="elo-select">
                        <option value="0">Maximum Strength (2850+)</option>
                        <option value="2400">Grandmaster (2400)</option>
                        <option value="2200">Master (2200)</option>
                        <option value="2000">Expert (2000)</option>
                        <option value="1800">Class A (1800)</option>
                        <option value="1600">Class B (1600)</option>
                        <option value="1400" selected>Class C (1400)</option>
                        <option value="1200">Class D (1200)</option>
                        <option value="1000">Beginner (1000)</option>
                    </select>
                </div>
                <div class="button-section">
                    <button id="engine-btn">Engine Move</button>
                    <button id="undo-btn">Undo Move</button>
                    <button id="flip-btn">Flip Board</button>
                    <button id="reset-btn">Reset Game</button>
                </div>
                <div id="game-message" class="message"></div>
                <div class="moves-section">
                    <h3>Move History</h3>
                    <div id="move-history"></div>
                </div>
            </div>
        </div>
    </div>
    <script src="/static/chess.js"></script>
</body>
</html>`

	t := template.Must(template.New("home").Parse(tmpl))
	t.Execute(w, nil)
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
			fmt.Printf("Warning: Invalid ELO rating %d (must be 1350-2850)\n", req.Elo)
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
	fromRank, fromFile := getSquareCoords(bestMove.From)
	if fromRank < 0 || fromFile < 0 {
		state.Error = fmt.Sprintf("Invalid UCI move from square: %s", bestMove.From)
		json.NewEncoder(w).Encode(state)
		return
	}

	toRank, toFile := getSquareCoords(bestMove.To)
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

	gameBoard = board.NewBoard()
	state := GameState{
		Board:   gameBoard,
		Message: "Game reset. White to move.",
	}

	json.NewEncoder(w).Encode(state)
}

// getSquareCoords converts algebraic notation to array coordinates
func getSquareCoords(square string) (rank int, file int) {
	if len(square) != 2 {
		return -1, -1
	}

	file = int(square[0] - 'a')
	if file < 0 || file > 7 {
		return -1, -1
	}

	rank = 7 - (int(square[1] - '1')) // Convert '1' to array index 7 from bottom
	if rank < 0 || rank > 7 {
		return -1, -1
	}

	return rank, file
}

// needsDisambiguation checks if there are other pieces of the same type that could move to the same destination
func needsDisambiguation(b *board.Board, pieceType int, fromRank, fromFile, toRank, toFile int) bool {
	// Check all squares on the board for pieces of the same type and color
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			// Skip the current piece
			if rank == fromRank && file == fromFile {
				continue
			}

			// Check if this square has the same piece type
			if b.GetPiece(rank, file) != pieceType {
				continue
			}

			// Check if this piece could also move to the same destination
			canMove := false
			switch pieceType {
			case board.WN, board.BN:
				canMove = canKnightMove(rank, file, toRank, toFile)
			case board.WB, board.BB:
				canMove = canBishopMove(b, rank, file, toRank, toFile)
			case board.WR, board.BR:
				canMove = canRookMove(b, rank, file, toRank, toFile)
			case board.WQ, board.BQ:
				canMove = canQueenMove(b, rank, file, toRank, toFile)
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
					return true // Disambiguation needed
				}
			}
		}
	}
	return false
}

// Piece movement validation functions (simplified versions)
func canKnightMove(fromRank, fromFile, toRank, toFile int) bool {
	rankDiff := abs(toRank - fromRank)
	fileDiff := abs(toFile - fromFile)
	return (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2)
}

func canBishopMove(b *board.Board, fromRank, fromFile, toRank, toFile int) bool {
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
		if b.GetPiece(rank, file) != board.Empty {
			return false
		}
		rank += rankStep
		file += fileStep
	}
	return true
}

func canRookMove(b *board.Board, fromRank, fromFile, toRank, toFile int) bool {
	if fromRank != toRank && fromFile != toFile {
		return false
	}

	// Check path for obstacles
	if fromRank == toRank {
		// Horizontal move
		step := sign(toFile - fromFile)
		for file := fromFile + step; file != toFile; file += step {
			if b.GetPiece(fromRank, file) != board.Empty {
				return false
			}
		}
	} else {
		// Vertical move
		step := sign(toRank - fromRank)
		for rank := fromRank + step; rank != toRank; rank += step {
			if b.GetPiece(rank, fromFile) != board.Empty {
				return false
			}
		}
	}
	return true
}

func canQueenMove(b *board.Board, fromRank, fromFile, toRank, toFile int) bool {
	return canBishopMove(b, fromRank, fromFile, toRank, toFile) ||
		canRookMove(b, fromRank, fromFile, toRank, toFile)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

// Helper function to get piece values for captured pieces
func getPieceValue(piece int) int {
	switch piece {
	case board.WP, board.BP:
		return 1
	case board.WN, board.BN, board.WB, board.BB:
		return 3
	case board.WR, board.BR:
		return 5
	case board.WQ, board.BQ:
		return 9
	default:
		return 0
	}
}

// Helper function to get piece type string
func getPieceTypeString(piece int) string {
	switch piece {
	case board.WP, board.BP:
		return "P"
	case board.WN, board.BN:
		return "N"
	case board.WB, board.BB:
		return "B"
	case board.WR, board.BR:
		return "R"
	case board.WQ, board.BQ:
		return "Q"
	case board.WK, board.BK:
		return "K"
	default:
		return ""
	}
}

// Helper function to calculate captured pieces (simplified version for now)
func getCapturedPieces(gameBoard *board.Board) ([]CapturedPiece, []CapturedPiece) {
	// For now, return empty slices - this would need move history to properly calculate
	// TODO: Track captured pieces during gameplay
	return []CapturedPiece{}, []CapturedPiece{}
}

// Helper function to create complete game state with evaluation
func createCompleteGameState(gameBoard *board.Board, message string, evaluation int) GameState {
	capturedWhite, capturedBlack := getCapturedPieces(gameBoard)

	state := GameState{
		Board:         gameBoard,
		Message:       message,
		Evaluation:    evaluation,
		CapturedWhite: capturedWhite,
		CapturedBlack: capturedBlack,
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
