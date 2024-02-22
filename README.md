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
- [ ] Try to improve make and unmake move

#### Evaluation:
- [ ] Refactor to use PeSTO evaluation function

#### Search:
- [ ] Aspiration windows
- [ ] Quiescent search
- [ ] Improve move ordering
    - [ ] Killer moves
    - [ ] Improve MVV-LVA
- [ ] Add ~~check~~, stealmate, etc detection while searching
- [ ] Detect threefold repetition - use zobrist hash
- [ ] Refactor Transposition table

### Engine
- [x] Create an engine struct with main parameters(current best move, depth, etc)
- [ ] Improve error handling in uci commands (eg. invalid fen, etc)
- [ ] Add/improve remaining uci commands
