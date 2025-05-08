# NC-DHT
NC-DHT Chord Routing Table Simulation

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

# New Version: Date 2025 April 17
## Design

- main.go: simulation logic
- coding/rs.go: Reed-Solomon encode/decode 
- crypto/: signature logic
- types/: shared structs like Quorum,LookupRequest

## Description
#### 1. Quorum Initialization 
- Simulate a quorum Q1 with n peers.
- Each peer is assigned a private key share using threshold cryptography.
- Threshold t is set, and t+1 signatures are required for legitimacy.

#### 2. Lookup Request + Signature Collection
- A peer constructs a lookup message: [IDp | addp | REQUEST | ts1].
- Peers in the quorum sign the message using schnorr.Sign(...).
- Collect t+1 signatures and simulate quorum signature S1.

#### 3. Signature Verification
- The next quorum verifies S1 using the quorum’s public key.
- This mimics NC-DHT’s inter-quorum proof of legitimacy.
- Routing Table Encoding (with Reed-Solomon)
- The routing table is a map[string]string, JSON-encoded.
- Encoded into dataShards + parityShards using reedsolomon.
- Metadata like DataShards, ParityShards, and OriginalSize is stored.

#### 4. Simulating Shard Loss + Recovery 
- Simulate missing shards by setting them to nil.
- RS library reconstructs missing data using available shards.
- Decoded data is parsed back into a JSON routing table.

# FInal Version
## Date:May 7, 2025

This repository contains the implementation, simulation, and evaluation of NC-DHT, a robust and anonymous Distributed Hash Table (DHT) designed for decentralized systems. Inspired by the research paper "NC-DHT: A Robust and Anonymous DHT for Blockchain Systems" by Tseng et al. (2024), this project focuses on enhancing security and fault tolerance through quorum-based routing, Reed-Solomon erasure coding, and threshold signatures.

The project is implemented in:

Go: For encoding/decoding operations using the klauspost/reedsolomon library.

Python: For threshold cryptography (key generation, signing, combining, and verification) using the gitzhou/threshold-signature-demo repository.

### Key features include:

Simulation of Chord-based routing.
Reed-Solomon encoding/decoding with recovery from missing shards.
Benchmarking of encoding/decoding and threshold operations.
Time complexity analysis comparing observed vs. theoretical models.

The full details, including benchmarking results and time complexity analysis, are documented in the white paper: Os_Project-2.pdf.

### Project Structure

/src/go/: Go source code for Reed-Solomon encoding/decoding and quorum-based routing simulation.
/src/python/: Python source code for threshold cryptography operations.
/docs/: Contains the white paper (Os_Project-2.pdf).
benchmark_plot.png: Encoding vs Decoding Time vs Number of Quorums.
keygen500_performance_boxplot.png: Key Generation, Signing, Combining, and Verification Times vs Number of Peers.

### Requirements

Go: Version 1.18 or higher.
Python: Version 3.8 or higher.
Go: github.com/klauspost/reedsolomon
Python: Clone the gitzhou/threshold-signature-demo repository (see instructions below).

#### Setup and Installation
Clone the Repository:

git clone https://github.com/kriti8671/NC-DHT
cd nc-dht

Set Up Go:
Install the Reed-Solomon library:go get github.com/klauspost/reedsolomon
Navigate to the Go source directory:cd src/go

Set Up Python:
Clone the threshold signature demo repository:git clone https://github.com/gitzhou/threshold-signature-demo.git
cd threshold-signature-demo

#### Running the Code
Run Go Benchmarks (Encoding/Decoding):
From the src/go directory, run the benchmarking script (adjust the filename as per your implementation):go run benchmark.go
This will generate the encoding/decoding performance data (e.g., benchmark_plot.png).

Run Python Benchmarks (Threshold Operations):
From the src/python directory, run the threshold cryptography benchmarking script:python threshold_benchmark.py
This will generate the threshold operation performance data (e.g., keygen500_performance_boxplot.png).


#### Key Findings

##### Reed-Solomon Encoding/Decoding:

Successfully recovers from missing shards but not corrupted shards.

Encoding/decoding times increase non-linearly with the number of quorums (see Figure 1 in the white paper).

##### Threshold Operations:
Key generation complexity: Observed (O(n^{1.35})) vs. theoretical (O(n^2)).

Verification complexity: Observed (O(1)), aligning with theoretical expectations.

See Figure 2 and Table 1 in the white paper for detailed results.

#### Future Work

Integrate threshold signatures using a stable Go library.

Enhance Reed-Solomon to handle corrupted shard recovery.

Simulate network behavior and test NC-DHT under diverse workloads.


## References
[NC-DHT Paper]( https://ieeexplore.ieee.org/document/10844445)

[Reed-Solomon Go Library](https://github.com/klauspost/reedsolomon)

[Threshold-Siganture-Demo](https://github.com/gitzhou/threshold-signature-demo)

[Chord DHT Paper]( https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf)

[Go library](https://pkg.go.dev/go.dedis.ch/kyber/v4)
