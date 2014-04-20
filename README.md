Made by Minh Le and Ibukun Adeleye for 15-440 S14 Project 3 

#Description

Tetris Co-op is a Tetris-based game that can be played cooperatively between two to four players. Each player is given a distinct stream of Tetris blocks,which float down from the top of the board. The players can each move their distinct piece anywhere down the board, at varying speeds. The goal is for the players to cooperative clear lines with the pieces given and prevent the Tetris blocks from reaching the top of the board. 

#Implementation

The implementation has three major components: the game clients, the game servers, and the storage servers. The game client and associated GUI will be implemented in Javascript. The game client will communicate with the game server via HTTP requests. The game servers and the storage servers will be written in Go, and communicate with each other via Go RPC. 

One storage server will store all the game state and associated data, but it will be replicated for fault tolerance using a 5-member Multi-PAXOS. A centralized data server will handle the membership for multi-PAXOS. If we have time, we will synchronize global shared resources with Lamport’s Distributed Mutual Exclusion.

