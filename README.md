# Chess Engine

This project is a chess engine implemented in Go. It provides functionalities for evaluating chess positions, generating legal moves, and searching for the best move.

## Project Structure

```
chess-engine
├── cmd
│   └── main.go          # Entry point of the application
├── internal
│   ├── board
│   │   └── board.go     # Board state representation and manipulation
│   ├── engine
│   │   ├── evaluation.go # Functions for evaluating board positions
│   │   ├── movegen.go    # Move generation logic
│   │   └── search.go      # Search algorithm for best move
│   ├── moves
│   │   └── moves.go      # Move struct and related functions
│   └── pieces
│       └── pieces.go     # Chess pieces definitions and movement rules
├── pkg
│   └── utils
│       └── utils.go      # Utility functions for the project
├── test
│   └── engine_test.go    # Unit tests for the engine
├── go.mod                 # Module definition and dependencies
└── go.sum                 # Dependency checksums
```

## Setup Instructions

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd chess-engine
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the application:
   ```bash
   go run cmd/main.go
   ```

## Features

- **Board Representation**: The chessboard state is represented using a `Board` struct, allowing for easy manipulation and move validation.
- **Move Generation**: The engine can generate all legal moves for the current position.
- **Position Evaluation**: The engine evaluates board positions and assigns scores based on material and positional advantages.
- **Search Algorithm**: Implements a search algorithm to find the best move within a specified depth.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.