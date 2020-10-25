package server

import (
	"fmt"
	"testing"
)

func TestNewBoard(t *testing.T) {
	board := NewFullBoard()
	fmt.Println(board)
}
