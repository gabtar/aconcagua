# Aconcagua

Chess engine in go - WIP

Setup:
```
git clone https://github.com/gabtar/aconcagua
cd aconcagua
go build .
```

Builds an `aconcagua` executable, a uci compatible engine that can be used with a gui like arena gui or pychess 


### TODO:

#### Move generation:
- [x] Refactor move encode
- [x] Refactor square to bitboard function
- [ ] Fix bug. Check last test added (need to debug, maybe unmakeMove and zobrist hash settings...)
- [ ] unmakeMove method on position - refactor
- [ ] Refactor zobrist hash generation/update
- [ ] Check a corner case when capturing with promotion a rook or a piece(maybe use a promotion_capture move type?)

#### Evaluation:
- [x] Add piece square tables to evaluation
- [ ] Mobility
- [ ] Tapered Eval

#### Search:
- [x] Add alpha-beta prunning
- [x] Use negamax algorithm
- [x] Transposition table
- [x] Iterative deepening
- [x] Principal Variation
- [ ] Improve move ordering by using transposition table? Access by zorbist key?
- [ ] Aspiration windows
- [ ] Add ~~check~~, stealmate, etc detection while searching
- [ ] Detect threefold repetition - use zobrist hash

### Engine
- [ ] Create an engine struct with main parameters(current best move, depth, etc)
- [ ] Add remaining uci commands
