# NC-DHT

## Description
This project simply simulates Chord DHT routing tables and applies Reed-Solomon erasure coding to provide fault tolerance for routing table data. 
It is based on the routing mechanism described in the Chord DHT paper and explores how routing tables can be encoded and reconstructed in case of missing data.

## What This Code Does
* Generates random quorum IDs in a Chord ring.
* Builds Chord finger tables (routing tables) for each quorum using standard Chord logic.
* Encodes a quorum's routing table with Reed-Solomon erasure coding for redundancy.
* Simulates missing shards and successfully reconstructs the routing table.

## How to Run

Clone this repository and run the program:
git clone https://github.com/kriti8671/NC-DHT.git

(If using Go 1.16 or later) Initialize the Go module before running go get: 
go mod init yourprojectname

## Dependencies
github.com/klauspost/reedsolomon (for erasure coding)

## References
[NC-DHT Paper:]( https://ieeexplore.ieee.org/document/10844445)

[Reed-Solomon Go Library:](https://github.com/klauspost/reedsolomon)

[Chord DHT Paper:]( https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf)
