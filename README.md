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
- [ ] unmakeMove method on position
- [x] Refactor move generator
- [ ] Refactor zobrist hash generation/update
- [ ] Improve move ordering
- [ ] Check a corner case when capturing with promotion a rook(i think castle rights are not updated...) 

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
- [ ] Aspiration windows
- [ ] Add ~~check~~, stealmate, etc detection while searching

### Engine
- [ ] Create an engine struct with main parameters(current best move, depth, etc)
- [ ] Add remaining uci commands
- [ ] Detect threefold repetition - use zobrist hash
