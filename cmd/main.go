package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/zully/chess-engine/internal/board"
	"github.com/zully/chess-engine/internal/engine"
)

func printHelp() {
	fmt.Println("\nChess Engine CLI Commands:")
	fmt.Println("  Enter moves directly using standard chess notation:")
	fmt.Println("    - Pawn moves: e4, d5")
	fmt.Println("    - Piece moves: Nf3, Bb5")
	fmt.Println("    - Captures: exd5, Bxe4")
	fmt.Println("    - Castling: O-O (kingside), O-O-O (queenside)")
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

	// Handle pawn moves (e.g., e4, exd5)
	if len(move) >= 2 && move[0] >= 'a' && move[0] <= 'h' {
		return true
	}

	return false
}

func processCommand(cmd string, b *board.Board) bool {
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

func main() {
	fmt.Println("Chess Engine CLI")
	fmt.Println("Type 'help' for available commands")

	b := board.NewBoard()
	reader := bufio.NewReader(os.Stdin)

	// Show initial position
	fmt.Println(b)

	// Main input loop
	for {
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
		if !processCommand(cmd, b) {
			break
		}
	}

	fmt.Println("Goodbye!")
}
