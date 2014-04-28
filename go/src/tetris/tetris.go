package tetris

import "fmt"

/*type Tetrini struct {
	Pos    *Coord
}*/

type Coord struct {
	Row int
	Col int
}

type Update struct {
	//the first element represents the row index  and the second is column
	Pos   Coord
	Value int
}

type Piece struct {
	ID  int
	Pos *Coord
}

type Board struct {
	Counter   int
	Bottom    int
	Left      int
	Right     int
	Landed    []([]int)
	CurrPiece *Piece
}

func NewBoard(height, width int) *Board {
	rows := make([]([]int), height)
	for i := 0; i < height; i++ {
		rows[i] = make([]int, width)
	}

	return &Board{
		Counter:   0,
		Bottom:    height - 1,
		Left:      0,
		Right:     width - 1,
		Landed:    rows,
		CurrPiece: &Piece{ID: 1, Pos: &Coord{Row: 0, Col: 1}}}
}

func (b *Board) Print() {
	fmt.Println("Printing board...")
	for i := 0; i < (b.Bottom + 1); i++ {
		fmt.Printf("%v \n", b.Landed[i])
	}
}

//returns true if piece at pos touched a piece UNDER it
func (b *Board) Touched() bool {
	//sanity check
	if b.CurrPiece.Pos.Row >= b.Bottom {
		fmt.Println("Touched Error: Piece is at bottom")
		return true
	}

	under := b.Landed[b.CurrPiece.Pos.Row+1][b.CurrPiece.Pos.Col]
	if under != 0 {
		fmt.Printf("Piece %d touched Piece %d \n", b.CurrPiece.ID, under)
		return true
	}
	return false
}

//returns true if board is not yet filled
//returns false if board is filled, signifying that game is over
func (b *Board) MoveDown() ([]Update, bool) {
	curr := b.CurrPiece.Pos
	oldRow := curr.Row

	//this would only happen the board is filled up to the top
	if b.Touched() {
		//sanity check
		if b.Landed[curr.Row][curr.Col] != 0 {
			fmt.Println("MoveDown Error: moved down on landed piece")
			return []Update{}, false
		}

		if !(curr.Row == 0 && curr.Col == 1) {
			fmt.Println("MoveDown Error: Piece should have already landed")
			return []Update{}, false
		}

		b.Landed[curr.Row][curr.Col] = b.CurrPiece.ID
		landUpdate := Update{
			Pos:   Coord{Row: curr.Row, Col: curr.Col},
			Value: 2}

		fmt.Println("Game Over")
		return []Update{landUpdate}, false
	}

	curr.Row = curr.Row + 1

	//erase piece at previous position on client's board
	remUpdate := Update{
		Pos:   Coord{Row: oldRow, Col: curr.Col},
		Value: 0}

	if curr.Row == b.Bottom || b.Touched() {

		//sanity check
		if b.Landed[curr.Row][curr.Col] != 0 {
			fmt.Println("MoveDown Error: moved down on landed piece")
			return []Update{}, true
		}

		//mark the piece as landed
		b.Landed[curr.Row][curr.Col] = b.CurrPiece.ID
		fmt.Printf("Landed Piece %d  at (%d, %d)\n", b.CurrPiece.ID, curr.Row, curr.Col)

		landedRow := curr.Row
		landedCol := curr.Col
		b.Counter = b.Counter + 1
		b.CurrPiece.ID = b.Counter
		curr.Row = 0
		curr.Col = 1

		//generate new piece
		newUpdate := Update{
			Pos:   Coord{Row: 0, Col: 1},
			Value: 1}

		landUpdate := Update{
			Pos:   Coord{Row: landedRow, Col: landedCol},
			Value: 2}

		return []Update{remUpdate, landUpdate, newUpdate}, true

	} else {
		fillUpdate := Update{
			Pos:   Coord{Row: curr.Row, Col: curr.Col},
			Value: 1}

		return []Update{remUpdate, fillUpdate}, true
	}
}

func (b *Board) MoveLeft() []Update {
	curr := b.CurrPiece.Pos

	oldRow := curr.Row
	oldCol := curr.Col

	//check if spurious command
	if curr.Col > b.Left {
		curr.Col = curr.Col - 1
	}

	remUpdate := Update{
		Pos:   Coord{Row: oldRow, Col: oldCol},
		Value: 0}

	fillUpdate := Update{
		Pos:   Coord{Row: curr.Row, Col: curr.Col},
		Value: 1}

	return []Update{remUpdate, fillUpdate}
}

func (b *Board) MoveRight() []Update {
	curr := b.CurrPiece.Pos

	oldRow := curr.Row
	oldCol := curr.Col

	//check if spurious command
	if curr.Col < b.Right {
		curr.Col = curr.Col + 1
	}
	remUpdate := Update{
		Pos:   Coord{Row: oldRow, Col: oldCol},
		Value: 0}

	fillUpdate := Update{
		Pos:   Coord{Row: curr.Row, Col: curr.Col},
		Value: 1}

	return []Update{remUpdate, fillUpdate}
}

//sends command to client to display the current piece on the board
func (b *Board) Display() {
	return
}

//checks to see if any lines have been formed in the pieces that have landed
//if so, tells the Conn handler to tell the client to update the board
func (b *Board) ClearLines() {
	return
}
