package aconcagua

import "testing"

func TestFindMagicNumberForRook(t *testing.T) {
	sq := 0 // a1
	magic := findMagicNumber(sq, true)
	mask := rookMask(sq)

	maskConfigs := 1 << mask.count()
	posibleConfigurations := make([]Bitboard, maskConfigs)
	attacksConfigurations := make([]Bitboard, maskConfigs)
	magicAttacks := make([]Bitboard, maskConfigs)
	used := make([]bool, maskConfigs)

	// Generate all posible configurations for the blockers
	for i := range maskConfigs {
		posibleConfigurations[i] = generateBlockConfiguration(i, mask)
		attacksConfigurations[i] = rooksAttacksWithBlockers(sq, posibleConfigurations[i])
	}

	for i, config := range posibleConfigurations {
		index := (magic * config) >> (64 - mask.count())

		if used[index] {
			// check colision
			if attacksConfigurations[i] != magicAttacks[index] {
				t.Errorf("Collision found at index %d", index)
			}
		} else {
			used[index] = true
			magicAttacks[index] = attacksConfigurations[i]
		}
	}
}

func TestFindMagicNumberForBishop(t *testing.T) {
	sq := 28 // e4
	magic := findMagicNumber(sq, false)
	mask := bishopMask(sq)

	maskConfigs := 1 << mask.count()
	posibleConfigurations := make([]Bitboard, maskConfigs)
	attacksConfigurations := make([]Bitboard, maskConfigs)
	magicAttacks := make([]Bitboard, maskConfigs)
	used := make([]bool, maskConfigs)

	for i := range maskConfigs {
		posibleConfigurations[i] = generateBlockConfiguration(i, mask)
		attacksConfigurations[i] = bishopAttacksWithBlockers(sq, posibleConfigurations[i])
	}

	for i, config := range posibleConfigurations {
		index := (magic * config) >> (64 - mask.count())

		if used[index] {
			if attacksConfigurations[i] != magicAttacks[index] {
				t.Errorf("Collision found at index %d", index)
			}
		} else {
			used[index] = true
			magicAttacks[index] = attacksConfigurations[i]
		}
	}
}
