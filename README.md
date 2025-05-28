# Aconcagua

A Chess engine in go

Setup:
```
git clone https://github.com/gabtar/aconcagua
cd aconcagua
go build .
```

Builds an `aconcagua` executable, a uci compatible engine that can be used with a gui like arena gui or pychess 


## Features

- UCI protocol compatible
- Bitboards representation
- Magic bitboards
- Iterative deepening
- Quiescence search
- Static Exchage Evaluation
- Late move reductions
- Futility pruning
- Aspiration window
- Pieces Square Tables
- Principal Variation Search
- Killer moves
- Transposition table
