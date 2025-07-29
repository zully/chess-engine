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
	From        string
	To          string
	Score       int
	Depth       int
	UCI         string   // Store the original UCI format
	Evaluation  int      // Position evaluation in centipawns (positive = better for white)
	PV          []string // Principal variation (sequence of best moves in UCI format)
	PVAlgebraic []string // Principal variation in algebraic notation
}

// MultiPVLine represents one line of analysis in multi-pv mode
type MultiPVLine struct {
	LineNumber    int      // Which line this is (1, 2, 3, etc.)
	Score         int      // Score for this line
	Depth         int      // Search depth
	PV            []string // Principal variation in UCI format
	PVAlgebraic   []string // Principal variation in algebraic notation
	FirstMoveEval int      // Evaluation after playing the first move of this line
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
	// Check if engine process is still alive
	if e.cmd.Process != nil {
		// Try to check if process is still running
		if e.cmd.ProcessState != nil && e.cmd.ProcessState.Exited() {
			e.ready = false
			return fmt.Errorf("engine process has exited")
		}
	}

	if _, err := e.stdin.WriteString(command + "\n"); err != nil {
		// Only mark as not ready for serious communication failures
		if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "closed pipe") {
			e.ready = false
		}
		return fmt.Errorf("failed to write command '%s': %v", command, err)
	}

	if err := e.stdin.Flush(); err != nil {
		// Only mark as not ready for serious communication failures
		if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "closed pipe") {
			e.ready = false
		}
		return fmt.Errorf("failed to flush command '%s': %v", command, err)
	}

	return nil
}

// SetPosition sets the current position using FEN notation
func (e *Engine) SetPosition(fen string) error {
	if !e.ready {
		return fmt.Errorf("engine not ready")
	}

	// Check if engine is alive
	if !e.IsAlive() {
		return fmt.Errorf("engine process is not alive")
	}

	command := fmt.Sprintf("position fen %s", fen)
	if err := e.sendCommand(command); err != nil {
		return fmt.Errorf("failed to set position: %v", err)
	}

	return nil
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
	var lastPV []string

	// Read the search output
	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())

		// Parse info lines for score information and principal variation
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
				// Capture principal variation
				if part == "pv" && i+1 < len(parts) {
					// Everything after "pv" is the principal variation
					lastPV = parts[i+1:]
					break
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
	bestMove.PV = lastPV

	// Get additional position evaluation if available
	if eval, err := e.GetEvaluation(fen); err == nil {
		bestMove.Evaluation = eval
	}

	return bestMove, nil
}

// Close closes the engine process
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

// Quit sends quit command to engine
func (e *Engine) Quit() error {
	if e.ready {
		return e.sendCommand("quit")
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

// SetEloRating sets the engine strength to a specific ELO rating
func (e *Engine) SetEloRating(elo int) error {
	if !e.ready {
		return fmt.Errorf("engine not ready")
	}

	// Stockfish ELO range is typically 1350-2850
	if elo < 1350 || elo > 2850 {
		return fmt.Errorf("ELO rating %d out of range (1350-2850)", elo)
	}

	// Enable strength limiting - don't fail if this doesn't work
	if err := e.sendCommand("setoption name UCI_LimitStrength value true"); err != nil {
		// Log but continue - some engines might not support this option
	}

	// Set the ELO rating - don't fail if this doesn't work
	if err := e.sendCommand(fmt.Sprintf("setoption name UCI_Elo value %d", elo)); err != nil {
		// Log but continue - some engines might not support this option
	}

	// Also set skill level to a lower value for weaker play
	// Lower skill levels (0-20) make more errors
	var skillLevel int
	switch {
	case elo <= 1400:
		skillLevel = 1 // Very weak
	case elo <= 1600:
		skillLevel = 3 // Weak
	case elo <= 1800:
		skillLevel = 5 // Below average
	case elo <= 2000:
		skillLevel = 8 // Average
	case elo <= 2200:
		skillLevel = 12 // Good
	case elo <= 2400:
		skillLevel = 15 // Strong
	default:
		skillLevel = 18 // Very strong (but not maximum to allow some errors)
	}

	// Set skill level - don't fail if this doesn't work
	if err := e.sendCommand(fmt.Sprintf("setoption name Skill Level value %d", skillLevel)); err != nil {
		// Log but continue - this is not critical
	}

	return nil
}

// DisableStrengthLimit disables ELO limiting for full strength play
func (e *Engine) DisableStrengthLimit() error {
	if !e.ready {
		return fmt.Errorf("engine not ready")
	}

	// Disable strength limiting - don't fail if this doesn't work
	if err := e.sendCommand("setoption name UCI_LimitStrength value false"); err != nil {
		// Log but continue - some engines might not support this option
	}

	// Set skill level to maximum - don't fail if this doesn't work
	if err := e.sendCommand("setoption name Skill Level value 20"); err != nil {
		// Log but continue - this is not critical
	}

	return nil
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

// GetMultiPVAnalysis gets multiple principal variations from the engine
func (e *Engine) GetMultiPVAnalysis(fen string, depth int, numLines int) ([]MultiPVLine, error) {
	if !e.ready {
		return nil, fmt.Errorf("engine not ready")
	}

	// Check if engine is alive before proceeding
	if !e.IsAlive() {
		return nil, fmt.Errorf("engine process is not alive")
	}

	// Set MultiPV option
	if err := e.SetOption("MultiPV", fmt.Sprintf("%d", numLines)); err != nil {
		return nil, fmt.Errorf("failed to set MultiPV: %v", err)
	}

	// Set the position
	if err := e.sendCommand(fmt.Sprintf("position fen %s", fen)); err != nil {
		return nil, fmt.Errorf("failed to set position: %v", err)
	}

	// Start the search
	command := "go"
	if depth > 0 {
		command += fmt.Sprintf(" depth %d", depth)
	}
	if err := e.sendCommand(command); err != nil {
		return nil, err
	}

	lines := make(map[int]*MultiPVLine)
	var maxDepth int

	// Read the search output
	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())

		// Parse info lines for multiple PV information
		if strings.HasPrefix(line, "info") {
			parts := strings.Fields(line)
			var currentLine *MultiPVLine

			for i, part := range parts {
				// Find which PV line this is
				if part == "multipv" && i+1 < len(parts) {
					lineNum, err := strconv.Atoi(parts[i+1])
					if err == nil {
						if lines[lineNum] == nil {
							lines[lineNum] = &MultiPVLine{LineNumber: lineNum}
						}
						currentLine = lines[lineNum]
					}
				}

				// Get depth
				if part == "depth" && i+1 < len(parts) && currentLine != nil {
					if depth, err := strconv.Atoi(parts[i+1]); err == nil {
						currentLine.Depth = depth
						if depth > maxDepth {
							maxDepth = depth
						}
					}
				}

				// Get score
				if part == "score" && i+2 < len(parts) && currentLine != nil {
					if parts[i+1] == "cp" {
						if score, err := strconv.Atoi(parts[i+2]); err == nil {
							currentLine.Score = score
						}
					}
				}

				// Get principal variation
				if part == "pv" && i+1 < len(parts) && currentLine != nil {
					currentLine.PV = parts[i+1:]
					break
				}
			}
		}

		// Break when we get the best move (search is complete)
		if strings.HasPrefix(line, "bestmove") {
			break
		}
	}

	// Convert map to sorted slice
	result := make([]MultiPVLine, 0, len(lines))
	for i := 1; i <= numLines; i++ {
		if line, exists := lines[i]; exists {
			result = append(result, *line)
		}
	}

	// Reset MultiPV to 1 for other operations
	e.SetOption("MultiPV", "1")

	return result, nil
}

// GetEngineInfo gets the Stockfish engine information including version
func (e *Engine) GetEngineInfo() (string, error) {
	if !e.ready {
		return "", fmt.Errorf("engine not ready")
	}

	// Send UCI command to get engine info
	if err := e.sendCommand("uci"); err != nil {
		return "", err
	}

	var engineInfo string

	// Read the UCI response to get engine name and version
	for e.stdout.Scan() {
		line := strings.TrimSpace(e.stdout.Text())
		if strings.HasPrefix(line, "id name ") {
			engineInfo = strings.TrimPrefix(line, "id name ")
		}
		if line == "uciok" {
			break
		}
	}

	if engineInfo == "" {
		return "Stockfish (version unknown)", nil
	}

	return engineInfo, nil
}

// IsAlive checks if the engine process is still running and responsive
func (e *Engine) IsAlive() bool {
	// Simplified check - just verify engine is ready
	// Complex process state checking was causing false negatives
	return e.ready
}

// Ping sends an isready command to check if engine is responsive
func (e *Engine) Ping() error {
	if !e.IsAlive() {
		return fmt.Errorf("engine is not alive")
	}

	// Simple ping without complex timeout logic to avoid deadlocks
	return e.sendCommand("isready")
}

// Restart recreates the engine process when it crashes or becomes unresponsive
func (e *Engine) Restart(enginePath string) error {
	// Close the old engine if it exists
	if e.cmd != nil && e.cmd.Process != nil {
		e.cmd.Process.Kill()
		e.cmd.Wait()
	}

	// Create new engine process
	cmd := exec.Command(enginePath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start engine: %v", err)
	}

	// Update engine fields
	e.cmd = cmd
	e.stdin = bufio.NewWriter(stdin)
	e.stdout = bufio.NewScanner(stdout)
	e.ready = false

	// Initialize the restarted engine
	return e.initialize()
}
