package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/engine"
	"github.com/zully/chess-engine/internal/moves"
)

type GameState struct {
	Board         *board.Board `json:"board"`
	Message       string       `json:"message"`
	Error         string       `json:"error,omitempty"`
	GameOver      bool         `json:"gameOver"`
	InCheck       bool         `json:"inCheck"`
	IsCheckmate   bool         `json:"isCheckmate"`
	Draw          bool         `json:"draw"`
	DrawReason    string       `json:"drawReason"`
	ThreefoldRep  bool         `json:"threefoldRepetition"`
	PositionCount int          `json:"positionCount"`
}

type MoveRequest struct {
	Move string `json:"move"`
}

type EngineRequest struct {
	Depth int `json:"depth,omitempty"`
}

var gameBoard *board.Board

func main() {
	// Initialize the game board
	gameBoard = board.NewBoard()

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

	fmt.Println("Chess Web GUI starting on http://localhost:8080")
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
                <h2>Chess Engine GUI</h2>
                <div class="engine-section">
                    <label class="checkbox-label">
                        <input type="checkbox" id="engine-white-checkbox"> Engine plays White
                    </label>
                    <label class="checkbox-label">
                        <input type="checkbox" id="engine-black-checkbox"> Engine plays Black
                    </label>
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

	if isDraw {
		state.Message = fmt.Sprintf("Draw! %s", drawReason)
	} else if state.InCheck && !state.IsCheckmate {
		if gameBoard.WhiteToMove {
			state.Message = "White is in check!"
		} else {
			state.Message = "Black is in check!"
		}
	} else if state.IsCheckmate {
		if gameBoard.WhiteToMove {
			state.Message = "Checkmate! Black wins!"
		} else {
			state.Message = "Checkmate! White wins!"
		}
	} else {
		if gameBoard.WhiteToMove {
			state.Message = "White to move"
		} else {
			state.Message = "Black to move"
		}
	}

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
	state := GameState{Board: gameBoard}

	if err != nil {
		state.Error = err.Error()
	} else {
		// Update check/checkmate status first
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

		// Set message with check/checkmate/draw announcement
		if isDraw {
			state.Message = fmt.Sprintf("Played %s - Draw! %s", req.Move, drawReason)
		} else if state.IsCheckmate {
			if gameBoard.WhiteToMove {
				state.Message = fmt.Sprintf("Played %s - Checkmate! Black wins!", req.Move)
			} else {
				state.Message = fmt.Sprintf("Played %s - Checkmate! White wins!", req.Move)
			}
		} else if state.InCheck {
			if gameBoard.WhiteToMove {
				state.Message = fmt.Sprintf("Played %s - White is in check!", req.Move)
			} else {
				state.Message = fmt.Sprintf("Played %s - Black is in check!", req.Move)
			}
		} else {
			state.Message = fmt.Sprintf("Played %s", req.Move)
		}
	}

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

	depth := 6
	if req.Depth > 0 && req.Depth <= 10 {
		depth = req.Depth
	}

	result := engine.FindBestMove(gameBoard, depth)
	state := GameState{Board: gameBoard}

	if result.BestMove.From == "" {
		state.Error = "No valid moves found"
	} else {
		err := engine.ExecuteEngineMove(gameBoard, result.BestMove)
		if err != nil {
			state.Error = err.Error()
		} else {
			// Update check/checkmate status first
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

			// Set message with check/checkmate/draw announcement
			baseMessage := fmt.Sprintf("Engine played %s (evaluation: %d)",
				formatEngineMove(result.BestMove), result.Score)

			if isDraw {
				state.Message = baseMessage + " - Draw! " + drawReason
			} else if state.IsCheckmate {
				if gameBoard.WhiteToMove {
					state.Message = baseMessage + " - Checkmate! Black wins!"
				} else {
					state.Message = baseMessage + " - Checkmate! White wins!"
				}
			} else if state.InCheck {
				if gameBoard.WhiteToMove {
					state.Message = baseMessage + " - White is in check!"
				} else {
					state.Message = baseMessage + " - Black is in check!"
				}
			} else {
				state.Message = baseMessage
			}
		}
	}

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

func formatEngineMove(move moves.Move) string {
	notation := ""
	if move.Piece != "P" {
		notation += move.Piece
	}
	if move.Capture {
		if move.Piece == "P" && move.From != "" {
			notation += string(move.From[0])
		}
		notation += "x"
	}
	notation += move.To
	if move.Promote != "" {
		notation += "=" + move.Promote
	}
	if move.Castle != "" {
		notation = move.Castle
	}
	return notation
}
