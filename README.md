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
- [ ] Refactor zobrist hash generation/update

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
- [ ] Refactor search
- [ ] Improve move ordering
- [ ] Aspiration windows
- [ ] Add ~~check~~, stealmate, etc detection while searching
- [ ] Detect threefold repetition - use zobrist hash

### Engine
- [x] Create an engine struct with main parameters(current best move, depth, etc)
- [ ] Add/improve remaining uci commands
