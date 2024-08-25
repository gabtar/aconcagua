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
- [ ] Add an specific function to get only captures/checks moves for quiescent
- [ ] Use a move list/array, so as to get rid of 'appends' in moves(it should improve search performance)

#### Evaluation:
- [x] Refactor to use PeSTO evaluation function(tapered eval)

#### Search:
- [x] Aspiration windows
- [x] Principal variation search
- [x] Quiescent search
- [x] Limit quiescent search (needs further work!)
- [x] Add checkmate detection
- [ ] Null move pruning
- [x] Improve move ordering
    - [x] Killer moves
    - [x] History heruistic
    - [x] Improve MVV-LVA
- [ ] Detect threefold repetition - use zobrist hash
- [ ] Refactor Transposition table

### Engine
- [x] Create an engine struct with main parameters(current best move, depth, etc) - Add search state
- [ ] Add Search as method to engine
- [ ] Improve error handling in uci commands (eg. invalid fen, etc)
- [ ] Add/improve remaining uci commands
