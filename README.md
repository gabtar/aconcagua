# Aconcagua

Chess engine in go - WIP

# Main parts of a chess engine

- [ ] Move generation/validation (returns all legal moves from a position)
- [ ] Evaluation function (scores out a position saying which side is better)
- [ ] Search algoritm (finds the best next move among all posibles moves based on the evaluation function)


# TODO, Move generation:
- [ ] Validate moves when king is in check
- [ ] Pawn moves generation
- [ ] Add castle moves to king(if available)
- [ ] Add en passant and pawn first move
- [ ] From a given position return all posibles next moves/positions(will be used later for searching best move)
    - [ ] Tests if given a position returns all availables moves correctly

