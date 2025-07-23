package uci

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Engine represents a UCI chess engine (Stockfish)
type Engine struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	ready  bool
}

// EngineMove represents a move from the engine
type EngineMove struct {
	From       string
	To         string
	Score      int
	Depth      int
	UCI        string // Store the original UCI format
	Evaluation int    // Position evaluation in centipawns (positive = better for white)
}

// NewEngine creates a new UCI engine instance
func NewEngine(enginePath string) (*Engine, error) {
	cmd := exec.Command(enginePath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start engine: %v", err)
	}

	engine := &Engine{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdin),
		stdout: bufio.NewScanner(stdout),
		ready:  false,
	}

	// Initialize the engine
	if err := engine.initialize(); err != nil {
		engine.Close()
		return nil, fmt.Errorf("failed to initialize engine: %v", err)
	}

	return engine, nil
}

// initialize sends UCI initialization commands
func (e *Engine) initialize() error {
	// Send UCI command
	if err := e.sendCommand("uci"); err != nil {
		return err
	}

	// Wait for uciok response
	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())
		if line == "uciok" {
			break
		}
	}

	// Send isready and wait for readyok
	if err := e.sendCommand("isready"); err != nil {
		return err
	}

	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())
		if line == "readyok" {
			e.ready = true
			break
		}
	}

	return nil
}

// sendCommand sends a command to the engine
func (e *Engine) sendCommand(command string) error {
	if _, err := e.stdin.WriteString(command + "\n"); err != nil {
		return err
	}
	return e.stdin.Flush()
}

// SetPosition sets the current position using FEN notation
func (e *Engine) SetPosition(fen string) error {
	if !e.ready {
		return fmt.Errorf("engine not ready")
	}

	command := fmt.Sprintf("position fen %s", fen)
	return e.sendCommand(command)
}

// SetPositionWithMoves sets position from start with move history
func (e *Engine) SetPositionWithMoves(moves []string) error {
	if !e.ready {
		return fmt.Errorf("engine not ready")
	}

	command := "position startpos"
	if len(moves) > 0 {
		command += " moves " + strings.Join(moves, " ")
	}
	return e.sendCommand(command)
}

// GetBestMove asks the engine for the best move with optional depth
func (e *Engine) GetBestMove(fen string, depth int) (*EngineMove, error) {
	if !e.ready {
		return nil, fmt.Errorf("engine not ready")
	}

	// Set the position
	if err := e.sendCommand(fmt.Sprintf("position fen %s", fen)); err != nil {
		return nil, err
	}

	// Start the search
	command := "go"
	if depth > 0 {
		command += fmt.Sprintf(" depth %d", depth)
	}
	if err := e.sendCommand(command); err != nil {
		return nil, err
	}

	var bestMove *EngineMove
	var lastScore int

	// Read the search output
	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())

		// Parse info lines for score information
		if strings.HasPrefix(line, "info") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "score" && i+2 < len(parts) {
					if parts[i+1] == "cp" { // centipawn score
						if score, err := strconv.Atoi(parts[i+2]); err == nil {
							lastScore = score
						}
					}
				}
			}
		}

		// Parse the bestmove line
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				uciMove := parts[1]
				bestMove = &EngineMove{
					From: uciMove[:2],
					To:   uciMove[2:4],
					UCI:  uciMove,
				}

				// Handle promotion moves
				if len(uciMove) == 5 {
					// Promotion moves have format like "a7a8q"
					bestMove.To = uciMove[2:4]
				}
			}
			break
		}
	}

	if bestMove == nil {
		return nil, fmt.Errorf("no best move found")
	}

	// Set the score and depth
	bestMove.Score = lastScore
	bestMove.Depth = depth
	bestMove.Evaluation = lastScore // Use the search score as evaluation

	// Get additional position evaluation if available
	if eval, err := e.GetEvaluation(fen); err == nil {
		bestMove.Evaluation = eval
	}

	return bestMove, nil
}

// UCIToAlgebraic converts a UCI move to algebraic notation using board state
func UCIToAlgebraic(uciMove string, fromPiece int, targetPiece int) (string, error) {
	if len(uciMove) < 4 {
		return "", fmt.Errorf("invalid UCI move: %s", uciMove)
	}

	from := uciMove[:2]
	to := uciMove[2:4]

	if fromPiece == 0 { // Empty
		return "", fmt.Errorf("no piece at source square %s", from)
	}

	// Check if it's a capture
	isCapture := targetPiece != 0 // Not empty

	// Convert based on piece type
	switch fromPiece {
	case 1, 7: // WP, BP
		return convertPawnMove(from, to, isCapture), nil
	case 2, 8: // WN, BN
		return "N" + to, nil
	case 3, 9: // WB, BB
		return "B" + to, nil
	case 4, 10: // WR, BR
		return "R" + to, nil
	case 5, 11: // WQ, BQ
		return "Q" + to, nil
	case 6, 12: // WK, BK
		// Check for castling
		if from == "e1" && (to == "g1" || to == "c1") ||
			from == "e8" && (to == "g8" || to == "c8") {
			if to[0] == 'g' {
				return "O-O", nil
			} else {
				return "O-O-O", nil
			}
		}
		return "K" + to, nil
	default:
		return "", fmt.Errorf("unknown piece type: %d", fromPiece)
	}
}

// convertPawnMove handles pawn move conversion
func convertPawnMove(from, to string, isCapture bool) string {
	if isCapture {
		// Pawn captures include the source file
		return from[:1] + "x" + to
	}
	// Regular pawn moves are just the destination
	return to
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

// Quit sends quit command to engine
func (e *Engine) Quit() error {
	if e.ready {
		return e.sendCommand("quit")
	}
	return nil
}

// Close closes the engine
func (e *Engine) Close() error {
	if e.cmd != nil && e.cmd.Process != nil {
		e.Quit()
		// Give engine time to quit gracefully
		time.Sleep(100 * time.Millisecond)
		e.cmd.Process.Kill()
		e.cmd.Wait()
	}
	return nil
}

// SetOption sets a UCI option (like Skill Level or UCI_Elo)
func (e *Engine) SetOption(name, value string) error {
	if !e.ready {
		return fmt.Errorf("engine not ready")
	}

	command := fmt.Sprintf("setoption name %s value %s", name, value)
	return e.sendCommand(command)
}

// SetSkillLevel sets the Stockfish skill level (0-20, where 20 is maximum strength)
func (e *Engine) SetSkillLevel(level int) error {
	if level < 0 || level > 20 {
		return fmt.Errorf("skill level must be between 0 and 20")
	}
	return e.SetOption("Skill Level", fmt.Sprintf("%d", level))
}

// SetEloRating sets a target ELO rating for Stockfish (1350-2850)
func (e *Engine) SetEloRating(elo int) error {
	if elo < 1350 || elo > 2850 {
		return fmt.Errorf("ELO rating must be between 1350 and 2850")
	}

	// Enable UCI_LimitStrength first
	if err := e.SetOption("UCI_LimitStrength", "true"); err != nil {
		return err
	}

	// Then set the target ELO
	return e.SetOption("UCI_Elo", fmt.Sprintf("%d", elo))
}

// DisableStrengthLimit disables ELO limiting (full strength)
func (e *Engine) DisableStrengthLimit() error {
	return e.SetOption("UCI_LimitStrength", "false")
}

// GetEvaluation gets the static evaluation of the current position
func (e *Engine) GetEvaluation(fen string) (int, error) {
	if !e.ready {
		return 0, fmt.Errorf("engine not ready")
	}

	// Set the position
	if err := e.sendCommand(fmt.Sprintf("position fen %s", fen)); err != nil {
		return 0, err
	}

	// Use a quick search instead of eval command (which might not be available)
	if err := e.sendCommand("go depth 1"); err != nil {
		return 0, err
	}

	var lastScore int

	// Read the search output
	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())

		// Parse info lines for score information
		if strings.HasPrefix(line, "info") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "score" && i+2 < len(parts) {
					if parts[i+1] == "cp" { // centipawn score
						if score, err := strconv.Atoi(parts[i+2]); err == nil {
							lastScore = score
						}
					}
				}
			}
		}

		// Break when we get the best move
		if strings.HasPrefix(line, "bestmove") {
			break
		}
	}

	return lastScore, nil
}
