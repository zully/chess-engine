package pieces

type PieceType int

const (
    Pawn PieceType = iota
    Knight
    Bishop
    Rook
    Queen
    King
)

type Piece struct {
    Type  PieceType
    Color string // "white" or "black"
}

// Movement rules for each piece can be defined here
func (p *Piece) CanMove(from, to Position) bool {
    // Implement movement logic based on piece type
    return false
}

type Position struct {
    X int
    Y int
}