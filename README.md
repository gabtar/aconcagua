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
- [x] Fix pin detection
- [x] Fix sometimes a black pawn in edges returns illegal moves
- [ ] Threefold repetition.

#### Evaluation:
- [ ] Add piece square tables to evaluation

#### Search:
- [ ] Add alpha-beta prunning
- [ ] Add check, stealmate, etc detection while searching

### Engine
- [ ] Fix - Panic before checkmate
