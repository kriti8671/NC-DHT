// main.go
package main

import (
	"encoding/json"
	"fmt"
	"ncdht/coding"
	"ncdht/crypto"
	"ncdht/types"
	"time"
)

func main() {
	// Step 1: Build quorum Q1
	size := 5
	threshold := 3
	quorum, err := types.BuildQuorum(1, size, threshold)
	if err != nil {
		panic(err)
	}
	fmt.Println("Quorum Q1 created with", size, "peers")

	// Step 2: Create Lookup Request
	req := types.LookupRequest{
		ID:        "peerP",
		Address:   "10.0.0.1",
		Request:   "LOOKUP",
		Timestamp: time.Now().Unix(),
	}
	message := []byte(fmt.Sprintf("%s|%s|%s|%d", req.ID, req.Address, req.Request, req.Timestamp))

	// Step 3: Collect signature shares from peers
	var shares [][]byte
	for i, peer := range quorum.Peers {
		sig, err := crypto.SignMessage(message, peer, i)
		if err != nil {
			fmt.Println("Error signing:", err)
			continue
		}
		shares = append(shares, sig)
		fmt.Printf(" Signature share from Peer %d\n", peer.ID)
		if len(shares) >= threshold+1 {
			break
		}
	}

	// Step 4: Combine signature (simulated)
	S1, err := crypto.CombineSignatures(shares)
	if err != nil {
		panic(err)
	}
	fmt.Println("Quorum signature S1 created.")

	// Step 5: Verify S1 using quorum public key
	err = crypto.VerifySignature(message, S1, quorum.PubKey)
	if err != nil {
		fmt.Println("Signature verification failed:", err)
	} else {
		fmt.Println("Signature verified by next quorum.")
	}

	// Step 6: Simulate routing table encoding (as key-value map)
	routingTable := map[string]string{
		"nextQ2Key": "0xA5B3",
		"nextQ3Key": "0xBEEF",
	}

	// Configurable parameters in one place
	dataShards := 2
	parityShards := 3

	encoded, err := coding.EncodeRoutingTableWithParams(routingTable, dataShards, parityShards)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nRouting table encoded into %d shards (data=%d, parity=%d):\n",
		len(encoded.Shards), encoded.Metadata.DataShards, encoded.Metadata.ParityShards)
	for i, shard := range encoded.Shards {
		fmt.Printf("Shard %d: %x\n", i, shard)
	}

	// Step 7: Simulate missing shards
	received := make([][]byte, len(encoded.Shards))
	for i := range encoded.Shards {
		if i == 2 || i == 4 {
			received[i] = nil // Simulate loss of 2 shards
		} else {
			received[i] = encoded.Shards[i]
		}
	}
	encoded.Shards = received

	decodedTable, err := coding.DecodeRoutingTable(encoded)
	if err != nil {
		fmt.Println("\nFailed to decode:", err)
		return
	}

	decodedJSON, _ := json.MarshalIndent(decodedTable, "", "  ")
	fmt.Println("\nReconstructed routing table:")
	fmt.Println(string(decodedJSON))
}
