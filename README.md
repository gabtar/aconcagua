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
- [ ] Fix pawn move(capture i think) when on an absolute pin(king - pawn)
- [ ] Refactor move encode
- [ ] Add undo move in the position
- [ ] Refactor move generator
- [ ] Improve move ordering

#### Evaluation:
- [x] Add piece square tables to evaluation
- [ ] Mobility
- [ ] Trapped pieces

#### Search:
- [x] Add alpha-beta prunning
- [ ] Use negamax algorithm
- [ ] Transposition table
- [ ] Iterative deepening
- [ ] Aspiration windows
- [ ] Add ~~check~~, stealmate, etc detection while searching

### Engine
- [ ] Create an engine struct with main parameters(current best move, depth, etc)
- [ ] Add remaining uci commands
- [ ] Detect threefold repetition
