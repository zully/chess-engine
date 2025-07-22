# Chess Engine

A fully functional chess engine with a modern web-based GUI implemented in Go. Play chess in your browser with complete rule validation, intelligent AI opponent, and all standard chess features.

## Features

### 🎯 **Complete Chess Implementation**
- **All piece movements**: Pawns, Knights, Bishops, Rooks, Queens, Kings with proper rules
- **Special moves**: En passant, castling (kingside and queenside), pawn promotion (to any piece)
- **Move validation**: Comprehensive rule enforcement including check/checkmate detection
- **Drag-and-drop interface**: Intuitive click-and-drag piece movement
- **Turn-based gameplay**: Proper alternating turns between White and Black

### 🤖 **Intelligent AI Engine**
- **Smart opponent**: AI that plays strategically using minimax with alpha-beta pruning
- **Position evaluation**: Advanced scoring with material, positional, and king safety factors
- **Opening knowledge**: Database of good opening moves for stronger early game play
- **Difficulty scaling**: Configurable search depth for different skill levels
- **Draw awareness**: Threefold repetition detection and stalemate recognition

### 🎮 **Modern Web Interface**
- **Beautiful visual board**: Clean, responsive design that works on desktop and mobile
- **Real-time updates**: Live game state with immediate feedback
- **Move history**: Complete game record with standard algebraic notation
- **Interactive controls**: Engine moves, auto-play mode, board flipping, game reset
- **Status indicators**: Clear notifications for check, checkmate, draws, and game state

### ⚡ **Advanced Features**
- **Threefold repetition**: Automatic draw detection when positions repeat
- **Position tracking**: Real-time monitoring of position frequency
- **Strategic repetition**: Engine avoids repetition when winning, allows it when losing
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
│   │   └── piece_moves.go   # Individual piece movement rules
│   ├── engine/
│   │   ├── evaluation.go    # Position evaluation with advanced heuristics
│   │   ├── search.go        # Minimax search with alpha-beta pruning
│   │   └── movegen.go       # Complete legal move generation
│   └── moves/
│       └── moves.go         # Algebraic notation parsing
├── web/
│   └── static/
│       ├── chess.css        # Modern responsive styling
│       └── chess.js         # Interactive frontend logic
├── test/
│   └── engine_test.go       # Unit tests
├── start-web.sh             # Quick start script
├── go.mod
└── go.sum
```

## Quick Start

### Installation
```bash
git clone https://github.com/zully/chess-engine.git
cd chess-engine
```

### Start the Game
```bash
# Option 1: Use the quick start script
chmod +x start-web.sh
./start-web.sh

# Option 2: Run directly
go run cmd/main.go
```

### Play Chess
1. **Open your browser** to `http://localhost:8080`
2. **Play moves** by clicking and dragging pieces or typing in algebraic notation
3. **Challenge the AI** by clicking "Engine Move" 
4. **Enable auto-play** to watch the computer play both sides
5. **Flip the board** to play from Black's perspective

### Game Controls
- **Manual Move**: Type moves like `e4`, `Nf3`, `O-O`, `exd5`, `a1=Q`
- **Drag & Drop**: Click and drag pieces to move them
- **Engine Move**: Let the AI make a move for the current player
- **Auto Play**: Computer plays both sides automatically
- **Flip Board**: Change perspective between White and Black
- **Reset Game**: Start a new game

## Example Game Features

### Smart Move Input
```
Manual moves: e4, d6, Nf3, Nf6, Bc4, Bg4
Engine suggestions with evaluation scores
Drag-and-drop piece movement
```

### Intelligent AI
- **Opening play**: Develops pieces logically in the opening
- **Tactical awareness**: Finds checks, captures, and threats
- **Strategic depth**: Long-term planning with configurable search depth
- **Endgame knowledge**: Improved king activity in simplified positions

### Advanced Rule Support
- **En passant captures**: Automatic detection and execution
- **Castling**: Both kingside (O-O) and queenside (O-O-O) supported
- **Pawn promotion**: Promote to Queen, Rook, Bishop, or Knight
- **Draw detection**: Threefold repetition and stalemate recognition

## API Endpoints

The web server provides RESTful API endpoints:

- `GET /api/state` - Get current game state
- `POST /api/move` - Make a player move
- `POST /api/engine` - Request engine move
- `POST /api/reset` - Reset the game

## Technical Highlights

- **2,000+ lines** of clean, well-structured Go code
- **Zero external dependencies** - pure Go implementation with native HTTP server
- **Complete chess rules** including all special moves and draw conditions
- **Advanced AI** with minimax search, alpha-beta pruning, and position evaluation
- **Modern web interface** with responsive design and intuitive controls
- **Real-time updates** using REST API with JSON communication

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
- **Touch devices**: Drag-and-drop optimized for touch screens

## Contributing

Contributions are welcome! This is a clean, well-structured codebase that's easy to extend.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details