# Aconcagua

Chess engine in go - WIP

# Main parts of a chess engine

- [ ] Move generation/validation (returns all legal moves from a position)
- [ ] Evaluation function (scores out a position saying which side is better)
- [ ] Search algoritm (finds the best next move among all posibles moves based on the evaluation function)


### TODO, Move generation:
- [x] From a given position return all posibles next moves/positions for a given side(will be used later for searching best move)
    - [x] Tests if given a position returns all availables moves correctly
- [x] Function to update the position(struct) when passing a legal move 
- [x] Detect on a given position
    - [x] Checkmate
    - [x] Stealmate
    - [x] Insuficient material
    - [x] 50 move rule (It's not automatically draw).
- [ ] Threefold repetition - Insuficient information on the position struct. I need to store the previous positions fens to check if there is a repetition.
