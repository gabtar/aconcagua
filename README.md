# Aconcagua

Chess engine in go - WIP

Setup:
```
git clone https://github.com/gabtar/aconcagua
cd aconcagua
go build .
```

Basic usage from command line(run ./aconcagua -h for help):
```
./aconcagua -fen="r5rk/5p1p/5R2/4B3/8/8/7P/7K w - - 0 1" -depth=3
```

It outputs the best move sequence found. NOTE: Not seeing the checkmate now, but it returns the second best line with any other engine.
```
Score:  330
0: f6 -> f7, g8 -> g7, 1: e5 -> g7,
```

### TODO:

#### Move generation:
- [ ] Threefold repetition - Insuficient information on the position struct. I need to store the previous positions fens to check if there is a repetition.

#### Evaluation:
- [x] Basic evaluation by comparing material in centipawn value
- [ ] Add piece square tables to evaluation

#### Search:
- [x] Minmax with best move tracking
- [ ] Add alpha-beta prunning
- [ ] Add check, checkmate, stealmate, etc detection while searching
