# Aconcagua Evaluation Function Tests

This document sumarizes the results of the research attempting to create a new evaluation function for Aconcagua. Previous versions of Aconcauga(<= v3.0.0) used the PeSTO evaluation function (using the PSQT tunned by Ronald Friederich, as tried in his chess engine RofChade) as described [here](https://www.chessprogramming.org/PeSTO%27s_Evaluation_Function)

### Research Metodology
The metodology for evaluating improvements in Aconcagua for the new evaluation function is the following:
1. Define a Base Pieces Square Tables (PSQT) and Pieces Value, mainly derived from [Simplified Evaluation Function](https://www.chessprogramming.org/Simplified_Evaluation_Function) and minor adjustments using common sense
2. Define new features/evaluation attributes for the evaluation function and implement them.
3. Compare each evaluation feature against the Base PSQT only version and Aconcagua-v3.0.0 to look for improvements, as follow:
    3.1 - Test each feature/new version against engine version w/ base PSQT only and Aconcagua-v3.0.0 in a 400 games match with 30s+1s Time Control.
    
    3.2 - For individual/small features use a smaller number of games(300) with 15s+1s Time Control to look for improvements by tuning penaltys/bonuses or the whole feature.

    3.3 - Make Incremental tests by grouping evaluation features to test improvements do not affect the rest of the features.

 I will use [Fastchess](https://github.com/Disservin/fastchess/), an amazing tool to run engine vs engine matches. To run a match against two engine versions with this command line tool, you can run the following command:

```
./fastchessbin -engine cmd=aconcagua-linux-v3.0.0-DEV name=Aconcagua-v3.0.0-DEV \
-engine cmd=aconcagua-linux-v3.0.0 name=Aconcagua-v3.0.0 \
-openings file=./openings/Blitz_Testing_4moves.epd format=epd \
-each tc=30+1 -rounds 200 -recover -concurrency 4
```

You need to provide the two engines executables in the same folder. Also it's recomended to use an opening book for the match to set up starting positions and have different setups. At the end of the match you will get the results with the estimated elo difference with the error margin(the more games you play, the lower the error), and some other useful metrics about the match.

### Time controls
I choosed bullet time controls to do the research, because it generally works well and mainly, it saves a lot of time. Generally speaking if your engine reaches up to depth 8-10 from the starting position within less than 1 second using this time controls works well to do rapid tests and obtain an objetive conclusion


### What to look for?
Mainly improvements in elo score against the base version. Fastchess also outputs some useful metrics about the engines comparison. One of them is LOS, that is also taken into account here. LOS means likehood of superiority, its a metric that shows the percentage of the probability the engine is better compared to the other version. The higher this metric, the higher chance the engine will improve

I consider the following: the engine is better when the elo improvment is higher than the previous version(elo > 0) and the LOS is > 50%. This is because there are some cases where the elo improvement is not enought to define superiority, because the error margin may be higher than the elo gain and cannot conclude if the new evaluation feature improves the engine 

So, by using simple trial and error metodology, i adjust the penalties/bonuses for the various positional evaluations and check for the new results to see if the engine is better. There are some cases in which a i have done a wrong implementation of a feature or misundertood it at all, so having this kind of tests and metrics help go in the rigth way.

The process is be long, because making matches engine vs engine over and over again, its too much time consuming, but by using this metodology i can be sure to track the improvement. 

There is a point in building a chess engine where you start to add features and see the improvement very easily(especially in low elo), but when you start reaching higher elo/strength you must have a way to track improvement and i belive this is an appropiate one, and its easy to implement, but takes time.

### Evaluation test

#### PSQT only evaluation

| Feature    | Elo vs NewPSQT (LOS %) | Elo vs Aconcagua-v3.0.0 (LOS %) | Observations                                                                      |
| ---------- | ---------------------- | ------------------------------- | --------------------------------------------------------------------------------- |
| New PSQT   |        0.0 (0.00%)     |     -213.60±36.24 (0.00 %)      | New PSQT derived from simplified evaluation function and concepts from CPW-engine |

> As expected, PeSTO evaluation function with RofChade PSQT tunned are much better than basic PSQT with some vague improvements.
> I don't tried adjusting different PSQT, because i plan to use later automated tuning to find the best PSQT.

#### Mobility

| Feature        | Elo vs NewPSQT (LOS %) | Elo vs Aconcagua-v3.0.0 (LOS %) | Observations                                                                                          |
| -------------- | ---------------------- | ------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Mobility-v1    | -39 elo aprox. aborted |     ----------------------      | Fixed penalty/bonus for each square for each piece. Test aborted at mid games, due to bad results     |
| Mobility-v2    | 54.29±30.33 (99.98 %)  |    -134.95±34.92 (0.00 %)       | Incremental bonus per square attacked, similar to CPW engine approach. Pseudo legal moves.            |

> Maybe i should have tried different approach for mobility, like safe mobility or something like that, but i think the improvements here are good enough and dont compromise much the overall speed during the search.

#### Pawn structure

| Feature               | Elo vs NewPSQT (LOS %) | Elo vs Aconcagua-v3.0.0 (LOS %) | Observations                                                                                 |
| --------------------- | ---------------------- | ------------------------------- | -------------------------------------------------------------------------------------------- |
| Doubled Pawns 5/15    | -3.47±27.61 (40.24 %)  |     ------------------------    | Penalties mg -5 / eg -15. Aplied to each pawn individually. Double counts penalties          |
| Doubled Pawns 20/25   | -24.36±37.21 (9.81 %)  |     ------------------------    | Only penalties changed                                                                       |
| Doubled Pawns 4/6     | 9.27±34.83  (69.97 %)* |     ------------------------    | Can't conclude they are better, but is the best result found for doubled pawns               |
| Isolated Pawns 5/15   | 5.79±30.27 (64.65%)*   |     ------------------------    | Checked pawn by pawn                                                                         |
| Backward Pawns 5/25   | -53.70±35.5  (0.13 %)  |     ------------------------    | First attempt. I believe its a bad implementaion of backward pawns, although tests passes    |
| Backward Pawns v2 5/25| -88.74±37.92 (0.00 %)  |     ------------------------    | Using the chess programing wiki routine. Checked pawn by pawn. Seems bug with black's pawns  |
| Backward Pawns v2 5/5 | 1.74±28.03 (54.84%)    |     ------------------------    | Calculate all backward pawns at once. Fix bug. Adjust penalties                              |
| Backward Pawns v3 6/12| -13.90±34.82 (21.59%)  |     ------------------------    | Use precalculated attacksFrontSpans array. Adjust penalties again                            |
| Backward Pawnsv3 3/8  | 32.52±31.38 (97.98 %)* |     ------------------------    | Penalties updated(works much better with lower penalties)                                    |
| Passed Pawns 20/40    | -19.71±31.90 (11,18 %) |     ------------------------    | Fixed bonus for passed pawns                                                                 |
| Passed Pawns incr     | 31.35±33.59 (96.75%)   |     ------------------------    | Incremental bonus on ranks to go to promotion (2-7) 10, 20, 30, 40, 50                       |
| Passed Pawns incr2    | 37.20±33.95 (98.51%)*  |     ------------------------    | Huge bonus near promotion rank (2-7) 10, 20, 30, 60, 100                                     |
| All Pawn Improvements*| 52.51±28.47 (99.99 %)  |      -149.26±31.46 (0.00 %)     | Nice improvement. Penalties/Bonus are best results in individual tests(400games 30s+1s TC)   |
| Doubled Pawnsv2 8/12  | 11.59±33.13 (75.42%)*  |     ------------------------    | Same penalty(not double counted). Calculate all doubled pawns at once.                       |
| Doubled Pawnsv2 8/4   | -2.32±33.15 (44.54%)   |     ------------------------    | Try different/lower penalty.                                                                 |
| Doubled Pawnsv2 12/15 |  3.47±33.38 (58.11%)   |     ------------------------    | Try higher penalty.                                                                          |
| Doubled Pawnsv2 6/10  |  2.32±30.87 (55.86%)   |     ------------------------    | Last attempt w/ doubled pawns.                                                               |
| Isolated Pawnsv2 10/15| -1.16±34.91 (47.40%)   |     ------------------------    | Calculate all isolated pawns at once. Updated penalties                                      |
| Isolated Pawnsv2 5/20 | -12.75±32.4 (21.96%)   |     ------------------------    | Trying different penalties                                                                   |
| Isolated Pawnsv2 8/15 | -8.69±29.75 (28.30%)   |     ------------------------    | Last attempt for isolated pawns.                                                             |
| All Pawn Structure v2 |  50.74±31.45 (99.94%)  |      -161.92±33.13 (0.00 %)     | Calculate all pawn penalties at once in a function. Used best penalties/bonus found          |

> I got a better result in intermediate analisys, but i was calculating all penalties separately, and some of them when were evaluating each pawn at one a time. So i will use the last implementation besides is not the best result.

##### Checkpoint 1: Mobility + Pawn structure evaluation

| Feature               | Elo vs NewPSQT (LOS %) | Elo vs Aconcagua-v3.0.0 (LOS %) | Observations                                                                                 |
| --------------------- | ---------------------- | ------------------------------- | -------------------------------------------------------------------------------------------- |
| Mobility + Pawns      | 81.37±31.63(100.00 %)  |      -80.45±33.30(0.00 %)       |  Great improvement. Still behind PeSTO evaluation, but not bad.                              |

> Looks like that both results combined improves more than the sum of the individua, it gives more positional knowledge to the engine and makes it more accurate. 
> Anyway a good tunned PSQT works better in this version of Aconcagua. Overal result against Aconcagua v3 - Games: 400, Wins: 123, Losses: 214, Draws: 63.
> It's still behind, but i think next improvements to the evaluation function will catch the PeSTO evaluation function performance and maybe beyond.

#### King Safety

| Feature               | Elo vs NewPSQT (LOS %) | Elo vs Aconcagua-v3.0.0 (LOS %) | Observations                                                                                     |
| --------------------- | ---------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------ |
| Pawn Shield           | -12.75±32.09(21.73 %)  |   ---------------------------   | Bonus for pawns in front of the king 5 * (7 - rank distance), -20 for no pawn.                   |
| Pawn Shield v2        |  6.95±33.92 (65.65%)*  |    ---------------------------  | Bonus for pawns in front of the king 20 1 rank, 10 2 rank, 0 more than 2 ranks, -20 for no pawn. |
| Pawn Storm            | -58.45±35.80 (0.05 %)  |    ---------------------------  | Penalty -20/rankdiff for each pawn in adjacent files to the king. If rankdiff > 4 no penalties.  |
| Pawn Storm v2         | -45.42±34.59 (0.45%)   |    ---------------------------  | Same, but if all pawns are blocked, give no penalty. Cannot open files so king should be safe.   |
| Pawn Storm v3         | -22.03±33.02 (9.42%)   |    ---------------------------  | v2 with penalty -15/rankdiff and only if rankdiff > 3. Only during middlegame.                   |
| King zone Attacks v1  |  26.69±32.18(94.92%)   |   ----------------------------  | Weighted attackers on king mobility zone(10*Q, 5*R, 3*B, 3*N, 1*P)                               |  
| King zone Attacks v2  |  9.27±33.30 (70.80%)   |   ----------------------------  | Forget to add the king itself to the king zone. The king + 1 square around it. Rest same as v1   |
| King zone Attacks v1  |        TODO            |   ----------------------------  | Only modify weights (15*Q, 8*R, 3*B, 3*N, 1*P). King not taking into account                     |


### TODO:

3. King safety
   3.1 Pawn shield
   3.2 Pawn storm
   3.3 Open file
   3.4 King attackers
   3.5 King Endgame Mobility/Centralization
4. Center control
5. Open files
6. Bishop pair bonus
7. Tempo?


