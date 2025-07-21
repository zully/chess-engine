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
	Board       *board.Board `json:"board"`
	Message     string       `json:"message"`
	Error       string       `json:"error,omitempty"`
	GameOver    bool         `json:"gameOver"`
	InCheck     bool         `json:"inCheck"`
	IsCheckmate bool         `json:"isCheckmate"`
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
                <div class="input-section">
                    <input type="text" id="move-input" placeholder="Enter move (e.g., e4, Nf3, O-O)" autofocus>
                    <button id="make-move-btn">Make Move</button>
                </div>
                <div class="button-section">
                    <button id="engine-btn">Engine Move</button>
                    <button id="auto-btn">Auto Play</button>
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

	state := GameState{
		Board:       gameBoard,
		InCheck:     gameBoard.IsInCheck(gameBoard.WhiteToMove),
		IsCheckmate: gameBoard.IsCheckmate(gameBoard.WhiteToMove),
		GameOver:    gameBoard.IsCheckmate(gameBoard.WhiteToMove),
	}

	if state.InCheck && !state.IsCheckmate {
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
		state.Message = fmt.Sprintf("Played %s", req.Move)
	}

	// Update check/checkmate status
	state.InCheck = gameBoard.IsInCheck(gameBoard.WhiteToMove)
	state.IsCheckmate = gameBoard.IsCheckmate(gameBoard.WhiteToMove)
	state.GameOver = state.IsCheckmate

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

	depth := 4
	if req.Depth > 0 && req.Depth <= 8 {
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
			state.Message = fmt.Sprintf("Engine played %s (evaluation: %d)",
				formatEngineMove(result.BestMove), result.Score)
		}
	}

	// Update check/checkmate status
	state.InCheck = gameBoard.IsInCheck(gameBoard.WhiteToMove)
	state.IsCheckmate = gameBoard.IsCheckmate(gameBoard.WhiteToMove)
	state.GameOver = state.IsCheckmate

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
