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

## Strength

Some of the Aconcagua releases have been tested by the CCRL Team (thank you). Here is a little summary of the playing strength of the engine:

<div align="center">


| Version         | CCRL 40/15 Rating            | Release Date                |
|-----------------|------------------------------|-----------------------------|
|Aconcagua v4.1.0 |  [2440](https://computerchess.org.uk/4040/cgi/engine_details.cgi?print=Details&each_game=0&eng=Aconcagua%204.1.0%2064-bit#Aconcagua_4_1_0_64-bit)       										 | [Dec 14, 2025](https://github.com/gabtar/aconcagua/releases/tag/v4.1.0)                            |
|Aconcagua v5.0.0 |  [2616](https://computerchess.org.uk/4040/cgi/engine_details.cgi?print=Details&each_game=0&eng=Aconcagua%205.0.0%2064-bit#Aconcagua_5_0_0_64-bit)       										 | [Jan 25, 2026](https://github.com/gabtar/aconcagua/releases/tag/v5.0.0)                            |
|Aconcagua v5.1.0 |  [2667](https://computerchess.org.uk/4040/cgi/engine_details.cgi?print=Details&each_game=0&eng=Aconcagua%205.1.0%2064-bit#Aconcagua_5_1_0_64-bit)       										 | [Mar 1, 2026](https://github.com/gabtar/aconcagua/releases/tag/v5.1.0)                            |
|Aconcagua v5.2.0 |  2800 (estimated)       										 | [May 31, 2026](https://github.com/gabtar/aconcagua/releases/tag/v5.2.0)                            |


</div>

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
- Check Extension
- Draw detection(by repetition/insufficient material/50 moves rule)
- Null move pruning
- Reverse Futitly Pruning
- Futility pruning
- Internal Iterative Deepening
- Late move reductions
- Late move pruning
- Static Exchage Evaluation pruning

#### Time management
- Soft/Hard time limits
- Search time adjustment based on (estimated) moves to go, score stability and node fraction searched at root

#### Move Ordering
- Hash move (from transposition table)
- Good Captures
- Killer moves
- Counter move
- Non Captures moves ordered by History Heuristic
- Bad Captures (Static Exchange Evaluation < 0)

#### Evaluation
- Hand Craft Tuned Evaluation, using Lichess Big3 Resolved dataset
- Pieces Square Tables
- Tappered Evaluation
- King Safety(Shiled, Storm, OpenFiles and King zone attacks)
- Mobility
- Isolated, Doubled, Passed and Backward Pawns
- Knight/Bishops Outpost
- Threats(Mayor pieces threated by pawns, minors by pawns, safe checks threats)
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
* [CCRL](https://computerchess.org.uk/) - The most realible engines rating list. They do an unvaluable work for all engine developers
* Some of the open source chess engines that Aconcagua has been inspired by:
    * [Ethereal](https://github.com/etherealengine/ethereal)
    * [Blunder](https://github.com/etherealengine/blunder)
    * [Zurichess](https://bitbucket.org/zurichess/zurichess/src)
    * [GoBit](https://github.com/carokanns/GoBit)
    * [Vice](https://github.com/bluefeversoft/vice)
    * [WukongJS](https://github.com/maksimKorzh/wukongJS)
    * [TSCP](https://sites.google.com/site/tscpchess/home)
