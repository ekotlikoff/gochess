package model

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	game := NewGame()
	if game.turn != White {
		t.Error("Expected turn = ", White, " got ", game.turn)
	}
}
