package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/engine"
	"github.com/zully/chess-engine/internal/moves"
)

func printHelp() {
	fmt.Println("\nChess Engine CLI Commands:")
	fmt.Println("  Enter moves directly using standard chess notation:")
	fmt.Println("    - Pawn moves: e4, d5")
	fmt.Println("    - Piece moves: Nf3, Bb5")
	fmt.Println("    - Captures: exd5, Bxe4")
	fmt.Println("    - Castling: O-O (kingside), O-O-O (queenside)")
	fmt.Println("\n  Engine commands:")
	fmt.Println("    engine (en) - Let computer play the next move")
	fmt.Println("    auto        - Toggle auto-play mode (computer vs computer)")
	fmt.Println("\n  Other commands:")
	fmt.Println("    display (d) - Show the current board position")
	fmt.Println("    eval (e)    - Show evaluation of current position")
	fmt.Println("    help (h)    - Show this help message")
	fmt.Println("    quit (q)    - Exit the program")
}

// isValidMoveNotation checks if a string looks like a valid chess move
func isValidMoveNotation(move string) bool {
	// Handle castling
	if strings.ToLower(move) == "o-o" || strings.ToLower(move) == "o-o-o" {
		return true
	}

	// Handle piece moves (e.g., Nf3, Bxe4)
	if len(move) >= 3 && (move[0] == 'N' || move[0] == 'B' || move[0] == 'R' || move[0] == 'Q' || move[0] == 'K') {
		return true
	}

	// Handle pawn moves (e.g., e4, exd5, a1=Q, a1Q) - must be exactly 2 chars for simple moves or 4 for captures
	if len(move) == 2 && move[0] >= 'a' && move[0] <= 'h' && move[1] >= '1' && move[1] <= '8' {
		return true
	}

	// Handle pawn captures (e.g., exd5)
	if len(move) == 4 && move[0] >= 'a' && move[0] <= 'h' && move[1] == 'x' &&
		move[2] >= 'a' && move[2] <= 'h' && move[3] >= '1' && move[3] <= '8' {
		return true
	}

	// Handle pawn promotion without = (e.g., a1Q, a8R)
	if len(move) == 3 && move[0] >= 'a' && move[0] <= 'h' && move[1] >= '1' && move[1] <= '8' &&
		(move[2] == 'Q' || move[2] == 'R' || move[2] == 'B' || move[2] == 'N') {
		return true
	}

	// Handle pawn promotion with = (e.g., a1=Q, a8=R)
	if len(move) == 4 && move[0] >= 'a' && move[0] <= 'h' && move[1] >= '1' && move[1] <= '8' &&
		move[2] == '=' && (move[3] == 'Q' || move[3] == 'R' || move[3] == 'B' || move[3] == 'N') {
		return true
	}

	// Handle pawn capture with promotion (e.g., exd8=Q, exd8Q)
	if len(move) >= 5 && move[0] >= 'a' && move[0] <= 'h' && move[1] == 'x' &&
		move[2] >= 'a' && move[2] <= 'h' && move[3] >= '1' && move[3] <= '8' {
		if len(move) == 5 && (move[4] == 'Q' || move[4] == 'R' || move[4] == 'B' || move[4] == 'N') {
			return true // exd8Q format
		}
		if len(move) == 6 && move[4] == '=' && (move[5] == 'Q' || move[5] == 'R' || move[5] == 'B' || move[5] == 'N') {
			return true // exd8=Q format
		}
	}

	return false
}

func processCommand(cmd string, b *board.Board, autoPlay *bool) bool {
	// Clean up the command
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return true
	}

	// Split into fields but preserve original case
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return true
	}

	// First check if it's a valid move notation (preserve case for piece moves)
	if isValidMoveNotation(parts[0]) {
		move := parts[0]
		err := b.MakeMove(move)
		if err != nil {
			fmt.Printf("Invalid move: %v\n", err)
			return true
		}
		fmt.Printf("Played %s\n", move)
		fmt.Println(b)
		return true
	}

	switch parts[0] {
	case "quit", "q":
		return false

	case "help", "h":
		printHelp()

	case "display", "d":
		fmt.Println(b)

	case "eval", "e":
		score := engine.Evaluate(b)
		fmt.Printf("Position evaluation: %d\n", score)

	case "engine", "en":
		fmt.Println("Engine is thinking...")
		result := engine.FindBestMove(b, 6) // Search depth 6 for stronger play
		if result.BestMove == (moves.Move{}) {
			fmt.Println("No legal moves found!")
			return true
		}

		// Execute the engine move directly
		err := engine.ExecuteEngineMove(b, result.BestMove)
		if err != nil {
			fmt.Printf("Engine error: %v\n", err)
			return true
		}

		// Convert to notation for display only
		moveNotation := formatMoveNotation(result.BestMove)
		fmt.Printf("Engine plays: %s (evaluation: %d)\n", moveNotation, result.Score)
		fmt.Println(b)

	case "auto":
		*autoPlay = !*autoPlay
		if *autoPlay {
			fmt.Println("Auto-play mode ON - Computer will play both sides")
		} else {
			fmt.Println("Auto-play mode OFF - Manual input required")
		}

	case "move", "m":
		if len(parts) != 2 {
			fmt.Println("Invalid move format. Examples: e4, Nf3, O-O, exd5")
			return true
		}
		move := parts[1]

		err := b.MakeMove(move)
		if err != nil {
			fmt.Printf("Invalid move: %v\n", err)
			return true
		}

		fmt.Printf("Played %s\n", move)
		fmt.Println(b)

	default:
		fmt.Println("Unknown command. Type 'help' for available commands")
	}

	return true
}

// formatMoveNotation converts a Move struct to standard algebraic notation
func formatMoveNotation(move moves.Move) string {
	notation := ""

	// Add piece prefix (except for pawns)
	if move.Piece != "P" {
		notation += move.Piece
	}

	// Add capture indicator
	if move.Capture {
		if move.Piece == "P" {
			// For pawn captures, include the file
			notation += move.From[0:1] + "x"
		} else {
			notation += "x"
		}
	}

	// Add destination square
	notation += move.To

	// Add promotion notation
	if move.Promote != "" {
		notation += "=" + move.Promote
	}

	return notation
}

func main() {
	fmt.Println("Chess Engine CLI")
	fmt.Println("Type 'help' for available commands")

	b := board.NewBoard()
	reader := bufio.NewReader(os.Stdin)
	autoPlay := false

	// Show initial position
	fmt.Println(b)

	// Main input loop
	for {
		// Auto-play mode: computer plays automatically
		if autoPlay {
			fmt.Println("Computer is thinking...")
			result := engine.FindBestMove(b, 5) // Depth 5 for stronger auto-play
			if result.BestMove == (moves.Move{}) {
				fmt.Println("Game over - no legal moves!")
				break
			}

			err := engine.ExecuteEngineMove(b, result.BestMove)
			if err != nil {
				fmt.Printf("Computer error: %v\n", err)
				autoPlay = false
				continue
			}

			moveNotation := formatMoveNotation(result.BestMove)
			fmt.Printf("Computer plays: %s\n", moveNotation)
			fmt.Println(b)

			// Brief pause for readability
			fmt.Println("Press Enter to continue or type 'auto' to stop...")
		}

		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			break
		}

		// Process the command, preserving case for moves
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}

		// Convert to lowercase only if it's not a potential move
		cmd := input
		if !isValidMoveNotation(strings.Fields(input)[0]) {
			cmd = strings.ToLower(input)
		}

		// Process the command
		if !processCommand(cmd, b, &autoPlay) {
			break
		}
	}

	fmt.Println("Goodbye!")
}
