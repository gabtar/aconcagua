package board

import "testing"

// Perft tests
// Data from https://gist.github.com/peterellisjones/8c46c28141c162d1d8a0f0badbc9cff9

func TestPerft1(t *testing.T) {
	pos := From("r6r/1b2k1bq/8/8/7B/8/8/R3K2R b KQ - 3 2")

	expected := uint64(8)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft2(t *testing.T) {
	pos := From("8/8/8/2k5/2pP4/8/B7/4K3 b - d3 0 3")

	expected := uint64(8)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft3(t *testing.T) {
	pos := From("r1bqkbnr/pppppppp/n7/8/8/P7/1PPPPPPP/RNBQKBNR w KQkq - 2 2")

	expected := uint64(19)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft4(t *testing.T) {
	pos := From("r3k2r/p1pp1pb1/bn2Qnp1/2qPN3/1p2P3/2N5/PPPBBPPP/R3K2R b KQkq - 3 2")

	expected := uint64(5)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft5(t *testing.T) {
	pos := From("2kr3r/p1ppqpb1/bn2Qnp1/3PN3/1p2P3/2N5/PPPBBPPP/R3K2R b KQ - 3 2")

	expected := uint64(44)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft6(t *testing.T) {
	pos := From("rnb2k1r/pp1Pbppp/2p5/q7/2B5/8/PPPQNnPP/RNB1K2R w KQ - 3 9")

	expected := uint64(39)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft7(t *testing.T) {
	pos := From("2r5/3pk3/8/2P5/8/2K5/8/8 w - - 5 4")

	expected := uint64(9)
	got := pos.Perft(1)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft8(t *testing.T) {
	pos := From("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")

	expected := uint64(62379)
	got := pos.Perft(3)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft9(t *testing.T) {
	pos := From("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10")

	expected := uint64(89890)
	got := pos.Perft(3)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft10(t *testing.T) {
	pos := From("3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1")

	expected := uint64(1134888)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft11(t *testing.T) {
	pos := From("8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1")

	expected := uint64(1015133)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft12(t *testing.T) {
	pos := From("8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1")

	expected := uint64(1440467)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft13(t *testing.T) {
	pos := From("5k2/8/8/8/8/8/8/4K2R w K - 0 1")

	expected := uint64(661072)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft14(t *testing.T) {
	pos := From("3k4/8/8/8/8/8/8/R3K3 w Q - 0 1")

	expected := uint64(803711)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft15(t *testing.T) {
	pos := From("r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1")

	expected := uint64(1274206)
	got := pos.Perft(4)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft16(t *testing.T) {
	pos := From("r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1")

	expected := uint64(1720476)
	got := pos.Perft(4)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft17(t *testing.T) {
	pos := From("2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1")

	expected := uint64(3821001)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft18(t *testing.T) {
	pos := From("8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1")

	expected := uint64(1004658)
	got := pos.Perft(5)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft19(t *testing.T) {
	pos := From("4k3/1P6/8/8/8/8/K7/8 w - - 0 1")

	expected := uint64(217342)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft20(t *testing.T) {
	pos := From("8/P1k5/K7/8/8/8/8/8 w - - 0 1")

	expected := uint64(92683)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft21(t *testing.T) {
	pos := From("K1k5/8/P7/8/8/8/8/8 w - - 0 1")

	expected := uint64(2217)
	got := pos.Perft(6)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft22(t *testing.T) {
	pos := From("8/k1P5/8/1K6/8/8/8/8 w - - 0 1")

	expected := uint64(567584)
	got := pos.Perft(7)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPerft23(t *testing.T) {
	pos := From("8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1")

	expected := uint64(23527)
	got := pos.Perft(4)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
