<div align="center">

# 🏔️ Aconcagua Chess Engine

</div>

A UCI-compatible chess engine written in Go.

## Installation

### 1. Build from source

#### Requirements
* Go 1.26.0 or later

```
git clone https://github.com/gabtar/aconcagua
cd aconcagua
go build -o aconcagua ./cmd/aconcagua
```

This will build an executable named `aconcagua` in the current directory.

> [!NOTE]  
> Depending on your operating system (especially on some Ubuntu/Debian based distributions) you might not have the required Go version to build Aconcagua. If that's your case, you might want to look at the steps posted by [tissatussa](https://github.com/tissatussa) in this [issue #9](https://github.com/gabtar/aconcagua/issues/9).

### 2. Download precompiled binaries

* Go to the [Releases page](https://github.com/gabtar/aconcagua/releases)
* Download the binary for your platform(linux, windows or macos)

## Usage 

Aconcagua is an UCI-compatible chess engine and it works as it with most popular chess GUI out of the box:

* [Arena Chess GUI](http://www.playwitharena.de/)
* [PyChess](https://github.com/pychess/pychess)
* [Scid vs PC](https://scidvspc.sourceforge.net/)
* Or any other GUI that supports the UCI protocol

Simply add the engine executable to your GUI and set it as an UCI protocol compatible engine to start playing.

## Features

- UCI protocol compatible
- Chess 960 / Fischer Random Chess suport
- Bitboards representation
- Magic bitboards for attacks/move generation

#### Search
- Iterative Deepening
- Aspiration window
- Principal Variation Search
- Quiescence search
- Static Exchage Evaluation
- Transposition table w/ buckets system
- Mate Distance Pruning
- Draw detection(by repetition/insufficient material/50 moves rule)
- Null move pruning
- Reverse Futitly Pruning
- Futility pruning
- Internal Iterative Deepening
- Late move reductions
- Late move pruning

#### Move Ordering
- Hash move (from transposition table)
- Good Captures
- Killer moves
- Counter move
- Non Captures moves ordered by History Heuristic
- Bad Captures (Static Exchange Evaluation < 0)

#### Evaluation
- Hand Craft Tuned Evaluation
- Pieces Square Tables
- Tappered Evaluation
- King Safety
- Mobility
- Isolated, Doubled, Passed and Backward Pawns
- Knight/Bishops Outpost
- Bishop Pairs
- Rooks on semi open/open files
- Tempo

## Lichess Bot

Thanks to the amazing [Lichess bot project](https://github.com/lichess-bot-devs/lichess-bot), Aconcagua is also available to play on Lichess.

Feel free to challenge AconcaguaBot on Lichess: [AconcaguaBot](https://lichess.org/@/AconcaguaBot)

## Acknowledgments

* [Chess Programing wiki](https://www.chessprogramming.org) - A must resource for anyone who wants to build/learn about chess engines
* [Lichess](https://lichess.org) - The best platform to play online chess
* [Lichess bot project](https://github.com/lichess-bot-devs/lichess-bot) - A bridge between Lichess bots and chess engines
* Some of the open source chess engines that Aconcagua has been inspired by:
    * [Ethereal](https://github.com/etherealengine/ethereal)
    * [Blunder](https://github.com/etherealengine/blunder)
    * [Zurichess](https://bitbucket.org/zurichess/zurichess/src)
    * [GoBit](https://github.com/carokanns/GoBit)
    * [Vice](https://github.com/bluefeversoft/vice)
    * [WukongJS](https://github.com/maksimKorzh/wukongJS)
    * [TSCP](https://sites.google.com/site/tscpchess/home)
