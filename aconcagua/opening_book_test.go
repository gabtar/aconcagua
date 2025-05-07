package aconcagua

import (
	"encoding/binary"
	"os"
	"testing"
)

func TestPolyglotHashKey(t *testing.T) {
	// Known Positions to test polyglot hash function
	polyglotHashSampleCases := []struct {
		fen  string
		hash uint64
	}{
		{fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", hash: 0x463b96181691fc9c},      // starting position
		{fen: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", hash: 0x823c9b50fd114196},   // position after e2e4
		{fen: "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2", hash: 0x0756b94461c50fb0}, // position after e2e4 d75
		{fen: "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2", hash: 0x662fafb965db29d4},   // position after e2e4 d7d5 e4e5
		{fen: "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3", hash: 0x22a48b5a8e47ff78}, // position after e2e4 d7d5 e4e5 f7f5
		{fen: "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR b kq - 0 3", hash: 0x652a607ca3f242c1},    // position after e2e4 d7d5 e4e5 f7f5 e1e2
		{fen: "rnbq1bnr/ppp1pkpp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR w - - 0 4", hash: 0x00fdd303c946bdd9},     // position after e2e4 d7d5 e4e5 f7f5 e1e2 e8f7
		{fen: "rnbq1bnr/ppp1pkpp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR w - - 0 4", hash: 0x00fdd303c946bdd9},     // position after e2e4 d7d5 e4e5 f7f5 e1e2 e8f7
		{fen: "rnbqkbnr/p1pppppp/8/8/PpP4P/8/1P1PPPP1/RNBQKBNR b KQkq c3 0 3", hash: 0x3c8123ea7b067637}, // position after a2a4 b7b5 h2h4 b5b4 c2c4
		{fen: "rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b Kkq - 0 4", hash: 0x5c3f9b829b279560},  // position after a2a4 b7b5 h2h4 b5b4 c2c4 b4c3 a1a3
	}

	for _, tc := range polyglotHashSampleCases {
		t.Run(tc.fen, func(t *testing.T) {
			pos := From(tc.fen)
			polyglotHash := PolyglotHashFromPosition(pos)

			if polyglotHash != tc.hash {
				t.Errorf("Expected: %v, got: %v", tc.hash, polyglotHash)
			}
		})
	}
}

func TestLoadBook(t *testing.T) {
	data := []PolyglotBookEntry{
		{Key: 0x463b96181691fc9c, Move: 0x0000, Weight: 0x0000, Learn: 0x00000000},
		{Key: 0x823c9b50fd114196, Move: 0x0000, Weight: 0x0000, Learn: 0x00000000},
		{Key: 0x0756b94461c50fb0, Move: 0x0000, Weight: 0x0000, Learn: 0x00000000},
		{Key: 0x662fafb965db29d4, Move: 0x0000, Weight: 0x0000, Learn: 0x00000000},
		{Key: 0x22a48b5a8e47ff78, Move: 0x0000, Weight: 0x0000, Learn: 0x00000000},
	}

	tmpfile, err := os.CreateTemp("", "sample.bin")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up after the test

	err = binary.Write(tmpfile, binary.LittleEndian, &data)
	if err != nil {
		t.Fatal(err)
	}

	polyglotBook := PolyglotBook{}
	err = polyglotBook.Load(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if polyglotBook.Size != 5 {
		t.Errorf("Expected: %v, got: %v", 5, polyglotBook.Size)
	}
}
