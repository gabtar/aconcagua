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
- [ ] Implement a sort whithin the move list depending on move score

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
