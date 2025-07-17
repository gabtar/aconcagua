# Aconcagua

A Chess engine in go

Setup:
```
git clone https://github.com/gabtar/aconcagua
cd aconcagua
go build .
```

Builds an `aconcagua` executable, a uci compatible engine that can be used with a gui like arena gui or pychess 
You can also download the precompiled binaries from [here](https://github.com/gabtar/aconcagua/releases)

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

## Variants
- Standard
- Chess 960

## Lichess Bot

Thanks to the people who made the [Lichess bot](https://github.com/lichess-bot-devs/lichess-bot) project, Aconcagua is also available to play on Lichess.

Feel free to challenge AconcaguaBot on Lichess: [AconcaguaBot](https://lichess.org/@/AconcaguaBot)
