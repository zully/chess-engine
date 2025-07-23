// Chess pieces mapping - now using SVG files
const PIECE_IMAGES = {
    // White pieces
    1: '/static/pieces/Chess_plt45.svg',  // White Pawn
    2: '/static/pieces/Chess_nlt45.svg',  // White Knight
    3: '/static/pieces/Chess_blt45.svg',  // White Bishop
    4: '/static/pieces/Chess_rlt45.svg',  // White Rook
    5: '/static/pieces/Chess_qlt45.svg',  // White Queen
    6: '/static/pieces/Chess_klt45.svg',  // White King
    // Black pieces
    7: '/static/pieces/Chess_pdt45.svg',  // Black Pawn
    8: '/static/pieces/Chess_ndt45.svg',  // Black Knight
    9: '/static/pieces/Chess_bdt45.svg',  // Black Bishop
    10: '/static/pieces/Chess_rdt45.svg', // Black Rook
    11: '/static/pieces/Chess_qdt45.svg', // Black Queen
    12: '/static/pieces/Chess_kdt45.svg'  // Black King
};

const PIECE_NAMES = {
    1: 'WP', 2: 'WN', 3: 'WB', 4: 'WR', 5: 'WQ', 6: 'WK',
    7: 'BP', 8: 'BN', 9: 'BB', 10: 'BR', 11: 'BQ', 12: 'BK'
};

let gameState = null;
let boardFlipped = false; // false = white perspective, true = black perspective
let selectedSquare = null; // Currently selected square for moves
let draggedPiece = null; // Currently being dragged piece

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    setupEventListeners();
    loadGameState();
});

function setupEventListeners() {
    try {
        // Control buttons
        document.getElementById('engine-btn').addEventListener('click', requestEngineMove);
        document.getElementById('undo-btn').addEventListener('click', undoLastMove);
        document.getElementById('flip-btn').addEventListener('click', flipBoard);
        document.getElementById('reset-btn').addEventListener('click', resetGame);
        
        // Engine checkboxes
        document.getElementById('engine-white-checkbox').addEventListener('change', handleEngineCheckboxChange);
        document.getElementById('engine-black-checkbox').addEventListener('change', handleEngineCheckboxChange);
        
        // Set initial flip button text
        document.getElementById('flip-btn').textContent = 'View as Black';
    } catch (error) {
        console.error('Error setting up event listeners:', error);
    }
}

// API Functions
function loadGameState() {
    fetch('/api/state')
        .then(response => response.json())
        .then(data => {
            gameState = data;
            updateGameState(data);
        })
        .catch(error => {
            console.error('Error loading game state:', error);
            document.getElementById('game-message').textContent = 'Error loading game state: ' + error.message;
        });
}

function makeMove(move) {
    fetch('/api/move', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({move: move})
    })
    .then(response => response.json())
    .then(data => {
        gameState = data;
        updateGameState(data);
        clearSelection();
    })
    .catch(error => {
        console.error('Error making move:', error);
        document.getElementById('game-message').textContent = 'Error making move: ' + error.message;
    });
}

function requestEngineMove() {
    const eloSelect = document.getElementById('elo-select');
    const selectedElo = parseInt(eloSelect.value);
    
    const requestData = {
        depth: 6,
        elo: selectedElo
    };

    fetch('/api/engine', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData)
    })
    .then(response => response.json())
    .then(data => {
        gameState = data;
        updateGameState(data);
        clearSelection();
    })
    .catch(error => {
        console.error('Error requesting engine move:', error);
        document.getElementById('game-message').textContent = 'Error with engine move: ' + error.message;
    });
}

function undoMove() {
    fetch('/api/undo', {
        method: 'POST'
    })
    .then(response => response.json())
    .then(data => {
        gameState = data;
        updateGameState(data);
        clearSelection();
    })
    .catch(error => {
        console.error('Error undoing move:', error);
        document.getElementById('game-message').textContent = 'Error undoing move: ' + error.message;
    });
}

function resetGame() {
    fetch('/api/reset', {
        method: 'POST'
    })
    .then(response => response.json())
    .then(data => {
        gameState = data;
        updateGameState(data);
        clearSelection();
        isFlipped = false;
        updateBoard(data.board);
    })
    .catch(error => {
        console.error('Error resetting game:', error);
        document.getElementById('game-message').textContent = 'Error resetting game: ' + error.message;
    });
}

function updateDisplay() {
    if (!gameState) return;
    
    // Clear any selections or highlights
    clearSelection();
    
    renderBoard();
    updateMoveHistory();
    updateGameMessage();
    updateEvaluationBar();
    updateCapturedPieces();
    
    // Check if we should make an automatic engine move
    checkForAutomaticEngineMove();
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
        div.className = 'rank-label';
        div.textContent = rank;
        rankLabels.appendChild(div);
    });
    
    // Render file labels (a-h or h-a depending on orientation)
    const fileLabels = document.getElementById('file-labels-bottom');
    fileLabels.innerHTML = '';
    const fileLetters = boardFlipped ? ['h','g','f','e','d','c','b','a'] : ['a','b','c','d','e','f','g','h'];
    
    fileLetters.forEach(file => {
        const div = document.createElement('div');
        div.className = 'file-label';
        div.textContent = file;
        fileLabels.appendChild(div);
    });
}

function renderBoard() {
    if (!gameState || !gameState.board) return;
    
    renderCoordinates();
    
    const board = document.getElementById('chess-board');
    board.innerHTML = '';
    
    const squares = gameState.board.Squares || [];
    
    for (let rank = 0; rank < 8; rank++) {
        for (let file = 0; file < 8; file++) {
            const square = document.createElement('div');
            const isLight = (rank + file) % 2 === 0;
            
            // Adjust for perspective - determine display position
            let displayRank = rank;
            let displayFile = file;
            
            if (boardFlipped) {
                displayRank = 7 - rank;
                displayFile = 7 - file;
            }
            
            square.className = `square ${isLight ? 'light' : 'dark'}`;
            square.dataset.rank = displayRank;
            square.dataset.file = displayFile;
            square.dataset.square = String.fromCharCode(97 + displayFile) + (8 - displayRank);
            
            // Add piece if present
            const squareData = squares[rank] && squares[rank][file];
            if (squareData && squareData.Piece && squareData.Piece !== 0) {
                const piece = document.createElement('img');
                piece.className = 'piece';
                piece.src = PIECE_IMAGES[squareData.Piece];
                piece.alt = PIECE_NAMES[squareData.Piece];
                piece.draggable = true;
                
                // Add color class for styling if needed
                if (squareData.Piece < 7) {
                    piece.classList.add('white');
                } else {
                    piece.classList.add('black');
                }
                
                // Add drag event listeners
                piece.addEventListener('dragstart', handleDragStart);
                piece.addEventListener('dragend', handleDragEnd);
                
                square.appendChild(piece);
            }
            
            // Add event listeners to squares
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
            const squareData = getSquareDataBySquare(square.dataset.square);
            if (squareData) {
                const isKing = (squareData.Piece === 6 || squareData.Piece === 12); // White or Black King
                const isCurrentPlayerKing = (gameState.board.WhiteToMove && squareData.Piece === 6) || 
                                          (!gameState.board.WhiteToMove && squareData.Piece === 12);
                
                if (isKing && isCurrentPlayerKing) {
                    piece.classList.add('check');
                } else {
                    piece.classList.remove('check');
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
    if (!gameState || gameState.gameOver) {
        e.preventDefault();
        return;
    }
    
    const piece = e.target;
    const square = piece.parentElement;
    const squareNotation = square.dataset.square;
    
    // Check if it's the current player's turn
    const pieceData = getSquareDataBySquare(squareNotation);
    if (!pieceData || pieceData.Piece === 0) {
        e.preventDefault();
        return;
    }
    
    const isWhitePiece = pieceData.Piece < 7;
    const isCurrentPlayerTurn = gameState.board.WhiteToMove === isWhitePiece;
    
    if (!isCurrentPlayerTurn) {
        e.preventDefault();
        return;
    }
    
    draggedPiece = {
        element: piece,
        from: squareNotation,
        pieceType: pieceData.Piece
    };
    
    piece.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', squareNotation);
}

function handleDragEnd(e) {
    if (draggedPiece) {
        draggedPiece.element.classList.remove('dragging');
        draggedPiece = null;
    }
    
    // Clear any drag-over highlights
    const squares = document.querySelectorAll('.square');
    squares.forEach(square => {
        square.classList.remove('drag-over');
    });
}

function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
}

function handleDrop(e) {
    e.preventDefault();
    
    if (!draggedPiece) return;
    
    const targetSquare = e.currentTarget;
    const toSquare = targetSquare.dataset.square;
    const fromSquare = draggedPiece.from;
    
    if (fromSquare === toSquare) {
        handleDragEnd(e);
        return;
    }
    
    // Construct proper algebraic move
    const move = constructAlgebraicMove(fromSquare, toSquare);
    if (move) {
        makeMove(move);
    }
    
    handleDragEnd(e);
}

function handleSquareClick(event) {
    const square = event.currentTarget;
    const squareNotation = square.dataset.square;
    
    if (!gameState || gameState.gameOver) {
        return;
    }
    
    // If no square is selected, select this square if it has a piece
    if (!selectedSquare) {
        const squareData = getSquareDataBySquare(squareNotation);
        if (squareData && squareData.Piece !== 0) {
            // Check if it's the current player's piece
            const isWhitePiece = squareData.Piece < 7;
            const isCurrentPlayerTurn = gameState.board.WhiteToMove === isWhitePiece;
            
            if (isCurrentPlayerTurn) {
                selectedSquare = square;
                square.classList.add('selected');
            }
        }
    } else {
        // A square is already selected, try to make a move
        const fromSquare = selectedSquare.dataset.square;
        const toSquare = squareNotation;
        
        // Clear selection
        selectedSquare.classList.remove('selected');
        selectedSquare = null;
        
        if (fromSquare !== toSquare) {
            // Try to make the move using algebraic notation
            const move = constructAlgebraicMove(fromSquare, toSquare);
            if (move) {
                makeMove(move);
            }
        }
    }
}

function constructAlgebraicMove(fromSquare, toSquare) {
    // Get piece information from the from square
    const fromSquareData = getSquareDataBySquare(fromSquare);
    if (!fromSquareData || fromSquareData.Piece === 0) {
        return null;
    }
    
    const pieceType = getPieceType(fromSquareData.Piece);
    const isWhitePiece = fromSquareData.Piece < 7;
    
    // Special case for castling
    if (pieceType === 'K') {
        if (fromSquare === 'e1' && toSquare === 'g1') return 'O-O';
        if (fromSquare === 'e1' && toSquare === 'c1') return 'O-O-O';
        if (fromSquare === 'e8' && toSquare === 'g8') return 'O-O';
        if (fromSquare === 'e8' && toSquare === 'c8') return 'O-O-O';
    }
    
    // For pawns, just return the target square (e.g., "e4")
    if (pieceType === 'P') {
        // Check if it's a capture
        const toSquareData = getSquareDataBySquare(toSquare);
        if (toSquareData && toSquareData.Piece !== 0) {
            // Pawn capture: include from file (e.g., "exd5")
            return fromSquare[0] + 'x' + toSquare;
        } else {
            // Check for en passant capture
            // (simplified - the backend will handle the complexity)
            return toSquare;
        }
    } else {
        // For other pieces, check if disambiguation is needed
        const toSquareData = getSquareDataBySquare(toSquare);
        const isCapture = toSquareData && toSquareData.Piece !== 0;
        
        // Check if there are other pieces of the same type that could also move to this square
        const disambiguation = getDisambiguation(fromSquare, toSquare, pieceType, isWhitePiece);
        
        let move = pieceType;
        if (disambiguation) {
            move += disambiguation;
        }
        if (isCapture) {
            move += 'x';
        }
        move += toSquare;
        
        return move;
    }
}

// Helper function to determine if disambiguation is needed and what form it should take
function getDisambiguation(fromSquare, toSquare, pieceType, isWhitePiece) {
    if (!gameState || !gameState.board) return '';
    
    const fromFile = fromSquare[0];
    const fromRank = fromSquare[1];
    const squares = gameState.board.Squares;
    
    // Find all other pieces of the same type and color that could move to the target square
    const conflictingPieces = [];
    
    for (let rank = 0; rank < 8; rank++) {
        for (let file = 0; file < 8; file++) {
            const square = squares[rank][file];
            if (!square || square.Piece === 0) continue;
            
            const squarePieceType = getPieceType(square.Piece);
            const squareIsWhite = square.Piece < 7;
            
            // Skip if not the same piece type and color
            if (squarePieceType !== pieceType || squareIsWhite !== isWhitePiece) continue;
            
            const squareName = square.Name;
            // Skip the piece we're actually moving
            if (squareName === fromSquare) continue;
            
            // Check if this piece could also move to the target square
            if (canPieceMoveTo(squareName, toSquare, pieceType)) {
                conflictingPieces.push(squareName);
            }
        }
    }
    
    // If no conflicting pieces, no disambiguation needed
    if (conflictingPieces.length === 0) {
        return '';
    }
    
    // Check if file disambiguation is sufficient
    const sameFile = conflictingPieces.some(square => square[0] === fromFile);
    if (!sameFile) {
        return fromFile; // Use file letter (e.g., "Ra" in "Rae8")
    }
    
    // Check if rank disambiguation is sufficient
    const sameRank = conflictingPieces.some(square => square[1] === fromRank);
    if (!sameRank) {
        return fromRank; // Use rank number (e.g., "R1" in "R1e8")
    }
    
    // If both file and rank have conflicts, use both (rare case)
    return fromSquare; // Use full square (e.g., "Ra1" in "Ra1e8")
}

// Helper function to check if a piece at a given square can move to a target square
function canPieceMoveTo(fromSquare, toSquare, pieceType) {
    const fromSquareData = getSquareDataBySquare(fromSquare);
    if (!fromSquareData || fromSquareData.Piece === 0) return false;
    
    const fromFile = fromSquare.charCodeAt(0) - 97; // a=0, b=1, etc.
    const fromRank = 8 - parseInt(fromSquare[1]);   // 1=7, 2=6, ..., 8=0
    const toFile = toSquare.charCodeAt(0) - 97;
    const toRank = 8 - parseInt(toSquare[1]);
    
    // Basic movement validation (simplified)
    switch (pieceType) {
        case 'N': // Knight
            const rankDiff = Math.abs(toRank - fromRank);
            const fileDiff = Math.abs(toFile - fromFile);
            return (rankDiff === 2 && fileDiff === 1) || (rankDiff === 1 && fileDiff === 2);
            
        case 'B': // Bishop
            const bishopRankDiff = Math.abs(toRank - fromRank);
            const bishopFileDiff = Math.abs(toFile - fromFile);
            if (bishopRankDiff !== bishopFileDiff) return false;
            return isPathClear(fromRank, fromFile, toRank, toFile);
            
        case 'R': // Rook
            if (fromRank !== toRank && fromFile !== toFile) return false;
            return isPathClear(fromRank, fromFile, toRank, toFile);
            
        case 'Q': // Queen
            const queenRankDiff = Math.abs(toRank - fromRank);
            const queenFileDiff = Math.abs(toFile - fromFile);
            const isDiagonal = queenRankDiff === queenFileDiff;
            const isStraight = fromRank === toRank || fromFile === toFile;
            if (!isDiagonal && !isStraight) return false;
            return isPathClear(fromRank, fromFile, toRank, toFile);
            
        case 'K': // King
            const kingRankDiff = Math.abs(toRank - fromRank);
            const kingFileDiff = Math.abs(toFile - fromFile);
            return kingRankDiff <= 1 && kingFileDiff <= 1;
            
        default:
            return false;
    }
}

// Helper function to check if the path between two squares is clear
function isPathClear(fromRank, fromFile, toRank, toFile) {
    if (!gameState || !gameState.board) return false;
    
    const rankStep = toRank === fromRank ? 0 : (toRank > fromRank ? 1 : -1);
    const fileStep = toFile === fromFile ? 0 : (toFile > fromFile ? 1 : -1);
    
    let currentRank = fromRank + rankStep;
    let currentFile = fromFile + fileStep;
    
    while (currentRank !== toRank || currentFile !== toFile) {
        if (gameState.board.Squares[currentRank][currentFile].Piece !== 0) {
            return false; // Path blocked
        }
        currentRank += rankStep;
        currentFile += fileStep;
    }
    
    return true;
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
    
    // Get piece data using the square names directly from board
    const fromSquareData = getSquareDataBySquare(fromSquare);
    if (!fromSquareData || !fromSquareData.Piece || fromSquareData.Piece === 0) {
        return null;
    }
    
    const pieceType = getPieceType(fromSquareData.Piece);
    
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
        // Other pieces 
        move = pieceType;
        
        if (isCapture) {
            move += 'x';
        }
        move += toSquare;
    }
    
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
    // Disabled - no visible dots during drag
    return;
}

function getSquareDataBySquare(squareNotation) {
    const file = squareNotation.charCodeAt(0) - 97; // a=0, b=1, etc.
    const rank = 8 - parseInt(squareNotation[1]);   // 1=7, 2=6, ..., 8=0
    
    if (rank < 0 || rank > 7 || file < 0 || file > 7) {
        return null;
    }
    
    const squareData = gameState.board.Squares[rank][file];
    return squareData;
}

function clearHighlights() {
    const squares = document.querySelectorAll('.square');
    squares.forEach(square => {
        square.classList.remove('dragging-from', 'possible-move', 'drag-over', 'check');
        // Also remove dragging and check classes from any pieces
        const piece = square.querySelector('.piece');
        if (piece) {
            piece.classList.remove('dragging', 'check');
        }
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
            updateDisplay();
        }
    } catch (error) {
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
    } else if (gameState.draw) {
        message = `Draw! ${gameState.drawReason}`;
        messageClass = 'success';
    } else if (gameState.isCheckmate) {
        // Use the backend message for checkmate announcement
        message = gameState.message || 'Checkmate!';
        messageClass = 'success';
    } else if (gameState.inCheck) {
        // Use the backend message for check announcement  
        message = gameState.message || 'Check!';
        messageClass = 'warning';
    } else if (gameState.threefoldRep) {
        message += ` (Position repeated ${gameState.positionCount} times - draw available!)`;
        messageClass = 'warning';
    } else if (gameState.positionCount >= 2) {
        message += ` (Position repeated ${gameState.positionCount} times)`;
        messageClass = 'warning';
    }
    
    messageDiv.className = `message ${messageClass}`;
    messageDiv.textContent = message;
}

function updateButtonStates() {
    const undoBtn = document.getElementById('undo-btn');
    
    // Enable undo button only if there are moves to undo and game is not over
    const canUndo = gameState.board && 
                   gameState.board.MovesPlayed && 
                   gameState.board.MovesPlayed.length > 0 && 
                   !gameState.gameOver;
    

    
    undoBtn.disabled = !canUndo;
}

// makeMove function removed - using drag-and-drop only

function handleEngineCheckboxChange() {
    // When checkboxes change, check if we should make an automatic move
    checkForAutomaticEngineMove();
}

function shouldEnginePlay(isWhite) {
    const whiteCheckbox = document.getElementById('engine-white-checkbox');
    const blackCheckbox = document.getElementById('engine-black-checkbox');
    
    return isWhite ? whiteCheckbox.checked : blackCheckbox.checked;
}

async function checkForAutomaticEngineMove() {
    // Don't make automatic moves if game is over
    if (!gameState || gameState.gameOver) return;
    
    // Check if current player should be played by engine
    if (shouldEnginePlay(gameState.board.WhiteToMove)) {
        // Small delay to allow UI updates
        setTimeout(() => {
            requestEngineMove();
        }, 500);
    }
}

async function undoLastMove() {
    if (gameState.gameOver) {
        showMessage('Cannot undo in finished game. Reset to start over.', 'error');
        return;
    }
    
    if (!gameState.board || !gameState.board.MovesPlayed || gameState.board.MovesPlayed.length === 0) {
        showMessage('No moves to undo!', 'error');
        return;
    }
    
    try {
        showMessage('Undoing last move...', 'info');
        const response = await fetch('/api/undo', {
            method: 'POST'
        });
        
        gameState = await response.json();
        updateDisplay();
        
        if (gameState.error) {
            showMessage('Undo failed: ' + gameState.error, 'error');
        } else {
            showMessage('Move undone!', 'success');
        }
    } catch (error) {
        showMessage('Failed to undo move: ' + error.message, 'error');
    }
}

// toggleAutoPlay function removed - no auto play button

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
    
    try {
        const response = await fetch('/api/reset', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        
        gameState = await response.json();
        updateDisplay();
        
        showMessage('Game reset successfully', 'success');
    } catch (error) {
        showMessage('Failed to reset game: ' + error.message, 'error');
    } finally {
        resetBtn.disabled = false;
        resetBtn.classList.remove('loading');
    }
}

function updateEvaluationBar() {
    const evaluationFill = document.getElementById('evaluation-fill');
    const evaluationText = document.getElementById('evaluation-text');
    
    if (!gameState || !evaluationFill || !evaluationText) return;
    
    const evaluation = gameState.evaluation || 0;
    
    // Convert centipawns to a more readable format
    const displayValue = (evaluation / 100).toFixed(2);
    
    // Calculate percentage for the bar (clamp between -1000 and +1000 centipawns)
    const clampedEval = Math.max(-1000, Math.min(1000, evaluation));
    const percentage = 50 + (clampedEval / 1000) * 50; // 50% center, ±50% for ±1000cp
    
    // Update bar
    evaluationFill.style.width = percentage + '%';
    
    // Color based on evaluation
    if (evaluation > 50) {
        evaluationFill.style.background = 'linear-gradient(90deg, #d4edda, #28a745)';
    } else if (evaluation < -50) {
        evaluationFill.style.background = 'linear-gradient(90deg, #f8d7da, #dc3545)';
    } else {
        evaluationFill.style.background = 'linear-gradient(90deg, #f8f9fa, #f8f9fa)';
    }
    
    // Update text
    evaluationText.textContent = displayValue;
    evaluationText.style.color = evaluation > 0 ? '#28a745' : evaluation < 0 ? '#dc3545' : '#495057';
}

function updateCapturedPieces() {
    const capturedWhiteDiv = document.getElementById('captured-white');
    const capturedBlackDiv = document.getElementById('captured-black');
    const capturedWhiteValue = document.getElementById('captured-white-value');
    const capturedBlackValue = document.getElementById('captured-black-value');
    
    if (!gameState || !capturedWhiteDiv || !capturedBlackDiv) return;
    
    const capturedWhite = gameState.capturedWhite || [];
    const capturedBlack = gameState.capturedBlack || [];
    
    // Clear current display
    capturedWhiteDiv.innerHTML = '';
    capturedBlackDiv.innerHTML = '';
    
    let whiteValue = 0;
    let blackValue = 0;
    
    // Display captured pieces by white
    capturedWhite.forEach(piece => {
        const pieceElement = document.createElement('span');
        pieceElement.className = 'captured-piece';
        pieceElement.textContent = getPieceSymbol(piece.type, false); // Black pieces captured by white
        capturedWhiteDiv.appendChild(pieceElement);
        whiteValue += piece.value;
    });
    
    // Display captured pieces by black
    capturedBlack.forEach(piece => {
        const pieceElement = document.createElement('span');
        pieceElement.className = 'captured-piece';
        pieceElement.textContent = getPieceSymbol(piece.type, true); // White pieces captured by black
        capturedBlackDiv.appendChild(pieceElement);
        blackValue += piece.value;
    });
    
    // Update values
    if (capturedWhiteValue) capturedWhiteValue.textContent = whiteValue;
    if (capturedBlackValue) capturedBlackValue.textContent = blackValue;
}

function getPieceSymbol(pieceType, isWhite) {
    const symbols = {
        'P': isWhite ? '♟' : '♟',
        'N': isWhite ? '♞' : '♞', 
        'B': isWhite ? '♝' : '♝',
        'R': isWhite ? '♜' : '♜',
        'Q': isWhite ? '♛' : '♛',
        'K': isWhite ? '♚' : '♚'
    };
    return symbols[pieceType] || '';
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

// Update game state display including Stockfish version
function updateGameState(data) {
    // Update the complete display using existing function
    updateDisplay();
    
    // Update Stockfish version
    if (data.stockfishVersion) {
        document.getElementById('stockfish-version').textContent = data.stockfishVersion;
    }
} 