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
git clone https://github.com/yourusername/nc-dht-chord-simulation.git

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
** 1. Quorum Initialization **
- Simulate a quorum Q1 with n peers.
- Each peer is assigned a private key share using threshold cryptography.
- Threshold t is set, and t+1 signatures are required for legitimacy.

** 2. Lookup Request + Signature Collection **
- A peer constructs a lookup message: [IDp | addp | REQUEST | ts1].
- Peers in the quorum sign the message using schnorr.Sign(...).
- Collect t+1 signatures and simulate quorum signature S1.

** 3. Signature Verification **
- The next quorum verifies S1 using the quorum’s public key.
- This mimics NC-DHT’s inter-quorum proof of legitimacy.
- Routing Table Encoding (with Reed-Solomon)
- The routing table is a map[string]string, JSON-encoded.
- Encoded into dataShards + parityShards using reedsolomon.
- Metadata like DataShards, ParityShards, and OriginalSize is stored.

** 4. Simulating Shard Loss + Recovery **
- Simulate missing shards by setting them to nil.
- RS library reconstructs missing data using available shards.
- Decoded data is parsed back into a JSON routing table.



## References
[NC-DHT Paper]( https://ieeexplore.ieee.org/document/10844445)
[Reed-Solomon Go Library](https://github.com/klauspost/reedsolomon)
[Chord DHT Paper]( https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf)
[Go library](https://pkg.go.dev/go.dedis.ch/kyber/v4)
