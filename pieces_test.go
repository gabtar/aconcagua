package main

import "testing"

func TestKingAttacks(t *testing.T){
  pos := InitialPosition()
  king, _ := pos.pieceAt("e1")

  expected := Bitboard(0b11100000101000)
  got := king.attacks(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
