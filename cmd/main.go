package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/uci"
	"github.com/zully/chess-engine/internal/web"
)

func main() {
	// Initialize the game board
	gameBoard := board.NewBoard()

	// Initialize Stockfish engine (Docker environment)
	stockfishPath := "/usr/local/bin/stockfish"

	var err error
	stockfishEngine, err := uci.NewEngine(stockfishPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize Stockfish engine: %v", err)
		log.Println("Engine features will be disabled")
	}

	// Create web server with dependencies
	server := web.NewServer(gameBoard, stockfishEngine)

	// Serve static files (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API endpoints - use server methods
	http.HandleFunc("/api/state", server.GetGameState)
	http.HandleFunc("/api/move", server.MakeMove)
	http.HandleFunc("/api/engine", server.EngineMove)
	http.HandleFunc("/api/analysis", server.GetEngineAnalysis)
	http.HandleFunc("/api/undo", server.UndoMove)
	http.HandleFunc("/api/reset", server.ResetGame)

	// Main page
	http.HandleFunc("/", server.HomePage)

	fmt.Println("Chess Web GUI with Stockfish starting on http://localhost:8080")
	if stockfishEngine != nil {
		fmt.Println("Stockfish engine initialized successfully")
	} else {
		fmt.Println("Running without engine (moves disabled)")
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
