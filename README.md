# Chess Engine with Stockfish Integration

A modern web-based chess engine powered by **Stockfish 17.1** and built with **Go** and **JavaScript**. Features a clean, responsive UI with drag-and-drop gameplay, real-time position evaluation, and configurable engine strength.

## ✨ Features

### 🎮 **Interactive Gameplay**
- **Drag-and-drop pieces** - Intuitive mouse controls
- **Click-to-move** - Alternative input method  
- **Move validation** - Prevents illegal moves with check validation
- **Visual feedback** - Blue highlighting for last move, red highlighting for check
- **Castling support** - Handles O-O and O-O-O notation
- **En passant** - Full pawn capture rules
- **Pawn promotion** - Automatic queen promotion
- **Board flipping** - Fully functional "View as Black" perspective

### 🤖 **Stockfish Integration**  
- **UCI Protocol** - Direct communication with Stockfish 17.1
- **Configurable strength** - Descriptive ELO labels (Beginner to Grand Master)
- **Real-time evaluation** - Live position analysis
- **Engine vs Engine** - Watch Stockfish play itself
- **Best move suggestions** - Get hints from the world's strongest engine

### 📊 **Advanced Features**
- **Position evaluation bar** - Visual advantage indicator
- **Captured pieces tracker** - See material balance with piece values
- **Move history** - Complete algebraic notation with proper disambiguation
- **Last move highlighting** - Blue squares show the most recent move
- **Check detection** - Red king highlighting and status messages
- **Board flipping** - Play from either perspective with proper piece reorientation
- **FEN support** - Standard position notation
- **Draw detection** - Stalemate and repetition handling

### 🎨 **Modern UI**
- **Responsive design** - Works on desktop and mobile
- **Lichess-style pieces** - High-quality SVG graphics
- **Clean interface** - Minimal, focused design
- **Real-time updates** - Instant feedback and validation
- **Visual indicators** - Color-coded feedback for moves, checks, and captures

## 🏗️ Architecture

### **Frontend (Pure JavaScript)**
- UCI move generation (e.g., `e2e4`, `a1e1`)
- Drag-and-drop interaction
- Real-time UI updates with visual feedback
- SVG piece rendering with proper board orientation
- Last move highlighting and check detection

### **Backend (Go)**
- **Simplified UCI-first design** - Reduced complexity after algebraic notation cleanup
- UCI move validation and execution
- Enhanced check validation logic
- FEN position management  
- Board state tracking with last move information
- RESTful API endpoints

### **Engine (Stockfish 17.1)**
- Built from source in Docker
- UCI protocol communication
- Position evaluation
- Best move calculation

## 🚀 Quick Start

### Docker Deployment (Recommended)

```bash
# Build the container
docker build -t chess-engine:latest .

# Run the application
docker run -p 8080:8080 chess-engine:latest
```

### Local Development

```bash
# Build and run locally
go build -o chess-engine ./cmd/main.go
./chess-engine
```

**Access the game:** http://localhost:8080

## 🎯 How to Play

1. **Make moves** by dragging pieces or clicking squares
2. **Set engine strength** using descriptive labels (Beginner, Intermediate, Advanced, Expert, etc.)
3. **Request engine moves** using the "Engine Move" button
4. **Watch visual feedback** - blue squares show last move, red king shows check
5. **Flip board** using "View as Black" button for different perspective
6. **View evaluation** in the real-time evaluation bar
7. **Track captures** in the material balance display

### Move Input Formats
- **Drag & Drop** - Natural piece movement
- **Click to Move** - Click source, then destination
- **All moves use UCI notation internally** (e.g., e2e4, g1f3)
- **Move history displays proper algebraic notation** (e.g., "Rae1", "Nbd2")

## 📡 API Endpoints

### Game Management
- `GET /api/state` - Current game state with last move and check status
- `POST /api/move` - Make a move (UCI format)
- `POST /api/engine` - Request engine move
- `POST /api/undo` - Undo last move  
- `POST /api/reset` - Reset game

### Enhanced Game State Response
```json
{
  "board": { /* board state */ },
  "inCheck": true,
  "isCheckmate": false,
  "lastUCIMove": "e2e4",
  "evaluation": 150,
  "message": "White is in check!"
}
```

### Move Request Format
```json
{
  "move": "e2e4"  // UCI notation: fromSquare + toSquare
}
```

### Engine Request Format  
```json
{
  "depth": 6,     // Search depth (1-15)
  "elo": 1800     // Engine strength (1350-2850)
}
```

## 🔧 Technical Highlights

### **Recent Improvements (2024)**
- ✅ **Last move highlighting** - Blue squares show recent moves
- ✅ **Enhanced check detection** - Red king highlighting with status messages
- ✅ **Code simplification** - Removed 120+ lines of unused algebraic notation code
- ✅ **Fixed board flipping** - Pieces now properly reorient when viewing as Black
- ✅ **Improved disambiguation** - Move history shows proper notation (e.g., "Rae1")
- ✅ **Enhanced validation** - Better check detection and illegal move prevention

### **UCI Integration**
- Native Stockfish communication
- Simplified move parsing (no disambiguation needed)
- Direct position evaluation
- Engine strength control

### **Performance**
- Efficient board representation
- Fast move validation
- Minimal memory footprint
- Docker containerization
- **Cleaned codebase** - Removed unused functions after UCI refactor

### **Reliability**
- Comprehensive move validation with check detection
- Error handling and recovery
- Draw detection
- Position repetition tracking

## 🏆 Engine Strength

Configure Stockfish strength with descriptive labels:
- **1350 Intermediate A** - Learning fundamentals
- **1500 Intermediate B** - Developing strategy
- **1600 Advanced A** - Tactical awareness
- **1800 Advanced B** - Strategic thinking
- **2000 Expert** - Club-level play
- **2200 National Master** - Tournament strength
- **2400 International Master** - Near-professional
- **2500 Grand Master** - World-class play
- **Full Strength** - Maximum engine power (~3500 ELO)

## 🐳 Docker Details

The application runs in a lightweight Ubuntu container with:
- **Stockfish 17.1** - Built from official source (stable branch)
- **Go 1.21** - Backend server
- **Static assets** - Optimized web resources

**Container size:** ~200MB  
**Startup time:** <5 seconds

## 🌟 Why This Implementation?

### **UCI Throughout**
- ✅ **Eliminates disambiguation complexity** (no more "Rae1" vs "Rfe1" parsing)
- ✅ **Native Stockfish format** - direct engine communication  
- ✅ **Simpler frontend** - just fromSquare + toSquare
- ✅ **More reliable** - fewer parsing errors
- ✅ **Single conversion point** - UCI → algebraic only for history
- ✅ **Cleaner codebase** - Removed complex unused algebraic parsing

### **Enhanced User Experience**
- ✅ **Visual feedback** - Last move highlighting and check indicators
- ✅ **Proper board flipping** - Pieces reorient correctly
- ✅ **Intuitive controls** - Drag-and-drop with click alternatives
- ✅ **Real-time validation** - Immediate feedback on illegal moves

### **Stockfish Integration**
- ✅ **World's strongest engine** - FIDE rating ~3500
- ✅ **Configurable strength** - suitable for all skill levels
- ✅ **Real-time evaluation** - instant position analysis
- ✅ **Standard UCI protocol** - industry standard

---

**Built with ❤️ using Go, JavaScript, and Stockfish**