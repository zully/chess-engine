#!/bin/bash

# Chess Engine with Stockfish Integration
# Startup script for the web-based chess GUI

echo "Starting Chess Engine with Stockfish..."

# Check if Stockfish binary exists
if [ ! -f "stockfish/stockfish-macos-m1-apple-silicon" ]; then
    echo "Error: Stockfish binary not found at stockfish/stockfish-macos-m1-apple-silicon"
    echo "Please ensure Stockfish is downloaded and extracted in the project directory."
    exit 1
fi

# Make sure Stockfish binary is executable
chmod +x stockfish/stockfish-macos-m1-apple-silicon

# Build the application if binary doesn't exist or source is newer
if [ ! -f "chess-stockfish" ] || [ "cmd/main.go" -nt "chess-stockfish" ]; then
    echo "Building chess-stockfish..."
    go build -o chess-stockfish ./cmd/main.go
    if [ $? -ne 0 ]; then
        echo "Error: Failed to build chess-stockfish"
        exit 1
    fi
fi

# Start the chess engine
echo "Starting web server on http://localhost:8080"
./chess-stockfish 