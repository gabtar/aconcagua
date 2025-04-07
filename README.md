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

- [ ] Do/clear TODOs/Ideas...

#### Move generation optimizations:
- [x] Try to improve make and unmake move
- [x] Use an uint16/32 for move encoding and an uint16/32 for a 'history board state'
- [x] Use a new MoveList with sort method depending on depth
- [x] Generate the legal moves on a pre-alocated move list
- [x] Make a separate function for ep captures, to just run once instead of running with each pawn move
- [x] Implement a sort whithin the move list depending on move score
- [x] Use a chessMove principal variation...

#### Evaluation:
- [x] Refactor to use PeSTO evaluation function(tapered eval)

#### Search:
- [ ] Adapt to new move format
    - [x] Quiescent search
    - [ ] History heruistic
    - [x] Add checkmate/stealmate detection
- [ ] Add time controls
- [ ] Refactor Transposition table to use w/ move ordering...
- [x] Limit quiescent search -> 5 depth
- [x] Null move pruning
- [x] Improve move ordering
- [ ] Detect threefold repetition - use zobrist hash

### Engine
- [x] Create an engine struct with main parameters(current best move, depth, etc) - Add search state
- [ ] Add Search as method to engine
- [ ] Improve error handling in uci commands (eg. invalid fen, etc)
- [ ] Add/improve remaining uci commands
