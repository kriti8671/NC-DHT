package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"

	// "time"

	"github.com/klauspost/reedsolomon"
)

// Quorum struct representing each node in the DHT
type Quorum struct {
	ID          int
	FingerTable []int
}

func main() {
	// Seed random for consistent randomness
	// rand.Seed(time.Now().UnixNano())

	// Simulation parameters
	m := 4 // bits for ring size (2^m space)
	ringSize := int(math.Pow(2, float64(m)))
	numQuorums := 8

	fmt.Printf("Ring size: %d, Number of Quorums: %d\n", ringSize, numQuorums)

	// Step 1: Create Random Quorums (Even + Odd IDs)
	quorums := createRandomQuorums(numQuorums, ringSize)

	// Step 2: Build Finger Tables
	buildFingerTables(quorums, ringSize, m)

	// Step 3: Print Finger Tables
	fmt.Println("\nFinger Tables for Each Quorum:")
	for _, q := range quorums {
		fmt.Printf("Quorum %d Finger Table: %v\n", q.ID, q.FingerTable)
	}

	// Step 4: Encode routing table of a specific quorum (e.g., first one)
	selectedQuorum := quorums[0]
	fmt.Printf("\nEncoding routing table of Quorum %d...\n", selectedQuorum.ID)

	// Routing entries for encoding
	routingTable := make([]string, len(selectedQuorum.FingerTable))
	for i, destID := range selectedQuorum.FingerTable {
		routingTable[i] = fmt.Sprintf("Q%d -> Q%d", selectedQuorum.ID, destID)
	}

	// Step 5: Reed-Solomon Encode/Decode
	runReedSolomon(routingTable)
}

// Creates random, unique quorum IDs across the ring space
func createRandomQuorums(numQuorums, ringSize int) []Quorum {
	quorumIDMap := make(map[int]bool)
	quorums := make([]Quorum, 0, numQuorums)

	for len(quorums) < numQuorums {
		id := rand.Intn(ringSize)
		if !quorumIDMap[id] {
			quorumIDMap[id] = true
			quorums = append(quorums, Quorum{ID: id})
		}
	}

	// Sort for easier routing table generation
	sort.Slice(quorums, func(i, j int) bool {
		return quorums[i].ID < quorums[j].ID
	})

	fmt.Println("Quorum IDs (random):")
	for _, q := range quorums {
		fmt.Printf("%d ", q.ID)
	}
	fmt.Println()

	return quorums
}

// Builds Chord finger tables for all quorums
func buildFingerTables(quorums []Quorum, ringSize, m int) {
	numQuorums := len(quorums)

	// Extract sorted quorum IDs
	quorumIDs := make([]int, numQuorums)
	for i, q := range quorums {
		quorumIDs[i] = q.ID
	}

	for idx, q := range quorums {
		q.FingerTable = make([]int, m)
		for i := 0; i < m; i++ {
			start := (q.ID + int(math.Pow(2, float64(i)))) % ringSize
			successor := findSuccessor(start, quorumIDs)
			q.FingerTable[i] = successor
		}
		quorums[idx] = q
	}
}

// Finds the successor quorum ID for a given start point
func findSuccessor(start int, quorumIDs []int) int {
	for _, id := range quorumIDs {
		if id >= start {
			return id
		}
	}
	// Wrap around to the smallest quorum ID
	return quorumIDs[0]
}

// Encodes and reconstructs routing table using Reed-Solomon
func runReedSolomon(routingTable []string) {
	dataShards := len(routingTable)
	parityShards := 2
	totalShards := dataShards + parityShards

	shardSize := getMaxEntryLength(routingTable)
	fmt.Printf("Shard size will be %d bytes.\n", shardSize)

	// Prepare data shards
	data := make([][]byte, dataShards)
	for i, entry := range routingTable {
		padded := make([]byte, shardSize)
		copy(padded, []byte(entry))
		data[i] = padded
	}

	// Create encoder
	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		log.Fatalf("Failed to create encoder: %v", err)
	}

	// Prepare all shards (data + parity)
	allShards := make([][]byte, totalShards)
	copy(allShards, data)
	for i := dataShards; i < totalShards; i++ {
		allShards[i] = make([]byte, shardSize)
	}

	// Encode
	err = enc.Encode(allShards)
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	fmt.Println("\nEncoding complete!")
	printShards(allShards)

	// Simulate missing shards (Byzantine behavior)
	fmt.Println("\nSimulating missing shards...")
	allShards[1] = nil // Simulate shard missing
	allShards[4] = nil // Simulate another missing shard

	// Reconstruct missing shards
	err = enc.Reconstruct(allShards)
	if err != nil {
		log.Fatalf("Reconstruction failed: %v", err)
	}

	// Verify data integrity
	ok, err := enc.Verify(allShards)
	if err != nil {
		log.Fatalf("Verification error: %v", err)
	}
	if !ok {
		log.Fatal("Verification failed after reconstruction!")
	}

	fmt.Println("\nReconstruction successful and verified!")
	printShards(allShards)

	// Display recovered routing table
	fmt.Println("\nRecovered Routing Table Entries:")
	for i := 0; i < dataShards; i++ {
		entry := string(allShards[i])
		fmt.Printf("Entry %d: %s\n", i, entry)
	}
}

// Finds the maximum length of routing entries
func getMaxEntryLength(entries []string) int {
	maxLen := 0
	for _, entry := range entries {
		if len(entry) > maxLen {
			maxLen = len(entry)
		}
	}
	return maxLen + 10 // Add some padding
}

// Prints shards for debugging
func printShards(shards [][]byte) {
	for i, shard := range shards {
		if shard == nil {
			fmt.Printf("Shard %d: MISSING\n", i)
		} else {
			fmt.Printf("Shard %d: %s\n", i, string(shard))
		}
	}
}
