// Chess pieces mapping
const PIECE_SYMBOLS = {
    // White pieces
    1: '♙', 2: '♘', 3: '♗', 4: '♖', 5: '♕', 6: '♔',
    // Black pieces  
    7: '♟', 8: '♞', 9: '♝', 10: '♜', 11: '♛', 12: '♚'
};

const PIECE_NAMES = {
    1: 'WP', 2: 'WN', 3: 'WB', 4: 'WR', 5: 'WQ', 6: 'WK',
    7: 'BP', 8: 'BN', 9: 'BB', 10: 'BR', 11: 'BQ', 12: 'BK'
};

let gameState = null;
let autoPlayInterval = null;
let boardFlipped = false; // false = white perspective, true = black perspective
let selectedSquare = null; // Currently selected square for moves
let draggedPiece = null; // Currently being dragged piece

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    setupEventListeners();
    loadGameState();
});

function setupEventListeners() {
    // Move input and button
    const moveInput = document.getElementById('move-input');
    const makeMoveBtn = document.getElementById('make-move-btn');
    
    moveInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            makeMove();
        }
    });
    
    makeMoveBtn.addEventListener('click', makeMove);
    
    // Control buttons
    document.getElementById('engine-btn').addEventListener('click', requestEngineMove);
    document.getElementById('auto-btn').addEventListener('click', toggleAutoPlay);
    document.getElementById('flip-btn').addEventListener('click', flipBoard);
    document.getElementById('reset-btn').addEventListener('click', resetGame);
    
    // Set initial flip button text
    document.getElementById('flip-btn').textContent = 'View as Black';
}

async function loadGameState() {
    try {
        const response = await fetch('/api/state');
        gameState = await response.json();
        updateDisplay();
    } catch (error) {
        showMessage('Failed to load game state: ' + error.message, 'error');
    }
}

function updateDisplay() {
    if (!gameState) return;
    
    // Clear any selections or highlights
    clearSelection();
    
    renderBoard();
    updateMoveHistory();
    updateGameMessage();
}

function renderCoordinates() {
    // Render rank labels to match board orientation
    const rankLabels = document.getElementById('rank-labels-left');
    rankLabels.innerHTML = '';
    // White perspective: 8 at top, 1 at bottom
    // Black perspective: 1 at top, 8 at bottom  
    const rankNumbers = boardFlipped ? ['1','2','3','4','5','6','7','8'] : ['8','7','6','5','4','3','2','1'];
    
    rankNumbers.forEach(rank => {
        const div = document.createElement('div');
        div.textContent = rank;
        rankLabels.appendChild(div);
    });
    
    // Render file labels (a-h or h-a depending on orientation)
    const fileLabels = document.getElementById('file-labels-bottom');
    fileLabels.innerHTML = '';
    const fileLetters = boardFlipped ? ['h','g','f','e','d','c','b','a'] : ['a','b','c','d','e','f','g','h'];
    
    fileLetters.forEach(file => {
        const div = document.createElement('div');
        div.textContent = file;
        fileLabels.appendChild(div);
    });
}

function renderBoard() {
    const board = document.getElementById('chess-board');
    board.innerHTML = '';
    
    if (!gameState.board || !gameState.board.Squares) {
        showMessage('No board data available', 'error');
        return;
    }
    
    // Render coordinate labels
    renderCoordinates();
    
    // Determine rank and file order based on board orientation
    // For white perspective: show rank 8 at top, rank 1 at bottom
    // For black perspective: show rank 1 at top, rank 8 at bottom
    const ranks = boardFlipped ? [7,6,5,4,3,2,1,0] : [0,1,2,3,4,5,6,7];
    const files = boardFlipped ? [7,6,5,4,3,2,1,0] : [0,1,2,3,4,5,6,7];
    
    // Create squares
    for (let rankIdx = 0; rankIdx < 8; rankIdx++) {
        for (let fileIdx = 0; fileIdx < 8; fileIdx++) {
            const rank = ranks[rankIdx];
            const file = files[fileIdx];
            const square = document.createElement('div');
            square.className = 'square';
            square.dataset.rank = rank;
            square.dataset.file = file;
            // Convert array indices back to algebraic notation
            square.dataset.square = String.fromCharCode(97 + file) + (8 - rank);
            
            // Add light/dark class
            const isLight = (rank + file) % 2 === 0;
            square.classList.add(isLight ? 'light' : 'dark');
            
            // Get piece from board data
            const squareData = gameState.board.Squares[rank][file];
            if (squareData && squareData.Piece && squareData.Piece !== 0) {
                const piece = document.createElement('span');
                piece.className = 'piece';
                piece.textContent = PIECE_SYMBOLS[squareData.Piece] || '?';
                
                // Add color class
                if (squareData.Piece <= 6) {
                    piece.classList.add('white');
                } else {
                    piece.classList.add('black');
                }
                
                // Make piece draggable if it's the current player's turn
                const isWhitePiece = squareData.Piece <= 6;
                const isCurrentPlayerPiece = (gameState.board.WhiteToMove && isWhitePiece) || 
                                           (!gameState.board.WhiteToMove && !isWhitePiece);
                
                if (isCurrentPlayerPiece && !gameState.gameOver) {
                    piece.draggable = true;
                    piece.addEventListener('dragstart', handleDragStart);
                    square.classList.add('draggable');
                }
                
                square.appendChild(piece);
            }
            
            // Add click and drop handlers
            square.addEventListener('click', handleSquareClick);
            square.addEventListener('dragover', handleDragOver);
            square.addEventListener('drop', handleDrop);
            
            board.appendChild(square);
        }
    }
    
    // Highlight king if in check
    if (gameState.inCheck) {
        highlightKingInCheck();
    }
}

function highlightKingInCheck() {
    const squares = document.querySelectorAll('.square');
    squares.forEach(square => {
        const piece = square.querySelector('.piece');
        if (piece) {
            // Check if it's a king of the current player
            const squareData = getSquareData(square.dataset.rank, square.dataset.file);
            if (squareData && squareData.Piece) {
                const isWhiteKing = squareData.Piece === 1;
                const isBlackKing = squareData.Piece === 7;
                
                if ((gameState.board.WhiteToMove && isWhiteKing) || 
                    (!gameState.board.WhiteToMove && isBlackKing)) {
                    square.classList.add('check');
                }
            }
        }
    });
}

function getSquareData(rank, file) {
    if (!gameState.board || !gameState.board.Squares) return null;
    return gameState.board.Squares[rank][file];
}

function handleDragStart(e) {
    const square = e.target.parentElement;
    draggedPiece = {
        element: e.target,
        fromSquare: square.dataset.square,
        fromRank: parseInt(square.dataset.rank),
        fromFile: parseInt(square.dataset.file)
    };
    
    // Visual feedback
    square.classList.add('dragging-from');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', square.dataset.square);
    
    // Add some visual feedback
    setTimeout(() => {
        highlightPossibleMoves(square.dataset.square);
    }, 0);
}

function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
}

function handleDrop(e) {
    e.preventDefault();
    
    if (!draggedPiece) return;
    
    const toSquare = e.currentTarget;
    const move = constructMoveFromDrag(draggedPiece.fromSquare, toSquare.dataset.square);
    
    // Clear visual feedback
    clearHighlights();
    
    if (move) {
        // Execute the move
        executeMoveFromGUI(move);
    }
    
    draggedPiece = null;
}

function handleSquareClick(square) {
    if (gameState.gameOver) return;
    
    const piece = square.querySelector('.piece');
    
    // If clicking on empty square and we have a selected piece, try to move
    if (selectedSquare && square !== selectedSquare) {
        const move = constructMoveFromClick(selectedSquare.dataset.square, square.dataset.square);
        if (move) {
            executeMoveFromGUI(move);
        }
        clearSelection();
        return;
    }
    
    // If clicking on a piece that can be moved
    if (piece && canMovePiece(square)) {
        // Clear previous selection
        clearSelection();
        
        // Select this square
        selectedSquare = square;
        square.classList.add('selected');
        highlightPossibleMoves(square.dataset.square);
    } else {
        clearSelection();
    }
}

function canMovePiece(square) {
    const squareData = getSquareData(square.dataset.rank, square.dataset.file);
    if (!squareData || !squareData.Piece || squareData.Piece === 0) return false;
    
    const isWhitePiece = squareData.Piece <= 6;
    return (gameState.board.WhiteToMove && isWhitePiece) || 
           (!gameState.board.WhiteToMove && !isWhitePiece);
}

function constructMoveFromDrag(fromSquare, toSquare) {
    return constructMove(fromSquare, toSquare);
}

function constructMoveFromClick(fromSquare, toSquare) {
    return constructMove(fromSquare, toSquare);
}

function constructMove(fromSquare, toSquare) {
    if (fromSquare === toSquare) return null;
    
    console.log(`Constructing move from ${fromSquare} to ${toSquare}`); // Debug log
    
    // Get piece data using the square names directly from board
    const fromSquareData = getSquareDataBySquare(fromSquare);
    if (!fromSquareData || !fromSquareData.Piece || fromSquareData.Piece === 0) {
        console.log('No piece found at from square'); // Debug log
        return null;
    }
    
    const pieceType = getPieceType(fromSquareData.Piece);
    console.log(`Moving piece type: ${pieceType} (value: ${fromSquareData.Piece})`); // Debug log
    
    // Check if it's a capture
    const toSquareData = getSquareDataBySquare(toSquare);
    const isCapture = toSquareData && toSquareData.Piece && toSquareData.Piece !== 0;
    
    // Special case for castling
    if (pieceType === 'K') {
        if (fromSquare === 'e1' && toSquare === 'g1') return 'O-O';
        if (fromSquare === 'e1' && toSquare === 'c1') return 'O-O-O';
        if (fromSquare === 'e8' && toSquare === 'g8') return 'O-O';
        if (fromSquare === 'e8' && toSquare === 'c8') return 'O-O-O';
    }
    
    // Construct move notation
    let move = '';
    
    if (pieceType === 'P') {
        // Pawn moves
        if (isCapture) {
            move = fromSquare[0] + 'x' + toSquare;
        } else {
            move = toSquare;
        }
        
        // Check for pawn promotion (simplified - assumes Queen promotion)
        const toRank = parseInt(toSquare[1]);
        const isWhitePiece = fromSquareData.Piece <= 6;
        const promoteRank = isWhitePiece ? 8 : 1;
        if (toRank === promoteRank) {
            move += '=Q';
        }
    } else {
        // Other pieces - just use the simple notation
        move = pieceType;
        if (isCapture) {
            move += 'x';
        }
        move += toSquare;
    }
    
    console.log(`Constructed move notation: "${move}"`); // Debug log
    return move;
}

function getPieceType(pieceValue) {
    switch (pieceValue) {
        case 1: case 7: return 'P';
        case 2: case 8: return 'N';
        case 3: case 9: return 'B';
        case 4: case 10: return 'R';
        case 5: case 11: return 'Q';
        case 6: case 12: return 'K';
        default: return '';
    }
}

function highlightPossibleMoves(fromSquare) {
    // This is a simplified version - highlights empty squares and opponent pieces
    // In a full implementation, you'd calculate actual legal moves
    const squares = document.querySelectorAll('.square');
    const fromSquareData = getSquareDataBySquare(fromSquare);
    if (!fromSquareData) return;
    
    const isWhitePiece = fromSquareData.Piece <= 6;
    
    squares.forEach(square => {
        if (square.dataset.square === fromSquare) return;
        
        const squareData = getSquareData(square.dataset.rank, square.dataset.file);
        
        // Highlight empty squares or squares with opponent pieces
        if (!squareData || squareData.Piece === 0) {
            square.classList.add('possible-move');
        } else {
            const isTargetWhite = squareData.Piece <= 6;
            if (isWhitePiece !== isTargetWhite) {
                square.classList.add('possible-move');
            }
        }
    });
}

function getSquareDataBySquare(squareNotation) {
    const file = squareNotation.charCodeAt(0) - 97; // a=0, b=1, etc.
    const rank = 8 - parseInt(squareNotation[1]);   // 1=7, 2=6, ..., 8=0
    console.log(`Getting square data for ${squareNotation}: [${rank}][${file}]`); // Debug log
    
    if (rank < 0 || rank > 7 || file < 0 || file > 7) {
        console.log(`Invalid indices for ${squareNotation}`); // Debug log
        return null;
    }
    
    const squareData = gameState.board.Squares[rank][file];
    console.log(`Square data:`, squareData); // Debug log
    return squareData;
}

function clearHighlights() {
    const squares = document.querySelectorAll('.square');
    squares.forEach(square => {
        square.classList.remove('dragging-from', 'possible-move', 'drag-over');
    });
}

function clearSelection() {
    if (selectedSquare) {
        selectedSquare.classList.remove('selected');
        selectedSquare = null;
    }
    clearHighlights();
}

async function executeMoveFromGUI(move) {
    if (!move) return;
    
    console.log('Executing move:', move); // Debug log
    
    try {
        const response = await fetch('/api/move', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ move: move }),
        });
        
        gameState = await response.json();
        
        if (gameState.error) {
            console.log('Move error:', gameState.error); // Debug log
            showMessage(gameState.error, 'error');
        } else {
            updateDisplay();
        }
    } catch (error) {
        console.log('API error:', error); // Debug log
        showMessage('Failed to make move: ' + error.message, 'error');
    }
}

function updateMoveHistory() {
    const moveHistory = document.getElementById('move-history');
    if (!gameState.board || !gameState.board.MovesPlayed) {
        moveHistory.innerHTML = '';
        return;
    }
    
    const moves = gameState.board.MovesPlayed;
    let html = '';
    
    for (let i = 0; i < moves.length; i += 2) {
        const moveNumber = Math.floor(i / 2) + 1;
        const whiteMove = moves[i] || '';
        const blackMove = moves[i + 1] || '';
        
        html += `<div>${moveNumber}. ${whiteMove}`;
        if (blackMove) {
            html += ` ${blackMove}`;
        }
        html += '</div>';
    }
    
    moveHistory.innerHTML = html;
    moveHistory.scrollTop = moveHistory.scrollHeight;
}

function updateGameMessage() {
    const messageDiv = document.getElementById('game-message');
    let messageClass = 'info';
    let message = gameState.message || 'Ready to play';
    
    if (gameState.error) {
        message = gameState.error;
        messageClass = 'error';
    } else if (gameState.isCheckmate) {
        messageClass = 'success';
    } else if (gameState.inCheck) {
        messageClass = 'warning';
    }
    
    messageDiv.className = `message ${messageClass}`;
    messageDiv.textContent = message;
}

async function makeMove() {
    const moveInput = document.getElementById('move-input');
    const move = moveInput.value.trim();
    
    if (!move) {
        showMessage('Please enter a move', 'error');
        return;
    }
    
    const makeMoveBtn = document.getElementById('make-move-btn');
    makeMoveBtn.disabled = true;
    makeMoveBtn.classList.add('loading');
    
    try {
        const response = await fetch('/api/move', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ move: move }),
        });
        
        gameState = await response.json();
        
        if (gameState.error) {
            showMessage(gameState.error, 'error');
        } else {
            moveInput.value = '';
            updateDisplay();
        }
    } catch (error) {
        showMessage('Failed to make move: ' + error.message, 'error');
    } finally {
        makeMoveBtn.disabled = false;
        makeMoveBtn.classList.remove('loading');
    }
}

async function requestEngineMove() {
    const engineBtn = document.getElementById('engine-btn');
    engineBtn.disabled = true;
    engineBtn.classList.add('loading');
    
    try {
        const response = await fetch('/api/engine', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ depth: 4 }),
        });
        
        gameState = await response.json();
        updateDisplay();
        
        if (gameState.error) {
            showMessage(gameState.error, 'error');
        }
    } catch (error) {
        showMessage('Failed to get engine move: ' + error.message, 'error');
    } finally {
        engineBtn.disabled = false;
        engineBtn.classList.remove('loading');
    }
}

function toggleAutoPlay() {
    const autoBtn = document.getElementById('auto-btn');
    
    if (autoPlayInterval) {
        // Stop auto play
        clearInterval(autoPlayInterval);
        autoPlayInterval = null;
        autoBtn.textContent = 'Auto Play';
        autoBtn.style.background = '#ffc107';
        
        // Re-enable other buttons
        document.getElementById('make-move-btn').disabled = false;
        document.getElementById('engine-btn').disabled = false;
        document.getElementById('reset-btn').disabled = false;
        document.getElementById('move-input').disabled = false;
    } else {
        // Start auto play
        autoBtn.textContent = 'Stop Auto';
        autoBtn.style.background = '#dc3545';
        
        // Disable other buttons during auto play
        document.getElementById('make-move-btn').disabled = true;
        document.getElementById('engine-btn').disabled = true;
        document.getElementById('reset-btn').disabled = true;
        document.getElementById('move-input').disabled = true;
        
        autoPlayInterval = setInterval(async () => {
            if (gameState && !gameState.gameOver) {
                await requestEngineMove();
            } else {
                toggleAutoPlay(); // Stop if game is over
            }
        }, 2000); // 2 second delay between moves
    }
}

function flipBoard() {
    boardFlipped = !boardFlipped;
    renderBoard(); // Re-render the board with new orientation
    
    const flipBtn = document.getElementById('flip-btn');
    flipBtn.textContent = boardFlipped ? 'View as White' : 'View as Black';
}

async function resetGame() {
    const resetBtn = document.getElementById('reset-btn');
    resetBtn.disabled = true;
    resetBtn.classList.add('loading');
    
    // Stop auto play if running
    if (autoPlayInterval) {
        toggleAutoPlay();
    }
    
    try {
        const response = await fetch('/api/reset', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        
        gameState = await response.json();
        updateDisplay();
        
        // Clear move input
        document.getElementById('move-input').value = '';
        
        showMessage('Game reset successfully', 'success');
    } catch (error) {
        showMessage('Failed to reset game: ' + error.message, 'error');
    } finally {
        resetBtn.disabled = false;
        resetBtn.classList.remove('loading');
    }
}

function showMessage(text, type = 'info') {
    const messageDiv = document.getElementById('game-message');
    messageDiv.className = `message ${type}`;
    messageDiv.textContent = text;
    
    // Auto-clear error messages after 5 seconds
    if (type === 'error' || type === 'success') {
        setTimeout(() => {
            if (gameState) {
                updateGameMessage();
            }
        }, 5000);
    }
} 