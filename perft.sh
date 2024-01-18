#!/bin/bash

# For use to debug with (needs rust):
# https://github.com/agausmann/perftree

depth=$1
fen=$2
moves=$3 # Optional


# Launchs aconcagua and sends uci commands via a named pipe(FIFO)
fifo="/tmp/aconcagua_fifo"

mkfifo "$fifo"
./aconcagua < "$fifo" &
pid=$!

# Excecute commands in aconcagua and exit
echo "position fen $fen" > "$fifo"
echo "divide $depth" > "$fifo"
(sleep 10 && echo "quit") > "$fifo"

rm "$fifo"



