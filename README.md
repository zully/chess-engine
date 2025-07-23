# Chess Engine with Stockfish Integration

A fully functional chess web application powered by the world-class **Stockfish chess engine**. Play chess in your browser with complete rule validation, powerful AI analysis, and all standard chess features.

## Features

### 🎯 **Complete Chess Implementation**
- **All piece movements**: Pawns, Knights, Bishops, Rooks, Queens, Kings with proper rules
- **Special moves**: En passant, castling (kingside and queenside), pawn promotion (to any piece)
- **Move validation**: Comprehensive rule enforcement including check/checkmate detection
- **Drag-and-drop interface**: Intuitive click-and-drag piece movement
- **Turn-based gameplay**: Proper alternating turns between White and Black

### 🤖 **Stockfish Engine Integration**
- **World-class AI**: Powered by Stockfish, the strongest open-source chess engine
- **UCI Protocol**: Full Universal Chess Interface implementation for engine communication
- **Advanced Analysis**: Deep positional evaluation with precise move scoring
- **Configurable Depth**: Adjustable search depth for different difficulty levels
- **Real-time Feedback**: Live evaluation scores and search depth information

### 🎮 **Modern Web Interface**
- **Beautiful visual board**: Clean, responsive design that works on desktop and mobile
- **Real-time updates**: Live game state with immediate feedback
- **Move history**: Complete game record with standard algebraic notation
- **Interactive controls**: Engine moves, auto-play mode, board flipping, game reset
- **Status indicators**: Clear notifications for check, checkmate, draws, and game state

### ⚡ **Advanced Features**
- **FEN Support**: Full Forsyth-Edwards Notation for position import/export
- **Threefold repetition**: Automatic draw detection when positions repeat
- **Position tracking**: Real-time monitoring of position frequency
- **Board orientation**: Flip board to play from either perspective
- **Game management**: Save/restore game state, move validation, error handling

## Project Structure

```
chess-engine/
├── cmd/
│   └── main.go              # Web server and API endpoints
├── internal/
│   ├── board/
│   │   ├── board.go         # Board representation and position tracking
│   │   ├── coordinates.go   # Square coordinate utilities
│   │   ├── moves.go         # Move execution and validation
│   │   ├── piece_moves.go   # Individual piece movement rules
│   │   └── fen.go           # FEN notation import/export
│   ├── uci/
│   │   └── uci.go           # UCI protocol implementation for Stockfish
│   └── moves/
│       └── moves.go         # Algebraic notation parsing
├── stockfish/
│   └── stockfish-macos-m1-apple-silicon  # Stockfish binary
├── web/
│   └── static/
│       ├── chess.css        # Modern responsive styling
│       └── chess.js         # Interactive frontend logic
├── chess-stockfish.sh       # Quick start script
├── go.mod
└── go.sum
```

## Quick Start

### Prerequisites
- Go 1.19 or later
- Stockfish binary (automatically downloaded)

### Installation
```bash
git clone https://github.com/zully/chess-engine.git
cd chess-engine
```

### Start the Game
```bash
# Option 1: Use the quick start script (recommended)
./chess-stockfish.sh

# Option 2: Manual start
go build -o chess-stockfish ./cmd/main.go
./chess-stockfish
```

### Play Chess
1. **Open your browser** to `http://localhost:8080`
2. **Play moves** by clicking and dragging pieces or typing in algebraic notation
3. **Challenge Stockfish** by clicking "Engine Move" 
4. **Enable auto-play** to watch Stockfish play both sides
5. **Flip the board** to play from Black's perspective

### Game Controls
- **Manual Move**: Type moves like `e4`, `Nf3`, `O-O`, `exd5`, `a1=Q`
- **Drag & Drop**: Click and drag pieces to move them
- **Engine Move**: Let Stockfish make a move for the current player
- **Auto Play**: Stockfish plays both sides automatically
- **Flip Board**: Change perspective between White and Black
- **Reset Game**: Start a new game
- **Undo Move**: Take back the last move

## Stockfish Integration

This application uses **Stockfish 17**, the world's strongest open-source chess engine, providing:

- **Tactical Excellence**: Finds complex tactical combinations and threats
- **Positional Understanding**: Deep evaluation of position structure and strategy
- **Opening Knowledge**: Extensive opening book and theory
- **Endgame Mastery**: Precise endgame technique and tablebases
- **Configurable Strength**: Adjustable search depth from beginner to grandmaster level

### Engine Features
- **Real-time Analysis**: Live position evaluation during play
- **Move Suggestions**: Best move recommendations with scoring
- **Depth Control**: Search depth from 1-15 moves ahead
- **UCI Communication**: Standard Universal Chess Interface protocol

## API Endpoints

The web server provides RESTful API endpoints:

- `GET /api/state` - Get current game state with FEN position
- `POST /api/move` - Make a player move
- `POST /api/engine` - Request Stockfish engine move
- `POST /api/reset` - Reset the game
- `POST /api/undo` - Undo the last move

## Technical Highlights

- **Stockfish Integration**: Full UCI protocol implementation
- **FEN Support**: Complete Forsyth-Edwards Notation handling  
- **Zero external dependencies**: Pure Go implementation with native HTTP server
- **Complete chess rules**: Including all special moves and draw conditions
- **Modern web interface**: Responsive design with intuitive controls
- **Real-time updates**: REST API with JSON communication

## Development

### Run Tests
```bash
go test ./...
```

### Build
```bash
go build ./...
```

### Development Server
```bash
go run cmd/main.go
# Server starts on http://localhost:8080
```

## Browser Compatibility

- **Chrome/Safari/Firefox**: Full support for all features
- **Mobile browsers**: Responsive design works on phones and tablets

## Stockfish Credits

This application is powered by [Stockfish](https://stockfishchess.org/), developed by the Stockfish team. Stockfish is free software licensed under the GNU General Public License v3.

## License

This project is open source and available under the MIT License.