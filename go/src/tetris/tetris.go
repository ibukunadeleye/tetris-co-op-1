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
	Counter    int
	Bottom     int
	Left       int
	Right      int
	Landed     []([]int)
	CurrPiece1 *Piece
	CurrPiece2 *Piece
	Init1      Coord //position where new pieces for player 1 should appear
	Init2      Coord //position where new pieces for player 2 should appear
}

func NewBoard(height int, width int, initCoord1 Coord, initCoord2 Coord) *Board {
	rows := make([]([]int), height)
	for i := 0; i < height; i++ {
		rows[i] = make([]int, width)
	}

	return &Board{
		Counter: 3,
		Bottom:  height - 1,
		Left:    0,
		Right:   width - 1,
		Landed:  rows,
		CurrPiece1: &Piece{
			ID:  1,
			Pos: &Coord{Row: initCoord1.Row, Col: initCoord1.Col}},
		CurrPiece2: &Piece{
			ID:  2,
			Pos: &Coord{Row: initCoord2.Row, Col: initCoord2.Col}},
		Init1: initCoord1,
		Init2: initCoord2}
}

func (b *Board) Print() {
	fmt.Println("Printing board...")
	for i := 0; i < (b.Bottom + 1); i++ {
		fmt.Printf("%v \n", b.Landed[i])
	}
}

//returns true if piece at pos touched a piece UNDER it
func (b *Board) Touched(p Piece) bool {
	//sanity check
	if p.Pos.Row >= b.Bottom {
		fmt.Println("Touched Error: Piece is at bottom")
		return true
	}

	under := b.Landed[((p.Pos.Row) + 1)][p.Pos.Col]
	if under != 0 {
		fmt.Printf("Piece %d touched Piece %d \n", p.ID, under)
		return true
	}
	return false
}

//returns true if board is not yet filled
//returns false if board is filled, signifying that game is over
func (b *Board) MoveDown(i int) ([]Update, bool) {
	var p Piece
	var curr *Coord
	var otherCurr Coord
	var newCoord Coord

	if i == 1 {
		p = *b.CurrPiece1
		newCoord = b.Init1
		curr = b.CurrPiece1.Pos
		otherCurr = *b.CurrPiece2.Pos
	} else {
		p = *b.CurrPiece2
		newCoord = b.Init2
		curr = b.CurrPiece2.Pos
		otherCurr = *b.CurrPiece1.Pos
	}

	oldRow := curr.Row

	//check if piece should land before it moves down
	//if the piece should land, then the board is filled to the top and
	//game should be over
	if b.Touched(p) {
		//sanity check
		if b.Landed[curr.Row][curr.Col] != 0 {
			fmt.Println("MoveDown Error: moved down on landed piece")
			return []Update{}, false
		}

		if !(curr.Row == 0 && curr.Col == 1) {
			fmt.Println("MoveDown Error: Piece should have already landed")
			return []Update{}, false
		}

		b.Landed[curr.Row][curr.Col] = p.ID
		landUpdate := Update{
			Pos:   Coord{Row: curr.Row, Col: curr.Col},
			Value: 2}

		fmt.Println("Game Over")
		return []Update{landUpdate}, false
	}

	//check if the other piece is in the way before the current piece moves down
	if (otherCurr.Row == (curr.Row + 1)) && (otherCurr.Col == curr.Col) {
		return []Update{}, true
	}

	curr.Row = curr.Row + 1

	//erase piece at previous position on client's board
	remUpdate := Update{
		Pos:   Coord{Row: oldRow, Col: curr.Col},
		Value: 0}

	if curr.Row == b.Bottom || b.Touched(p) {

		//sanity check
		if b.Landed[curr.Row][curr.Col] != 0 {
			fmt.Println("MoveDown Error: moved down on landed piece")
			return []Update{}, true
		}

		//mark the piece as landed
		b.Landed[curr.Row][curr.Col] = p.ID
		fmt.Printf("Landed Piece %d  at (%d, %d)\n", p.ID, curr.Row, curr.Col)

		landedRow := curr.Row
		landedCol := curr.Col
		b.Counter = b.Counter + 1
		p.ID = b.Counter
		curr.Row = newCoord.Row
		curr.Col = newCoord.Col

		//generate new piece
		newUpdate := Update{
			Pos:   newCoord,
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

func (b *Board) MoveLeft(i int) []Update {
	var curr *Coord
	var otherCurr Coord

	if i == 1 {
		curr = b.CurrPiece1.Pos
		otherCurr = *b.CurrPiece2.Pos
	} else {
		curr = b.CurrPiece2.Pos
		otherCurr = *b.CurrPiece1.Pos
	}

	oldRow := curr.Row
	oldCol := curr.Col

	//check if other piece is in the way
	if (otherCurr.Col == (curr.Col - 1)) && (otherCurr.Row == curr.Row) {
		return []Update{}
	}

	//check if there's a landed piece in the way
	if b.Landed[curr.Row][curr.Col-1] != 0 {
		return []Update{}
	}

	//check if spurious command
	if curr.Col > b.Left {
		curr.Col = curr.Col - 1
	} else {
		return []Update{}
	}

	remUpdate := Update{
		Pos:   Coord{Row: oldRow, Col: oldCol},
		Value: 0}

	fillUpdate := Update{
		Pos:   Coord{Row: curr.Row, Col: curr.Col},
		Value: 1}

	return []Update{remUpdate, fillUpdate}
}

func (b *Board) MoveRight(i int) []Update {
	var curr *Coord
	var otherCurr Coord

	if i == 1 {
		curr = b.CurrPiece1.Pos
		otherCurr = *b.CurrPiece2.Pos
	} else {
		curr = b.CurrPiece2.Pos
		otherCurr = *b.CurrPiece1.Pos
	}

	oldRow := curr.Row
	oldCol := curr.Col

	//check if other piece is in the way
	if (otherCurr.Col == (curr.Col + 1)) && (otherCurr.Row == curr.Row) {
		return []Update{}
	}

	//check if there's a landed piece in the way
	if b.Landed[curr.Row][curr.Col+1] != 0 {
		return []Update{}
	}

	//check if spurious command
	if curr.Col < b.Right {
		curr.Col = curr.Col + 1
	} else {
		return []Update{}
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
